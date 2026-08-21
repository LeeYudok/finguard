//
//  룰 회귀 픽스처 — finguard.swift.insecure-trust (#33)
//
//  표준 API 두 개(SecTrustSetExceptions / URLCredential(trust:)) 외에 국내 금융 iOS 의
//  사실상 표준 스택인 Alamofire·Starscream 의 검증 무력화 관용구를 함께 본다.
//  합성 코드이며 실제 앱 코드가 아니다.
//

import Foundation

final class TrustBypassSamples {

    // MARK: - 표준 API (기존 커버리지 회귀 방어)

    func acceptAnyCertificate(_ trust: SecTrust, blob: CFData) {
        // EXPECT: finguard.swift.insecure-trust
        SecTrustSetExceptions(trust, blob)
    }

    func credentialFromTrust(_ trust: SecTrust) -> URLCredential {
        // EXPECT: finguard.swift.insecure-trust
        return URLCredential(trust: trust)
    }

    // MARK: - Alamofire 5 — 평가기 자체를 비활성화

    func alamofire5Session() -> Any {
        // EXPECT: finguard.swift.insecure-trust
        let evaluators = ["api.example-bank.co.kr": DisabledTrustEvaluator()]
        return evaluators
    }

    // MARK: - Alamofire 5 — 모든 호스트 평가 강제 해제

    func alamofire5Manager() -> Any {
        // EXPECT: finguard.swift.insecure-trust
        return ServerTrustManager(allHostsMustBeEvaluated: false, evaluators: [:])
    }

    // MARK: - Alamofire 4 — 정책을 disableEvaluation 으로 지정

    func alamofire4Policies() -> Any {
        // EXPECT: finguard.swift.insecure-trust
        let policies = ["api.example-bank.co.kr": ServerTrustPolicy.disableEvaluation]
        return policies
    }

    // MARK: - Starscream 4 — 자가서명 인증서 허용

    func starscream4Security() -> Any {
        // EXPECT: finguard.swift.insecure-trust
        return FoundationSecurity(allowSelfSigned: true)
    }

    // MARK: - Starscream 3 — 인증서 검증 플래그를 끔

    func starscream3Socket(_ socket: WebSocketLike) {
        // EXPECT: finguard.swift.insecure-trust
        socket.disableSSLCertValidation = true
    }
}
