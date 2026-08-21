// Kotlin 서버 대조군 (#28) — 어떤 룰에도 걸리면 안 되는 정상 코드.
// 룰 확장이 정상 서버 코드에 오탐을 쏟지 않는지 고정한다.

package fixtures.kotlin.server

import java.io.File
import java.net.URL
import java.security.MessageDigest
import java.security.SecureRandom
import jakarta.servlet.http.HttpServletRequest
import jakarta.servlet.http.HttpServletResponse
import org.slf4j.LoggerFactory
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.PostMapping
import org.springframework.web.bind.annotation.RequestParam
import org.springframework.web.bind.annotation.RestController

private val log = LoggerFactory.getLogger("safe-fixture")

// 레이트리밋 인터셉터 — availableTokens/tokenBucket 은 민감정보가 아니다.
// log-sensitive 의 민감어휘 조건이 \btoken\b 경계라 여기에 걸리지 않아야 한다.
class ThrottlingInterceptor(private val buckets: FakeBucketRegistry) {

    fun preHandle(request: HttpServletRequest): Boolean {
        val remoteAddr = request.remoteAddr
        log.info("remoteAddr : $remoteAddr")

        val bucket = buckets.resolveBucket(remoteAddr)
        val availableTokens = bucket.availableTokens
        log.info("$remoteAddr : $availableTokens")
        log.debug("tokenBucket refillTokens=${bucket.refillTokens} tokenCount=${bucket.tokenCount}")

        return bucket.tryConsume(1)
    }
}

class FakeBucket(val availableTokens: Long, val refillTokens: Long, val tokenCount: Long) {
    fun tryConsume(n: Int): Boolean = availableTokens >= n
}

class FakeBucketRegistry {
    fun resolveBucket(key: String): FakeBucket = FakeBucket(10, 10, 10)
}

// 정상 로깅 — 값 자체를 남기지 않는다.
class AuthService {
    fun issue(accessToken: String) {
        log.info("access token issued (length={})", accessToken.length)
        log.info("login succeeded for user")
    }
}

// 바인딩 파라미터를 쓰는 정상 쿼리.
class SafeRepository(private val entityManager: SafeEntityManager, private val jdbcTemplate: SafeJdbcTemplate) {

    fun findByStatus(status: String): List<String> =
        entityManager.createQuery("SELECT o FROM Orders o WHERE o.status = :status", status)

    fun countByBranch(branch: String): Int =
        jdbcTemplate.queryForObject("SELECT COUNT(*) FROM ACCT WHERE BRANCH = ?", branch)
}

class SafeEntityManager {
    fun createQuery(jpql: String, param: String): List<String> = listOf(jpql, param)
}

class SafeJdbcTemplate {
    fun queryForObject(sql: String, param: String): Int = sql.length + param.length
}

// 강한 해시 · 안전한 난수.
object SafeCrypto {
    fun digest(payload: ByteArray): ByteArray = MessageDigest.getInstance("SHA-256").digest(payload)

    fun nonce(): ByteArray {
        val out = ByteArray(16)
        SecureRandom().nextBytes(out)
        return out
    }
}

// CORS 는 와일드카드가 아니라 명시 오리진 목록 — 걸리면 안 된다.
class CorsConfig {
    val allowedOriginPatterns = listOf(
        "https://bank.example.com",
        "https://admin.example.com"
    )
}

// 검증을 거친 파일 접근 · 리다이렉트 · 외부 호출.
@RestController
class SafeController {

    private val allowedHosts = setOf("api.example.com")

    @GetMapping("/statement")
    fun statement(@RequestParam name: String): String {
        val safeName = File(name).name
        return File("/data/statements", safeName).readText()
    }

    @GetMapping("/proxy")
    fun proxy(@RequestParam target: String): String {
        val u = URL(target)
        require(u.host in allowedHosts) { "blocked host" }
        return "ok"
    }

    @GetMapping("/go")
    fun go(@RequestParam next: String, response: HttpServletResponse) {
        // 외부 입력은 분기 키로만 쓰고, 리다이렉트 대상은 상수 리터럴에서 나온다.
        val dest = when (next) {
            "home" -> "/home"
            "mypage" -> "/mypage"
            else -> "/home"
        }
        response.sendRedirect(dest)
    }

    // 운영 API 는 dev/local 세그먼트가 없어 backdoor 룰에 걸리지 않는다.
    @PostMapping("/api/v1/auth/token/refresh")
    fun refresh(): String = "refreshed"
}

// 예외는 로거로 보낸다 — printStackTrace 를 쓰지 않는다.
class SafeMailSender {
    fun send() {
        try {
            error("boom")
        } catch (e: IllegalStateException) {
            log.warn("메일 발송 실패: {}", e.javaClass.simpleName)
        }
    }
}
