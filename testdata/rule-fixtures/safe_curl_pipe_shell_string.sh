#!/usr/bin/env bash
# finguard.shell.curl-pipe-shell 오탐 방지 픽스처 (#19).
# 아래 줄들은 명령을 "실행" 하지 않는다 — 어떤 룰에도 걸리면 안 된다.
set -euo pipefail

info() { printf '[info] %s\n' "$1"; }
warn() { printf '[warn] %s\n' "$1"; }

# 1) 안내 로그 — 실코드(zusik setup.sh)에서 실제로 오탐났던 형태
info "설치 중: curl -fsSL https://example.com/install.sh | sh"
info "설치 중: curl -fsSL https://example.com/install.sh | bash"

# 2) echo / printf 로 사용법 출력
echo "수동 설치: curl -LsSf https://example.com/install.sh | sh"
printf '%s\n' "curl https://example.com/install.sh | bash"

# 3) 경고 문구 자체가 이 패턴을 인용
warn "curl ... | sh 는 무결성 검증이 없어 권장하지 않습니다"

# 4) 변수에 담기만 함
INSTALL_HINT="curl https://example.com/install.sh | sh"

# 5) 주석
# curl https://example.com/install.sh | sh

# 6) 권장 형태 — 다운로드와 실행을 분리하고 체크섬을 검증한다
curl -fsSLo install.sh https://example.com/install.sh
echo "0000000000000000000000000000000000000000000000000000000000000000  install.sh" \
  | sha256sum -c - && bash install.sh
