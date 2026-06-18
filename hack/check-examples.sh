#!/usr/bin/env bash
# Generated example install, build, and typecheck checks.
#
# Go builds and vets, TypeScript installs and typechecks against its real
# dependencies, and Python installs its dependencies and byte-compiles its
# sources. Byte-compile checks Python syntax only, not that the example uses the
# SDK correctly. It does not run workers or reach a Hatchet engine; runtime
# behavior is covered by hack/e2e-examples.sh.
#
# Run from the repository root. Requires go, node with pnpm, and python with
# poetry on PATH.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

# Remove gitignored build artifacts these checks create under the examples, so a
# later `go run ./cmd/generate-examples --check` is not tripped by them.
clean_example_artifacts() {
  rm -rf examples/simple-python/.venv \
         examples/simple-typescript/node_modules \
         examples/simple-typescript/dist
  find examples/simple-python -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
}
trap clean_example_artifacts EXIT

echo "== Go example: build and vet =="
(
  cd examples/simple-go
  go build ./...
  go vet ./...
)

echo "== TypeScript example: install and typecheck =="
(
  cd examples/simple-typescript
  pnpm install --frozen-lockfile
  pnpm exec tsc --noEmit
)

echo "== Python example: install and byte-compile =="
(
  cd examples/simple-python
  poetry install --no-interaction
  poetry run python -m compileall -q src
)

echo "All example checks passed."
