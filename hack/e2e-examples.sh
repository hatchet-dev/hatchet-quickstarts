#!/usr/bin/env bash
# Runtime end-to-end test for the generated examples.
#
# It starts one local Hatchet engine, then for each selected language starts the
# generated worker, triggers the workflow, and asserts the trigger exits zero
# with the expected output. Workers and containers are removed on exit.
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
         examples/simple-typescript/dist
  find examples/simple-python -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
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
docker info >/dev/null 2>&1 || { echo "ERROR: docker daemon is not available" >&2; exit 2; }

run_language() {
  local lang="$1" dir="$2"; shift 2
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
    if ( cd "$dir" && timeout "$TRIGGER_TIMEOUT" hatchet trigger simple --profile local ) >"$trigf" 2>&1; then
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

  echo "== ${lang}: PASS =="
  kill_tree "$wpid"
  return 0
}

echo "Starting local Hatchet runtime (tag: ${SERVER_TAG})"
hatchet server start --profile local --tag "$SERVER_TAG"

for lang in $LANGUAGES; do
  case "$lang" in
    go)         run_language go         examples/simple-go         "hello, world!" ;;
    python)     run_language python     examples/simple-python     "42" ;;
    typescript) run_language typescript examples/simple-typescript "Hello, world!" "Hello, moon!" ;;
    *) echo "ERROR: unknown language '${lang}'" >&2; exit 2 ;;
  esac
done

echo "All requested example e2e checks passed: ${LANGUAGES}"
