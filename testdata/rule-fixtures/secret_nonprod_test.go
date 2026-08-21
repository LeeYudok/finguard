// 룰 회귀 픽스처 — finguard.go.hardcoded-secret-test (#25). 값은 전부 합성 더미다.
package fixtures

const (
	// EXPECT: finguard.go.hardcoded-secret-test
	apiKey = "1234567890123456"

	// EXPECT: finguard.go.hardcoded-secret-test
	password = "correcthorsebatterystaple"

	// EXPECT: finguard.go.hardcoded-secret-test
	dbPassword = "Pa$$w0rd!ProdBank"

	// 게이트에서 걸러지는 값 — 짧은 더미는 잡지 않는다.
	testPassword = "test1234"
)
