// 룰 회귀 픽스처 — finguard.swift.hardcoded-secret 공통 어휘 통일 (#36).
// 기존 단독 `token`·`credential` 은 유지하고 PASSPHRASE·ENCRYPTION_KEY 계열을 더한다.
// 값은 전부 합성 더미다.
struct SecretVocabSample {

    // EXPECT: finguard.swift.hardcoded-secret
    let ledgerEncryptionKey = "8f2b41c7d90e5a63b18c47f2e0d95a31"

    // EXPECT: finguard.swift.hardcoded-secret
    let keystorePassphrase = "Vault-Store-2026-Qz73"

    // 기존 커버리지(단독 token)가 유지되는지 고정한다.
    // EXPECT: finguard.swift.hardcoded-secret
    let sessionToken = "eyJhbGciOiJIUzI1NiJ9.c3ViOjEyMzQ1"

    // 단독 `key` 는 공통 어휘에서 제외한다.
    let key = "Authorization-Bearer-Header"
}
