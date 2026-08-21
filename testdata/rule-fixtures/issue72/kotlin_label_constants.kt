package com.example.finguard

// #72 FP3 — "식별자 == 값" 인 대문자 라벨 상수는 맵 키·화면 코드지 시크릿이 아니다.
// 값이 대문자라는 이유만으로 배제하면 실키가 죽으므로, 아래 정탐 4건은 계속 검출돼야 한다.
object WebAppConstant {

    // 오탐 — 값이 식별자의 마지막 세그먼트와 정확히 같은 라벨 상수
    const val MAIN_CODE_ENCKEYPAD = "ENCKEYPAD"
    const val LOGINPWD = "LOGINPWD"

    // 정탐 — 값이 전부 대문자지만 식별자와 다르다(실제 키)
    // EXPECT: finguard.kotlin.hardcoded-secret
    const val API_KEY = "AKIAIOSFODNN7EXAMPLE"

    // 정탐 — 라벨처럼 생겼어도 식별자와 값이 다르면 배제하지 않는다
    // EXPECT: finguard.kotlin.hardcoded-crypto-material
    const val ENCKEY = "ENCKEYPAD"

    // 정탐 — 하드코드 salt
    // EXPECT: finguard.kotlin.hardcoded-crypto-material
    const val SALT = "Q0ZFRTQ1NkE3ODkw"

    // 정탐 — 평범한 하드코드 비밀번호
    // EXPECT: finguard.kotlin.hardcoded-secret
    private val loginPassword = "P@ssw0rd2026!x"
}
