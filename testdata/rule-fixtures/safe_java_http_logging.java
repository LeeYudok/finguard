// finguard.java.http-client-full-logging (#32) 대조군 — 어떤 룰에도 걸리면 안 된다.
package fixtures;

import feign.Logger;
import okhttp3.logging.HttpLoggingInterceptor;

public class SafeHttpClientLoggingConfig {

    // 메서드·URL·상태코드·소요시간만 남는다.
    public Logger.Level feignLoggerLevel() { return Logger.Level.BASIC; }

    // NONE 은 아무것도 남기지 않는다.
    public Logger.Level feignNoneLevel() { return Logger.Level.NONE; }

    public HttpLoggingInterceptor okHttpInterceptor() {
        HttpLoggingInterceptor interceptor = new HttpLoggingInterceptor();
        interceptor.setLevel(HttpLoggingInterceptor.Level.NONE);
        return interceptor;
    }

    // 레벨 상수를 참조만 하고 클라이언트에 적용하지 않는 헬퍼.
    public String describeLevel() { return "BASIC"; }
}
