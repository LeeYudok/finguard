"""룰 회귀 픽스처 — finguard.python.hardcoded-secret-test (#25).

파일명이 test_*.py 라는 이유로 blanket exclude 되던 구간이다. 값은 전부 합성 더미다.
"""

# EXPECT: finguard.python.hardcoded-secret-test
API_KEY = "1234567890123456"

# EXPECT: finguard.python.hardcoded-secret-test
PASSWORD = "correcthorsebatterystaple"

# EXPECT: finguard.python.hardcoded-secret-test
DB_PASSWORD = "Pa$$w0rd!ProdBank"

# 게이트에서 걸러지는 값 — 짧은 더미는 잡지 않는다.
TEST_PASSWORD = "test1234"
