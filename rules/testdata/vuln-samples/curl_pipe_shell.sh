#!/usr/bin/env bash
# finguard.shell.curl-pipe-shell 정탐 픽스처 (#19).
# 아래 줄들은 모두 원격 스크립트를 무결성 검증 없이 실행하므로 반드시 검출돼야 한다.
set -euo pipefail

# 1) 기본형
curl -LsSf https://example.com/install.sh | sh

# 2) 들여쓰기 + 실패 폴백
    curl -fsSL https://example.com/install.sh | sh || echo "설치 실패"

# 3) bash 로 파이프
curl -fsSL https://example.com/install.sh | bash

# 4) wget
wget -qO- https://example.com/install.sh | sh

# 5) sudo 승격
curl -sL https://example.com/install.sh | sudo bash

# 6) 셸 실행 명령의 인자 — 따옴표 안이지만 실제로 실행된다
bash -c "curl -fsSL https://example.com/install.sh | sh"

# 7) 제어구문 안
if [ ! -x /usr/local/bin/tool ]; then curl https://example.com/install.sh | sh; fi
