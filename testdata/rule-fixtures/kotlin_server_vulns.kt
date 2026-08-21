// Kotlin Spring Boot 백엔드 취약 패턴 픽스처 (#28).
// 기존 kotlin 룰은 전부 Android 관용구라 이런 서버 코드에서 한 건도 발화하지 않았다.
// 기대값은 검출 줄 바로 위의 EXPECT 마커에 적는다.

package fixtures.kotlin.server

import feign.Logger
import java.io.File
import java.io.FileInputStream
import java.net.URL
import java.security.MessageDigest
import javax.crypto.Mac
import jakarta.servlet.http.HttpServletRequest
import jakarta.servlet.http.HttpServletResponse
import org.slf4j.LoggerFactory
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.PathVariable
import org.springframework.web.bind.annotation.PostMapping
import org.springframework.web.bind.annotation.RequestParam
import org.springframework.web.bind.annotation.RestController

private val log = LoggerFactory.getLogger("fixture")

// ── 서버측 민감정보 로깅 (log-sensitive 싱크 확장) ───────────────────────────

class LoggingSamples {

    fun issueToken(token: String, password: String, secret: String) {
        // EXPECT: finguard.kotlin.log-sensitive
        log.info("issued: $token")
        // EXPECT: finguard.kotlin.log-sensitive
        log.debug("login password=$password")
        // EXPECT: finguard.kotlin.log-sensitive
        println("client secret=$secret")
        // EXPECT: finguard.kotlin.log-sensitive
        System.out.println("pin=$password")
    }

    // 부분 마스킹 — 카드번호는 가렸지만 password 는 그대로 흐른다.
    // 라인에 mask 어휘가 있다는 이유로 통째로 억제하면 이 실취약이 사라진다(#28 검증 반박).
    fun partiallyMasked(cardNo: String, password: String) {
        // EXPECT: finguard.kotlin.log-sensitive
        log.debug("card=${mask(cardNo)}, password=$password")
    }

    private fun mask(v: String): String = v.takeLast(4).padStart(v.length, '*')
}

// ── 서버측 SQL 조립 (sql-raw 싱크 확장) ─────────────────────────────────────

class OrderQueryRepository(private val entityManager: FakeEntityManager, private val jdbcTemplate: FakeJdbcTemplate) {

    fun findByStatus(status: String): List<String> {
        // EXPECT: finguard.kotlin.sql-raw
        return entityManager.createQuery("SELECT o FROM Orders o WHERE o.status = '$status'")
    }

    fun findNative(userId: String): List<String> {
        // EXPECT: finguard.kotlin.sql-raw
        return entityManager.createNativeQuery("SELECT * FROM ORDERS WHERE USER_ID = '" + userId + "'")
    }

    fun countByBranch(branch: String): List<String> {
        // EXPECT: finguard.kotlin.sql-raw
        return jdbcTemplate.queryForObject("SELECT COUNT(*) FROM ACCT WHERE BRANCH = '$branch'")
    }
}

class FakeEntityManager {
    fun createQuery(sql: String): List<String> = listOf(sql)
    fun createNativeQuery(sql: String): List<String> = listOf(sql)
}

class FakeJdbcTemplate {
    fun queryForObject(sql: String): List<String> = listOf(sql)
    fun query(sql: String): List<String> = listOf(sql)
}

// ── 취약 해시 (weak-hash 확장: 알고리즘명을 문자열로 받는 래퍼) ──────────────

object ExchangeSignUtil {

    fun legacyDigest(payload: ByteArray): ByteArray {
        // EXPECT: finguard.kotlin.weak-hash
        return MessageDigest.getInstance("MD5").digest(payload)
    }

    fun legacyMac(): Mac {
        // EXPECT: finguard.kotlin.weak-hash
        return Mac.getInstance("HmacSHA1")
    }

    // 거래소 서명 유틸이 실제로 쓰던 형태 — 알고리즘명이 사내 래퍼의 문자열 인자로 온다.
    // 인자가 다음 줄로 넘어가도 잡혀야 한다.
    fun sign(plain: String, secret: String): String {
        // EXPECT: finguard.kotlin.weak-hash
        return CryptUtil.encrypt(
            "MD5",
            (plain + secret).toByteArray(Charsets.UTF_8)
        ).toString()
    }
}

object CryptUtil {
    fun encrypt(algorithm: String, data: ByteArray): ByteArray =
        MessageDigest.getInstance(algorithm).digest(data)
}

// ── Feign 전체 로깅 ─────────────────────────────────────────────────────────

class PaymentFeignConfig {
    // EXPECT: finguard.kotlin.feign-full-logging
    fun feignLoggerLevel(): Logger.Level = Logger.Level.FULL
}

// ── 개발용 인증 우회 엔드포인트 ─────────────────────────────────────────────

@RestController
class DevAuthController {

    // EXPECT: finguard.kotlin.dev-backdoor-endpoint
    @PostMapping("/oauth/local/login")
    fun localDevLogin(): String = "issued"
}

// ── 스택트레이스 노출 ───────────────────────────────────────────────────────

class MailSender {
    fun send() {
        try {
            error("boom")
        } catch (e: IllegalStateException) {
            // EXPECT: finguard.kotlin.stacktrace-exposure
            e.printStackTrace()
        }
    }
}

// ── taint 룰 (Kotlin 문법 sink) ─────────────────────────────────────────────

@RestController
class TaintController {

    @GetMapping("/ops/ping")
    fun ping(@RequestParam host: String): String {
        // 외부 입력이 명령 문자열에 섞인다. OS 명령 삽입 전용 taint 룰은 두지 않았고,
        // 기존 finguard.kotlin.os-command 가 모든 exec 호출을 ERROR 로 잡는다.
        // EXPECT: finguard.kotlin.os-command
        val p = Runtime.getRuntime().exec("ping -c 1 $host")
        return p.toString()
    }

    @GetMapping("/statement/{name}")
    fun statement(@PathVariable name: String): String {
        // EXPECT: finguard.kotlin.path-traversal
        val f = File("/data/statements/$name")
        return f.readText()
    }

    @GetMapping("/statement/stream")
    fun stream(@RequestParam("file") fileName: String): Int {
        // EXPECT: finguard.kotlin.path-traversal
        return FileInputStream("/data/statements/$fileName").read()
    }

    @GetMapping("/proxy")
    fun proxy(@RequestParam target: String): String =
        // EXPECT: finguard.kotlin.ssrf
        URL(target).readText()

    fun legacyProxy(request: HttpServletRequest): String {
        val target = request.getParameter("url")
        // EXPECT: finguard.kotlin.ssrf
        return URL(target).readText()
    }

    @GetMapping("/go")
    fun go(@RequestParam next: String, response: HttpServletResponse) {
        // EXPECT: finguard.kotlin.open-redirect
        response.sendRedirect(next)
    }
}
