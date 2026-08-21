// 룰 회귀 픽스처 — finguard.ts.hardcoded-secret 공통 어휘·억제 이식 (#70, 원 이슈 #36·#23).
// 값은 전부 합성 더미다.

// ── 정탐: 공통 어휘 (기존 TS 어휘 5개로는 전부 미탐이었다) ──────────────

// EXPECT: finguard.ts.hardcoded-secret
const ledgerEncryptionKey = "8f2b41c7d90e5a63b18c47f2e0d95a31";

// SCREAMING_SNAKE 식별자도 같은 어휘로 잡는다.
// (#70 이 예로 든 `ENCRYPT_KEY` 는 finguard.ts.hardcoded-crypto-material 이 이미 잡아
//  한 줄에 두 룰이 걸린다. 마커는 줄당 하나라 여기서는 어휘가 겹치지 않는 값을 쓴다.)
// EXPECT: finguard.ts.hardcoded-secret
const ADMIN_PASSWD = "Zt91KbW3Nn74Rp08";

// EXPECT: finguard.ts.hardcoded-secret
const tradingAccessToken = "eyJhbGciOiJIUzI1NiJ9.c3ViOjEyMzQ1";

// EXPECT: finguard.ts.hardcoded-secret
const keystorePassphrase = "Vault-Store-2026-Qz73";

// EXPECT: finguard.ts.hardcoded-secret
const settlementSigningKey = "b71f4e0a92d63c85f0a7e14b2c69d830";

// ── 정탐: TS 고유 구문 ────────────────────────────────────────────────

// 타입 주석이 붙으면 식별자 바로 뒤가 `=` 가 아니다.
// EXPECT: finguard.ts.hardcoded-secret
const gatewayCredential: string = "Zt91-KbW3-Nn74-Rp08";

const pgConfig = {
  // 객체 리터럴 프로퍼티 — 선언 키워드가 없다.
  // EXPECT: finguard.ts.hardcoded-secret
  apiKey: "live_sk_9f31c8a05be24d77",
  timeoutMs: 3000,
};

export class VaultClient {
  // 클래스 필드.
  // EXPECT: finguard.ts.hardcoded-secret
  private readonly masterKey = "0a5f8c3e91b74d26af08e5b1c7d39240";
}

export function verify(inputSecret: string): boolean {
  // 하드코드 자격증명과의 동등 비교.
  // EXPECT: finguard.ts.hardcoded-secret
  return inputSecret === "Adm1n-Console-2026!";
}

// ── 오탐 방지 ────────────────────────────────────────────────────────

// 단독 `key` 는 공통 어휘에서 제외한다 — map key·헤더명 오탐이 폭증한다.
const key = "Authorization-Bearer-Header";
const sortKey = "settlement_date_desc";
const rowKey = "PAYMENT_HISTORY_ID";

// 값이 환경변수 이름(SCREAMING_SNAKE)이거나 snake_case 식별자면 시크릿이 아니다.
const apiKeyEnvName = "PAYMENT_API_KEY";
const secretLookupField = "merchant_secret_column";

// 마스킹 플레이스홀더는 시크릿이 아니라 가리는 코드다 (#23).
const maskedPassword = "****************";
const clientSecret = "your-client-secret";
const refreshToken = "CHANGE_ME";

// 한국어 안내문은 입력 지시문이다 (#23).
const merchantApiKey = "발급받은 API 키를 입력하세요";
const storePassphrase = "여기에 입력";

// 값이 짧으면(8자 미만) 자격증명으로 보지 않는다.
const pwd = "1234";

// 값이 아니라 참조다.
const accessToken = process.env.ACCESS_TOKEN;
const privateKey = loadPrivateKey();

declare function loadPrivateKey(): string;
