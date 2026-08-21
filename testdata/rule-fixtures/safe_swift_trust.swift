//
//  대조군 — 어떤 룰에도 걸리면 안 된다 (#33).
//
//  기본 신뢰 평가를 그대로 쓰고, 자가서명 허용·평가 비활성화 플래그를 전부 끈 정상 구현.
//

import Foundation

final class TrustSafeSamples {

    func alamofire5Session() -> Any {
        // 기본 평가기 — 인증서 체인과 호스트명을 모두 검증한다
        let evaluators = ["api.example-bank.co.kr": DefaultTrustEvaluator()]
        return ServerTrustManager(allHostsMustBeEvaluated: true, evaluators: evaluators)
    }

    func starscream4Security() -> Any {
        return FoundationSecurity(allowSelfSigned: false)
    }

    func starscream3Socket(_ socket: WebSocketLike) {
        socket.disableSSLCertValidation = false
    }

    func challengeHandler(_ challenge: URLAuthenticationChallenge) -> URLSession.AuthChallengeDisposition {
        // 기본 처리에 위임 — 시스템 신뢰 평가를 우회하지 않는다
        return .performDefaultHandling
    }
}
