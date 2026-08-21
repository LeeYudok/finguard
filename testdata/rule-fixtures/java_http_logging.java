// finguard.java.http-client-full-logging (#32) 정탐 픽스처.
//
// 설정 한 줄로 요청·응답 전체를 로그에 흘리는 구조. 마커 없는 줄은 오탐 방지 대조군이다.
package fixtures;

import feign.Logger;
import okhttp3.logging.HttpLoggingInterceptor;

public class HttpClientLoggingConfig {

    // Feign 전역 로깅 레벨 — 헤더와 본문이 통째로 남는다.
    // EXPECT: finguard.java.http-client-full-logging
    public Logger.Level feignLoggerLevel() { return Logger.Level.FULL; }

    // HEADERS 도 Authorization 헤더를 그대로 남긴다.
    // EXPECT: finguard.java.http-client-full-logging
    public Logger.Level feignHeaderLevel() { return Logger.Level.HEADERS; }

    // OkHttp 본문 로깅.
    public HttpLoggingInterceptor okHttpInterceptor() {
        HttpLoggingInterceptor interceptor = new HttpLoggingInterceptor();
        // EXPECT: finguard.java.http-client-full-logging
        interceptor.setLevel(HttpLoggingInterceptor.Level.BODY);
        return interceptor;
    }

    // 대조군 1 — BASIC 은 메서드·URL·상태코드만 남긴다.
    public Logger.Level safeLoggerLevel() { return Logger.Level.BASIC; }

    // 대조군 2 — wire 로거를 끄는 설정은 오히려 수정된 코드다.
    public void disableWireLog() {
        java.util.logging.Logger.getLogger("org.apache.http.wire").setLevel(java.util.logging.Level.OFF);
    }

    // 대조군 3 — 주석 안의 설정 예시는 실제 설정이 아니다.
    // interceptor.setLevel(HttpLoggingInterceptor.Level.BODY);
    public void documented() { }
}
