// 룰 회귀 픽스처 — finguard.ts.hardcoded-secret-test (#70, 원 정책 #25). 값은 전부 합성 더미다.
// 이 파일은 `*.test.ts` 라 운영 룰(finguard.ts.hardcoded-secret)의 대상에서 제외되고
// 값 형태 조건이 강한 테스트 전용 룰만 적용된다.

// EXPECT: finguard.ts.hardcoded-secret-test
const apiKey = "1234567890123456";

// EXPECT: finguard.ts.hardcoded-secret-test
const password = "correcthorsebatterystaple";

// EXPECT: finguard.ts.hardcoded-secret-test
const dbPassword = "Pa$$w0rd!ProdBank";

const fixture = {
  // EXPECT: finguard.ts.hardcoded-secret-test
  accessToken: "eyJhbGciOiJIUzI1NiJ9.c3ViOjEyMzQ1",
};

// 값 형태 게이트에서 걸러지는 짧은 더미 — 테스트 픽스처의 관용값은 잡지 않는다.
const testPassword = "test1234";
const shortSecret = "pw1234";

// 플레이스홀더·안내문 억제는 운영 룰과 동일하다.
const maskedApiKey = "****************";
const merchantSecret = "발급받은 시크릿을 입력하세요";
