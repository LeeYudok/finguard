// 언어 패리티 보강 룰(#27) 정탐 픽스처 — SQL 조립 · 암호재료 · 난수 · TLS.
// 대조군은 safe_ts_parity.ts.
import crypto from "crypto";
import https from "https";

declare const prisma: any;
declare const pool: any;
declare const conn: any;
declare const Prisma: any;
declare const rawQuery: string;
declare const schema: string;
declare const reader: string;
declare const tableName: string;
declare const filter: string;
declare const acct: string;
declare const data: Buffer;
declare const keyBuf: Buffer;
declare const passphrase: string;

export async function run(): Promise<void> {
  // 변수 하나만 넘기는 raw 실행 — 조립 흔적이 호출부에 없어도 unsafe API 는 잡는다
  // EXPECT: finguard.ts.sql-injection
  await prisma.$queryRawUnsafe(rawQuery);

  // 템플릿 보간 DDL
  // EXPECT: finguard.ts.sql-injection
  await prisma.$executeRawUnsafe(`GRANT SELECT ON SCHEMA ${schema} TO ${reader}`);

  // 태그드 템플릿 안의 Prisma.raw — 파라미터화를 재개통하는 경로
  // EXPECT: finguard.ts.sql-injection
  await prisma.$queryRaw`SELECT * FROM ${Prisma.raw(tableName)}`;

  // 보간과 바인딩을 섞은 형태 — 인자 개수로 안전을 판정하지 않는다
  // EXPECT: finguard.ts.sql-injection
  await pool.query(`SELECT * FROM users WHERE id = ${filter} AND status = $1`, ["A"]);

  // 문자열 연결
  // EXPECT: finguard.ts.sql-injection
  await conn.query("SELECT * FROM ledger WHERE acct = '" + acct + "'");
}

// 하드코드된 암호화 키
// 식별자를 ENCRYPT_KEY 에서 CIPHER_KEY 로 바꿨다 (#70). 어휘 이식 후 ENCRYPT_KEY 는
// hardcoded-crypto-material 과 hardcoded-secret 에 동시에 걸리는데(둘 다 정탐이고
// 근거 조항이 다르다), 마커는 줄당 하나뿐이라 이 줄에서 두 기대값을 표현할 수 없다.
// cipher 어휘는 시크릿 공통 어휘에 없어 이 룰만 단독으로 검증된다.
// EXPECT: finguard.ts.hardcoded-crypto-material
const CIPHER_KEY = "A1b2C3d4E5f6G7h8";

// IV 가 없는 deprecated 암호 API
// EXPECT: finguard.ts.hardcoded-crypto-material
const legacyCipher = crypto.createCipher("aes-256-cbc", passphrase);

// IV 가 문자열 리터럴
// EXPECT: finguard.ts.hardcoded-crypto-material
const decipher = crypto.createDecipheriv("aes-256-cbc", keyBuf, "0123456789abcdef");

// 취약한 해시
// EXPECT: finguard.ts.weak-crypto
const digest = crypto.createHash("md5").update(data).digest("hex");

// ECB 운영모드
// EXPECT: finguard.ts.weak-crypto
const ecbCipher = crypto.createCipheriv("aes-256-ecb", keyBuf, null);

// 예측 가능한 난수로 OTP 생성 — 이름 조건 충족
// EXPECT: finguard.ts.insecure-random
const otpCode = String(Math.floor(Math.random() * 1000000));

// TLS 검증 비활성화
// EXPECT: finguard.ts.tls-verify-disabled
const agent = new https.Agent({ rejectUnauthorized: false });

export { CIPHER_KEY, legacyCipher, decipher, digest, ecbCipher, otpCode, agent };
