#!/usr/bin/env bash
# Pre-commit verification. Exits 2 to block the commit on failure.
# Stack presets append build/test verification to the STACK CHECKS section below.
set -euo pipefail

# --- Common: first line of defense against leaking secrets ---
staged=$(git diff --cached --name-only)
if printf '%s\n' "$staged" | grep -qE '(^|/)\.env($|\.)'; then
  echo "Blocked: a .env-type file is staged. Commit is not allowed." >&2
  exit 2
fi

# --- STACK CHECKS (presets append here) ---
# --- Go build + vet gate ---
if [ -f go.mod ]; then
  echo "go build + vet..." >&2
  if ! go build ./... 2>&1; then
    echo "차단: go build 실패." >&2
    exit 2
  fi
  if ! go vet ./... 2>&1; then
    echo "차단: go vet 실패." >&2
    exit 2
  fi
  # golangci-lint (있으면) — v1 은 --fast, v2(2025-03+)는 --fast-only 로 플래그가 바뀜
  if command -v golangci-lint >/dev/null 2>&1; then
    fast_flag="--fast"
    golangci-lint run --help 2>/dev/null | grep -q -- '--fast-only' && fast_flag="--fast-only"
    if ! golangci-lint run "$fast_flag" ./... 2>&1; then
      echo "차단: golangci-lint 실패." >&2
      exit 2
    fi
  fi
fi

echo "pre-commit 통과"
exit 0
