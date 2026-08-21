// Spring MVC 컨트롤러 파라미터 애노테이션을 taint source 로 인식하는지 확인하는 정탐 픽스처 (#29).
// 확장 전에는 pattern-sources 가 Servlet 원시 API 5개뿐이라 아래가 전부 미검출이었다.
// 대조군은 safe_java_spring_taint.java.
package finguard.fixtures;

import java.io.File;
import java.net.URL;
import java.sql.Connection;
import java.sql.Statement;
import javax.naming.directory.DirContext;
import javax.servlet.http.HttpServletResponse;
import org.springframework.expression.ExpressionParser;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestHeader;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class SpringTaintController {

    private Connection conn;
    private DirContext ctx;
    private ExpressionParser parser;
    private HttpServletResponse response;

    @PostMapping("/webhooks")
    public String registerWebhook(@RequestBody WebhookForm form) throws Exception {
        // EXPECT: finguard.java.ssrf
        return new URL(form.getCallbackUrl()).getHost();
    }

    @GetMapping("/receipts")
    public String receipt(@RequestParam("name") String name) throws Exception {
        // EXPECT: finguard.java.path-traversal
        File f = new File(name);
        return f.getPath();
    }

    @GetMapping("/go/{to}")
    public void go(@PathVariable("to") String to) throws Exception {
        // EXPECT: finguard.java.open-redirect
        response.sendRedirect(to);
    }

    @GetMapping("/reports")
    public void report(@RequestHeader("X-Report-Cmd") String cmd) throws Exception {
        // EXPECT: finguard.java.os-command-injection
        Runtime.getRuntime().exec(cmd);
    }

    // 같은 source 블록을 공유하는 나머지 taint 룰도 함께 확장된다.
    @GetMapping("/accounts")
    public void find(@RequestParam("q") String q) throws Exception {
        Statement st = conn.createStatement();
        // EXPECT: finguard.java.jdbc-sqli
        st.executeQuery(q);
    }

    @GetMapping("/directory")
    public void directory(@RequestParam("filter") String filter) throws Exception {
        // EXPECT: finguard.java.ldap-injection
        ctx.search("ou=users,dc=example,dc=com", filter, null);
    }

    @PostMapping("/expressions")
    public void evaluate(@RequestBody String expression) throws Exception {
        // EXPECT: finguard.java.ssti
        parser.parseExpression(expression);
    }
}
