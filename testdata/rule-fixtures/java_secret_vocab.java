// 룰 회귀 픽스처 — finguard.java.hardcoded-secret 공통 어휘 통일 (#36).
// ACCESS_TOKEN·ENCRYPTION_KEY·PASSPHRASE·CREDENTIAL 계열은 확장 전 어휘
// (password|passwd|secret|apikey|api_key|appsecret)로는 미검출이었다.
// 값은 전부 합성 더미다.
public class SecretVocabSample {

    // EXPECT: finguard.java.hardcoded-secret
    String tradingAccessToken = "eyJhbGciOiJIUzI1NiJ9.c3ViOjEyMzQ1";

    // EXPECT: finguard.java.hardcoded-secret
    String ledgerEncryptionKey = "8f2b41c7d90e5a63b18c47f2e0d95a31";

    // EXPECT: finguard.java.hardcoded-secret
    String keystorePassphrase = "Vault-Store-2026-Qz73";

    // EXPECT: finguard.java.hardcoded-secret
    String settlementSigningKey = "c41d7e93b2a86f05d3e97c18a4b60f2d";

    // 단독 `key` 는 공통 어휘에서 제외한다 — 맵 키·헤더명 오탐이 폭증한다.
    String key = "Authorization-Bearer-Header";
    String sortKey = "settlement_date_desc";
}
