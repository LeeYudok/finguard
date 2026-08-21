// 룰 회귀 픽스처 — finguard.kotlin.hardcoded-secret-test (#25). 값은 전부 합성 더미다.
object SecretNonprodTest {

    // EXPECT: finguard.kotlin.hardcoded-secret-test
    val apiKey = "1234567890123456"

    // EXPECT: finguard.kotlin.hardcoded-secret-test
    val password = "correcthorsebatterystaple"

    // EXPECT: finguard.kotlin.hardcoded-secret-test
    private val dbPassword = "Pa\$\$w0rd!ProdBank"

    // 게이트에서 걸러지는 값 — 짧은 더미는 잡지 않는다.
    val testPassword = "test1234"
}
