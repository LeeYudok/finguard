// 룰 회귀 픽스처 — finguard.java.hardcoded-secret-test (#25).
// 파일명이 *Test.java 라는 이유로 blanket exclude 되던 구간이다. 값의 "시크릿다움"
// 게이트로 노이즈만 거르고 실키 형태는 ERROR 로 잡는다. 값은 전부 합성 더미다.
public class SecretNonprodTest {

    // 이슈 #25 의 헤드라인 증거 형태 — 16자리 "숫자" API 키.
    // 원안의 AND 게이트(영문 AND 숫자 AND 20자 이상)는 이 값을 못 잡았다.
    // EXPECT: finguard.java.hardcoded-secret-test
    String apiKey = "1234567890123456";

    // 순영문 패스프레이즈 — 원안 게이트는 숫자가 없어 못 잡는다.
    // EXPECT: finguard.java.hardcoded-secret-test
    String password = "correcthorsebatterystaple";

    // 특수문자 포함 — 원안 게이트는 허용 문자군 밖이라 못 잡는다.
    // EXPECT: finguard.java.hardcoded-secret-test
    String dbPassword = "Pa$$w0rd!ProdBank";

    // 80자 랜덤 시크릿
    // EXPECT: finguard.java.hardcoded-secret-test
    String apiSecret = "9c1f4a7e2b8d06531fa4c7e90b2d5638a1c4f7e0b3d6592c8f1a4e7b0d3c6592f8a1b4e7c0d3f6a92";

    // 게이트에서 걸러지는 값들 — 짧은 더미·상수명·플레이스홀더는 잡지 않는다.
    String testPassword = "test1234";
    String apiKeyName = "PAYMENT_API_KEY";
    String secretPlaceholder = "changeme";
}
