// 룰 회귀 픽스처 — finguard.kotlin.hardcoded-secret 공통 어휘 통일 (#36).
// 값은 전부 합성 더미다.
object SecretVocabSample {

    // EXPECT: finguard.kotlin.hardcoded-secret
    const val TRADING_ACCESS_TOKEN = "eyJhbGciOiJIUzI1NiJ9.c3ViOjEyMzQ1"

    // EXPECT: finguard.kotlin.hardcoded-secret
    private val ledgerEncryptionKey = "8f2b41c7d90e5a63b18c47f2e0d95a31"

    // EXPECT: finguard.kotlin.hardcoded-secret
    val keystorePassphrase: String = "Vault-Store-2026-Qz73"

    // 단독 `key` 는 공통 어휘에서 제외한다.
    val key = "Authorization-Bearer-Header"
    val sortKey = "settlement_date_desc"
}
