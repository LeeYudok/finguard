// finguard.java.dev-auth-endpoint (#37) 대조군 — 어떤 룰에도 걸리면 안 된다.
//
// 검증 항목
//   - 정상 로그인: 자격증명 검증 후 토큰 발급. 경로에 개발 어휘가 없다.
//   - /api/v1/devices: "dev" 를 부분문자열로 포함하지만 경로 세그먼트가 아니다.
//   - /local/config: 개발 경로지만 토큰 발급 호출이 없다.
//   - testReport: 메서드명에 test 가 들어간 정상 업무 엔드포인트.
package fixtures;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/v1")
public class SafeAuthController {

    private JwtTokenProvider jwtTokenProvider;
    private AuthService authService;
    private DeviceService deviceService;

    @PostMapping("/auth/login")
    public TokenResponse login(@RequestBody LoginRequest request) {
        authService.verifyCredential(request.getEmail(), request.getRawPassword());
        return jwtTokenProvider.generateAccessToken(request.getEmail());
    }

    @GetMapping("/devices/list")
    public DeviceList devices() {
        return deviceService.list();
    }

    @GetMapping("/local/config")
    public ConfigView localConfig() {
        return ConfigView.current();
    }

    @PostMapping("/settlement/test-report")
    public ReportView testReport(@RequestBody ReportRequest request) {
        return deviceService.report(request);
    }
}
