// java_spring_taint_sources.java 의 대조군 — 어떤 룰에도 걸리면 안 된다.
// safe_ 접두 파일은 EXPECT 마커를 가질 수 없다.
package finguard.fixtures;

import java.net.URL;
import java.sql.Connection;
import java.sql.PreparedStatement;
import java.util.List;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class SafeSpringController {

    private Connection conn;
    private List<String> allowedHosts;

    // 애노테이션 파라미터를 받되 허용 목록으로 검증하고, 호출 대상은 고정 상수다.
    @GetMapping("/safe/webhooks")
    public String registerWebhook(@RequestParam("target") String target) throws Exception {
        URL candidate = new URL("https://api.internal.example.com/webhooks/register");
        if (!allowedHosts.contains(target)) {
            throw new IllegalArgumentException("허용되지 않은 호스트");
        }
        return candidate.getHost();
    }

    // 바인딩 파라미터 — 조립이 없다.
    @GetMapping("/safe/accounts")
    public void find(@RequestParam("q") String q) throws Exception {
        PreparedStatement ps = conn.prepareStatement("SELECT * FROM ACCT WHERE NAME = ?");
        ps.setString(1, q);
        ps.executeQuery();
    }
}
