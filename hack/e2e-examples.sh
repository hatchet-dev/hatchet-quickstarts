#!/usr/bin/env bash
# Runtime end-to-end test for the generated examples.
#
# It starts one local Hatchet engine, then for each selected language starts
# each generated worker (simple and scheduled), runs its trigger, and asserts
# the trigger exits zero with the expected output. Workers and containers are removed on exit.
#
# Readiness comes from the trigger itself: the worker is started, then the
# trigger is retried within a bounded window until it succeeds or the worker
# process exits. This avoids "hatchet worker list", which needs a TTY, and
# avoids guessing language-specific readiness log lines.
#
# Run from the repository root. Requires hatchet and docker on PATH, plus the
# toolchain each selected example needs (go, poetry, or node with pnpm).
#
# Environment:
#   HATCHET_SERVER_TAG      hatchet-lite image tag (default "latest")
#   HATCHET_E2E_LANGUAGES   space-separated subset of "go python typescript"
#                           (default all three)
set -euo pipefail

SERVER_TAG="${HATCHET_SERVER_TAG:-latest}"
LANGUAGES="${HATCHET_E2E_LANGUAGES:-go python typescript}"

# Bounded timing. A single trigger attempt is capped, and the per-language wait
# is capped, so a stuck worker or engine cannot hang the run.
TRIGGER_TIMEOUT="${HATCHET_E2E_TRIGGER_TIMEOUT:-180}"
MAX_WAIT="${HATCHET_E2E_MAX_WAIT:-420}"
RETRY_INTERVAL="${HATCHET_E2E_RETRY_INTERVAL:-5}"

# The scheduled templates ship a */5 cron, so we size the wait to one
# five-minute interval plus slack for the run to complete.
CRON_MAX_WAIT="${HATCHET_E2E_CRON_MAX_WAIT:-390}"
CRON_POLL_INTERVAL="${HATCHET_E2E_CRON_POLL_INTERVAL:-10}"

COMPOSE_PROJECT="hatchet-cli"
COMPOSE_NETWORK="hatchet-cli_default"

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

LOGDIR="$(mktemp -d)"
WORKER_PIDS=()

# kill_tree terminates a specific process and its descendants. It is scoped to
# one pid, never a broad pattern match.
kill_tree() {
  local pid="$1" child
  for child in $(pgrep -P "$pid" 2>/dev/null || true); do
    kill_tree "$child"
  done
  kill -TERM "$pid" 2>/dev/null || true
}

# clean_example_artifacts removes gitignored build artifacts the workers create
# under the examples (poetry .venv, __pycache__, pnpm node_modules), so a later
# `go run ./cmd/generate-examples --check` is not tripped by them.
clean_example_artifacts() {
  rm -rf examples/simple-python/.venv \
         examples/simple-typescript/node_modules \
         examples/simple-typescript/dist \
         examples/use-cases/scheduled/python/.venv \
         examples/use-cases/scheduled/typescript/node_modules \
         examples/use-cases/scheduled/typescript/dist
  find examples/simple-python examples/use-cases/scheduled/python \
    -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
}

cleanup() {
  if [ "${#WORKER_PIDS[@]}" -gt 0 ]; then
    for pid in "${WORKER_PIDS[@]}"; do
      kill_tree "$pid"
    done
  fi
  docker ps -aq --filter "label=com.docker.compose.project=${COMPOSE_PROJECT}" \
    | xargs -r docker rm -f >/dev/null 2>&1 || true
  docker network rm "${COMPOSE_NETWORK}" >/dev/null 2>&1 || true
  clean_example_artifacts
  rm -rf "$LOGDIR"
}
trap cleanup EXIT

require() {
  command -v "$1" >/dev/null 2>&1 || { echo "ERROR: '$1' is required on PATH" >&2; exit 2; }
}

require hatchet
require docker
require curl
require python3
docker info >/dev/null 2>&1 || { echo "ERROR: docker daemon is not available" >&2; exit 2; }

# The profiles.yaml schema comes from pkg/config/cli/config.go in the
# main hatchet repo. The parse is simple and uses only stdlib python to
# avoid adding a PyYAML dependency
read_profile() {
  python3 - "$HOME/.hatchet/profiles.yaml" "$1" <<'PY'
import re
import sys

path, want = sys.argv[1], sys.argv[2]
profile = None
out = {}
for line in open(path):
    m = re.match(r"^(\s*)([^\s:]+):\s*(.*)$", line.rstrip("\n"))
    if not m:
        continue
    indent, key, value = len(m.group(1)), m.group(2), m.group(3)
    if indent == 4 and value == "":
        profile = key
    elif indent == 8 and profile == want:
        out[key] = value
print("\t".join([out.get("token", ""), out.get("tenantid", ""), out.get("apiserverurl", "")]))
PY
}

# The engine adds the hatchet__cron_expression metadata key to every run
# a cron creates, so counting completed runs after `since` that have the
# key proves a cron run completed. That run could belong to a cron from
# an earlier scheduled example in this tenant, so we also check this
# worker's log for a new "scheduled task ran" line.
assert_cron_run() {
  local lang="$1" logf="$2"

  local token tenant api_url
  IFS=$'\t' read -r token tenant api_url < <(read_profile local)
  if [ -z "$token" ] || [ -z "$tenant" ] || [ -z "$api_url" ]; then
    echo "ERROR: could not read the local profile credentials" >&2
    return 1
  fi

  # The token goes into a header file rather than onto a command line,
  # where ps could show it.
  local authf="$LOGDIR/auth-header"
  (
    umask 077
    printf 'Authorization: Bearer %s\n' "$token" > "$authf"
  )

  local since
  since=$(date -u +%Y-%m-%dT%H:%M:%S.000Z)
  local runs_before
  runs_before=$(grep -c "scheduled task ran" "$logf" 2>/dev/null || true)

  echo "== ${lang}: waiting for a cron run (max ${CRON_MAX_WAIT}s) =="

  local deadline=$(( $(date +%s) + CRON_MAX_WAIT ))
  while [ "$(date +%s)" -lt "$deadline" ]; do
    if grep -qiE "task run failed|TypeError|Traceback|panic:" "$logf"; then
      echo "ERROR: ${lang} worker log reports a task failure" >&2
      grep -iE -A6 "task run failed|TypeError|Traceback|panic:" "$logf" | head -20 >&2
      return 1
    fi

    local completed runs_now
    # Each request is time bound to prevent a hung API socket from
    # outliving the poll interval.
    completed=$(curl -sf --connect-timeout 5 --max-time 5 -H @"$authf" \
      "$api_url/api/v1/stable/tenants/$tenant/workflow-runs?since=$since&statuses=COMPLETED&only_tasks=false&limit=5" \
      | python3 -c 'import sys, json; rows = json.load(sys.stdin).get("rows") or []; print(sum(1 for r in rows if "hatchet__cron_expression" in (r.get("additionalMetadata") or {})))' \
      2>/dev/null || echo 0)
    runs_now=$(grep -c "scheduled task ran" "$logf" 2>/dev/null || true)

    if [ "$completed" -gt 0 ] && [ "$runs_now" -gt "$runs_before" ]; then
      echo "== ${lang}: a cron run completed =="
      return 0
    fi
    sleep "$CRON_POLL_INTERVAL"
  done

  echo "ERROR: no cron run completed for ${lang} within ${CRON_MAX_WAIT}s" >&2
  echo "--- last 40 lines of ${lang} worker log ---" >&2
  tail -40 "$logf" >&2 || true
  return 1
}

run_language() {
  local lang="$1" dir="$2" trigger="$3"; shift 3
  local expected=("$@")
  local logf="$LOGDIR/${lang}-worker.log"
  local trigf="$LOGDIR/${lang}-trigger.log"

  echo "== ${lang}: starting worker =="
  ( cd "$dir" && exec hatchet worker dev --profile local --no-reload ) >"$logf" 2>&1 &
  local wpid=$!
  WORKER_PIDS+=("$wpid")

  echo "== ${lang}: triggering until ready (max ${MAX_WAIT}s) =="
  local deadline=$(( $(date +%s) + MAX_WAIT ))
  local ok=0
  while [ "$(date +%s)" -lt "$deadline" ]; do
    if ! kill -0 "$wpid" 2>/dev/null; then
      echo "ERROR: ${lang} worker exited before the trigger succeeded" >&2
      echo "--- last 40 lines of ${lang} worker log ---" >&2
      tail -40 "$logf" >&2 || true
      return 1
    fi
    if ( cd "$dir" && timeout "$TRIGGER_TIMEOUT" hatchet trigger "$trigger" --profile local ) >"$trigf" 2>&1; then
      ok=1
      break
    fi
    sleep "$RETRY_INTERVAL"
  done

  if [ "$ok" -ne 1 ]; then
    echo "ERROR: ${lang} trigger did not succeed within ${MAX_WAIT}s" >&2
    echo "--- last 40 lines of ${lang} trigger log ---" >&2
    tail -40 "$trigf" >&2 || true
    return 1
  fi

  local missing=0 exp
  for exp in "${expected[@]}"; do
    if ! grep -qiF "$exp" "$trigf"; then
      echo "ERROR: ${lang} trigger did not print expected output: ${exp}" >&2
      missing=1
    fi
  done
  if [ "$missing" -ne 0 ]; then
    echo "--- ${lang} trigger output ---" >&2
    cat "$trigf" >&2 || true
    return 1
  fi

  # We wait for a cron run only in TypeScript. Its SDK passes null cron
  # input through to the task, so a null-input bug only shows when cron
  # fires; the manual trigger always supplies input. Python and Go fill
  # a missing input with defaults and cannot fail this way.
  if [ "$lang" = "typescript-scheduled" ]; then
    if ! assert_cron_run "$lang" "$logf"; then
      return 1
    fi
  fi

  echo "== ${lang}: PASS =="
  kill_tree "$wpid"
  return 0
}

echo "Starting local Hatchet runtime (tag: ${SERVER_TAG})"
hatchet server start --profile local --tag "$SERVER_TAG"

for lang in $LANGUAGES; do
  case "$lang" in
    go)
      run_language go examples/simple-go simple "hello, world!"
      run_language go-scheduled examples/use-cases/scheduled/go manual-run "hello from a manual run"
      ;;
    python)
      run_language python examples/simple-python simple "42"
      run_language python-scheduled examples/use-cases/scheduled/python manual-run "hello from a manual run"
      ;;
    typescript)
      run_language typescript examples/simple-typescript simple "Hello, world!" "Hello, moon!"
      run_language typescript-scheduled examples/use-cases/scheduled/typescript manual-run "hello from a manual run"
      ;;
    *) echo "ERROR: unknown language '${lang}'" >&2; exit 2 ;;
  esac
done

echo "All requested example e2e checks passed: ${LANGUAGES}"
