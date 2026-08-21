// 룰 회귀 픽스처 — finguard.go.hardcoded-secret 공통 어휘 통일 (#36).
// 값은 전부 합성 더미다. testdata 하위라 go 툴체인이 컴파일 대상으로 삼지 않는다.
package fixtures

const (
	// EXPECT: finguard.go.hardcoded-secret
	tradingAccessToken = "eyJhbGciOiJIUzI1NiJ9.c3ViOjEyMzQ1"

	// EXPECT: finguard.go.hardcoded-secret
	ledgerEncryptionKey = "8f2b41c7d90e5a63b18c47f2e0d95a31"

	// EXPECT: finguard.go.hardcoded-secret
	keystorePassphrase = "Vault-Store-2026-Qz73"

	// 단독 `key` 는 공통 어휘에서 제외한다.
	key     = "Authorization-Bearer-Header"
	sortKey = "settlement_date_desc"
)
