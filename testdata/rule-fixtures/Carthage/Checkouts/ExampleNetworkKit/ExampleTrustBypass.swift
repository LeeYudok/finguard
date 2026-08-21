//
//  룰 회귀 픽스처 — finguard.swift.insecure-trust-vendor (#33, 픽스처는 #60)
//
//  이 룰은 `paths.include` 로 **경로 구조 자체**를 판정 기준으로 삼는다
//  (Pods/**, Carthage/**, *.xcframework/**). 그래서 최상위에 파일을 두는 것으로는
//  검증할 수 없고, 반드시 벤더 경로 하위에 픽스처가 있어야 한다.
//  마커 파서가 하위 디렉터리를 순회하지 않던 동안에는 그게 불가능해서 이 룰만
//  회귀 픽스처 없이 머지됐다 (#60).
//
//  경로가 Pods/ 가 아니라 Carthage/ 인 이유: finguard 의 DefaultExcludes 에
//  "Pods" 가 들어 있어 `CLI.Scan` 이 semgrep 에 `--exclude Pods` 를 넘긴다.
//  즉 Pods/ 하위는 룰의 include 와 무관하게 스캔 자체에서 빠진다. Carthage/ 는
//  DefaultExcludes 에 없어 실제 운영 스캔에서도 이 룰이 발화하는 경로다.
//
//  합성 코드이며 실제 라이브러리 코드가 아니다.
//

import Foundation

final class VendorSessionDelegate {

    // 표준 API — 벤더 소스라도 최종 바이너리에서 그대로 실행된다.
    func credential(for trust: SecTrust) -> URLCredential {
        // EXPECT: finguard.swift.insecure-trust-vendor
        return URLCredential(trust: trust)
    }

    // Alamofire 5 관용구.
    func evaluators() -> Any {
        // EXPECT: finguard.swift.insecure-trust-vendor
        return ["api.example-bank.co.kr": DisabledTrustEvaluator()]
    }

    // 대조군 — 정상 평가기는 대상이 아니다.
    func pinnedEvaluators() -> Any {
        return ["api.example-bank.co.kr": PublicKeysTrustEvaluator()]
    }
}
