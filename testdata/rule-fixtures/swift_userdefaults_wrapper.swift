//
//  룰 회귀 픽스처 — finguard.swift.userdefaults-sensitive (#31)
//
//  @propertyWrapper 로 저장 메커니즘과 키 이름이 다른 선언으로 갈라지는 Swift 관용구.
//  단일 라인 정규식은 이 형태에서 연결이 끊긴다. 합성 코드이며 실제 키 값은 없다.
//

import Foundation
import SwiftUI

@propertyWrapper
struct UserInfo<T> {
    private let key: String
    private let defaultValue: T

    var wrappedValue: T {
        // 저장 메커니즘 선언부는 잡지 않는다 — 무엇을 담는지 이 줄만으로는 알 수 없다
        get { return UserDefaults.standard.object(forKey: key) as? T ?? defaultValue }
        set { UserDefaults.standard.set(newValue, forKey: key) }
    }

    init(key: String, defaultValue: T) {
        self.key = key
        self.defaultValue = defaultValue
    }
}

struct ExchangeKeys {
    // EXPECT: finguard.swift.userdefaults-sensitive
    @UserInfo(key: ExchangeKey.accessKey.rawValue, defaultValue: "") var accessToken: String
    // EXPECT: finguard.swift.userdefaults-sensitive
    @UserInfo(key: ExchangeKey.secretKey.rawValue, defaultValue: "") var secretToken: String
    // 민감 어휘가 없는 키는 대상이 아니다
    @UserInfo(key: ExchangeKey.refreshInterval.rawValue, defaultValue: 1) var refreshInterval: Int
}

struct SettingsView {
    // EXPECT: finguard.swift.userdefaults-sensitive
    @AppStorage("auth_access_token") var authToken: String = ""
    // 화면 설정 값은 대상이 아니다
    @AppStorage("is_dark_mode") var isDarkMode: Bool = false
}
