// finguard.java.dev-auth-endpoint (#37) 정탐 픽스처.
//
// 자격증명 검증 없이 요청 본문만으로 정식 토큰을 발급하는 개발/로컬 전용 엔드포인트.
// 매핑 애노테이션의 경로 세그먼트 어휘(dev/local/debug/impersonate/...) 와 같은 메서드
// 본문의 토큰 발급 호출이 동시에 성립할 때만 검출된다.
package fixtures;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/v1/auth")
public class DevAuthController {

    private JwtTokenProvider jwtTokenProvider;
    private TokenService tokenService;
    private UserRepository userRepository;

    // 요청 본문의 email 만으로 사용자를 upsert 하고 정식 access 토큰을 발급한다.
    @PostMapping("/oauth/local/login")
    public TokenResponse localLogin(@RequestBody RegisterRequest request) {
        User user = userRepository.findOrCreate(request.getEmail());
        // EXPECT: finguard.java.dev-auth-endpoint
        return jwtTokenProvider.generateAccessToken(user.getId());
    }

    // 명시 인자형 애노테이션 + 계정 전환(임의 계정 사칭) 경로.
    @RequestMapping(value = "/debug/switch-user", method = RequestMethod.POST)
    public TokenResponse switchUser(@RequestBody RegisterRequest request) {
        User user = userRepository.findByEmail(request.getEmail());
        // EXPECT: finguard.java.dev-auth-endpoint
        return tokenService.createToken(user);
    }

    // 제공자 이름에 Jwt 가 들어간 일반 발급 메서드도 토큰 발급으로 본다.
    @PostMapping("/dev-login")
    public TokenResponse devLogin(@RequestBody RegisterRequest request) {
        // EXPECT: finguard.java.dev-auth-endpoint
        return jwtProvider.generate(request.getEmail());
    }

    // 개발 전용 경로지만 토큰을 발급하지 않으므로 검출 대상이 아니다.
    @GetMapping("/dev/health")
    public String devHealth() {
        return "ok";
    }
}
