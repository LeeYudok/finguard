#!/usr/bin/env bash
# #72 FP2 — heredoc 본문은 출력되는 안내 문구지 실행되는 요청이 아니다.
# heredoc 밖의 실제 요청은 정탐이므로 반드시 계속 검출돼야 한다.
set -euo pipefail

# EXPECT: finguard.shell.curl-insecure
curl -sk https://api.example.com/health

cat <<EOF
사용법:
    검증: curl -sk https://localhost:8543/JBN  → 200
    curl --insecure https://localhost:8543/status
EOF

# EXPECT: finguard.shell.curl-insecure
curl --insecure https://internal.example.com/v1

cat <<'MSG'
    curl -k https://doc.example.com
MSG

cat <<-EOT
	curl -k https://tab.example.com
EOT

# EXPECT: finguard.shell.curl-insecure
sudo curl -kL https://real.example.com/install
