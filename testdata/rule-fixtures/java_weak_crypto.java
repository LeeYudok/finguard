// finguard.java.weak-hash · finguard.java.weak-cipher-wrapper (#34) 정탐 픽스처.
//
// 금융권 공용 CryptUtil/SignUtil 은 알고리즘명을 문자열 파라미터로 넘기는 구조가 표준에
// 가깝다. 그러면 정의부에는 리터럴이 없고(getInstance(algorithm)) 호출부에는 MessageDigest
// 문자열이 없어 종전 정규식이 양쪽 다 보지 못했다.
// 마커 없는 줄은 오탐 방지 대조군이다.
package fixtures;

import java.security.MessageDigest;
import javax.crypto.Mac;
import org.apache.commons.codec.digest.DigestUtils;

public class SignUtil {

    // 알고리즘명을 상수로 빼고 래퍼에 넘기는 최빈 형태 — 호출부 정규식만으로는 못 본다.
    // EXPECT: finguard.java.weak-hash
    private static final String HASH_ALGORITHM = "MD5";

    // EXPECT: finguard.java.weak-cipher-wrapper
    private static final String CIPHER_ALGORITHM = "DESede/CBC/PKCS5Padding";

    // 공용 래퍼 경유 — 거래소 API 요청 서명 생성 형태.
    public String sign(String plainText) {
        // EXPECT: finguard.java.weak-hash
        return CryptUtil.encrypt("MD5", plainText);
    }

    public String legacyDigest(String plainText) {
        // EXPECT: finguard.java.weak-hash
        return DigestUtils.md5Hex(plainText);
    }

    // 직접 호출 형태도 계속 검출된다(회귀 방지).
    public byte[] directDigest(byte[] raw) throws Exception {
        // EXPECT: finguard.java.weak-hash
        MessageDigest md = MessageDigest.getInstance("SHA-1");
        return md.digest(raw);
    }

    public Mac legacyMac() throws Exception {
        // EXPECT: finguard.java.weak-hash
        return Mac.getInstance("HmacMD5");
    }

    // 래퍼 경유 취약 암호.
    public String encryptLegacy(String plainText) {
        // EXPECT: finguard.java.weak-cipher-wrapper
        return CryptUtil.encrypt("DES/ECB/PKCS5Padding", plainText);
    }

    // 대조군 — SHA-256 은 대상이 아니다.
    public String safeSign(String plainText) {
        return CryptUtil.encrypt("SHA-256", plainText);
    }

    // 대조군 — 함수명 휴리스틱에 걸리지 않는 임의 함수의 첫 인자.
    public String lookupCode() {
        return registry.resolve("MD5");
    }

    // 대조군 — AES 는 대상이 아니다.
    public String encryptModern(String plainText) {
        return CryptUtil.encrypt("AES/GCM/NoPadding", plainText);
    }
}
