#!/bin/bash
# 룰 회귀 픽스처 — finguard.shell.hardcoded-secret 공통 어휘 통일 (#36).
# 값은 전부 합성 더미다.

# EXPECT: finguard.shell.hardcoded-secret
export LEDGER_ENCRYPTION_KEY=8f2b41c7d90e5a63b18c47f2e0d95a31
# EXPECT: finguard.shell.hardcoded-secret
export KEYSTORE_PASSPHRASE=Vault-Store-2026-Qz73
# 단독 `key` 는 공통 어휘에서 제외한다.
export SORT_KEY=settlement_date_desc
