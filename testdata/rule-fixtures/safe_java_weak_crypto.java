// finguard.java.weak-hash · finguard.java.weak-cipher-wrapper (#34) 대조군 —
// 어떤 룰에도 걸리면 안 된다.
package fixtures;

import java.security.MessageDigest;
import javax.crypto.Mac;

public class SafeSignUtil {

    private static final String HASH_ALGORITHM = "SHA-256";
    private static final String CIPHER_ALGORITHM = "AES/GCM/NoPadding";

    public byte[] digest(byte[] raw) throws Exception {
        MessageDigest md = MessageDigest.getInstance(HASH_ALGORITHM);
        return md.digest(raw);
    }

    public Mac mac() throws Exception {
        return Mac.getInstance("HmacSHA256");
    }

    public String sign(String plainText) {
        return CryptUtil.encrypt("SHA-256", plainText);
    }

    // 함수명 휴리스틱에 걸리지 않는 임의 함수의 첫 인자는 알고리즘 지정이 아니다.
    public String describe() {
        return registry.resolve("MD5");
    }

    // 알고리즘명이 아니라 지원 목록 표기다.
    public String[] supported() {
        return new String[] {"SHA-256", "SHA-512"};
    }
}
