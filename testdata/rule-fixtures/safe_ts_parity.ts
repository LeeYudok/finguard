// 언어 패리티 보강 룰(#27) 대조군 — 어떤 룰에도 걸리면 안 된다.
// safe_ 접두 파일은 EXPECT 마커를 가질 수 없다.
import crypto from "crypto";
import https from "https";

declare const prisma: any;
declare const pool: any;
declare const gqlClient: any;
declare const userId: string;
declare const uid: string;
declare const keyBuf: Buffer;
declare const data: Buffer;

export async function run(): Promise<void> {
  // 상수 리터럴만 넘기는 raw 실행 — 조립 여지가 없다
  await prisma.$queryRawUnsafe("SELECT 1");

  // 태그드 템플릿은 진짜 파라미터화다
  await prisma.$queryRaw`SELECT * FROM users WHERE id = ${userId}`;

  // 완전 바인딩
  await pool.query("SELECT * FROM users WHERE id = $1", [userId]);

  // SQL 이 아닌 query 싱크 — 보간이 있어도 SQL 키워드가 없다
  await gqlClient.query(`{ user(id: "${uid}") { name } }`);
}

// 비보안 용도 난수 — 이름에 보안 어휘가 없다
const elementId = Math.random().toString(36).slice(2);

// iv·seed 를 부분문자열로 보면 걸리는 흔한 식별자들 (drive·receive)
const driveJitter = Math.random() * 100;
const receiveDelay = Math.random() * 100;

// 검증을 켜는 쪽 대입은 걸리지 않아야 한다
process.env.NODE_TLS_REJECT_UNAUTHORIZED = "1";

// 안전한 난수원
const nonce = crypto.randomBytes(12);

// 난수 IV · AEAD 운영모드
const cipher = crypto.createCipheriv("aes-256-gcm", keyBuf, crypto.randomBytes(12));

// 안전한 해시
const digest = crypto.createHash("sha256").update(data).digest("hex");

// 인증서 검증 유지
const agent = new https.Agent({ rejectUnauthorized: true });

export { elementId, driveJitter, receiveDelay, nonce, cipher, digest, agent };
