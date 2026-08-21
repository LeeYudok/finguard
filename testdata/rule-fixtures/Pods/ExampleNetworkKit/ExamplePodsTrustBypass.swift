//
//  룰 회귀 픽스처 — finguard.swift.insecure-trust-vendor, Pods/ 경로 (#75)
//
//  이 파일의 존재 자체가 #75 의 회귀 방어다.
//
//  이전에는 `DefaultExcludes` 에 "Pods" 가 있어 `CLI.Scan` 이 semgrep 에
//  `--exclude Pods` 를 넘겼고, 전역 exclude 는 룰의 `paths.include` 보다 먼저
//  타겟을 걸러내므로 이 룰의 `Pods/**` 분기는 **운영 스캔에서 도달 불가**였다.
//  즉 있지도 않은 커버리지를 있는 것처럼 보이게 하는 죽은 설정이었다.
//
//  #75 에서 벤더 제외를 전역이 아니라 룰별 `paths.exclude` 로 옮겼다. 일반 Swift 룰
//  9개는 각자 Pods/Carthage 를 제외하고, 벤더 경로를 의도적으로 점검하는 이 룰만
//  거기에 도달한다. DefaultExcludes 에 "Pods" 를 되돌리면 아래 마커가 미탐 회귀로
//  즉시 실패한다.
//
//  합성 코드이며 실제 라이브러리 코드가 아니다.
//

import Foundation

final class PodsVendorSessionDelegate {

    // 벤더 소스의 인증서 검증 무력화는 최종 바이너리에 그대로 실려 나간다.
    func credential(for trust: SecTrust) -> URLCredential {
        // EXPECT: finguard.swift.insecure-trust-vendor
        return URLCredential(trust: trust)
    }

    // Starscream 3 관용구.
    func makeSocket() -> Any {
        let socket = ExampleWebSocket()
        // EXPECT: finguard.swift.insecure-trust-vendor
        socket.disableSSLCertValidation = true
        return socket
    }
}
