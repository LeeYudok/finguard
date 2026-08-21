// ts_framework_taint_sources.ts 의 대조군 — 어떤 룰에도 걸리면 안 된다.
// safe_ 접두 파일은 EXPECT 마커를 가질 수 없다.
import { Body, Controller, Post } from "@nestjs/common";

declare const ALLOW_HOSTS: string[];

@Controller("safe-payments")
export class SafePaymentController {
  // 데코레이터 파라미터를 받되 허용 목록으로 검증하고, 호출 대상은 고정 상수다.
  @Post("webhook")
  async register(@Body() input: { url: string }): Promise<unknown> {
    const u = new URL(input.url);
    if (!ALLOW_HOSTS.includes(u.hostname)) {
      throw new Error("허용되지 않은 호스트");
    }
    return fetch("https://api.internal.example.com/webhooks/register");
  }
}

// 설정 상수(SCREAMING_CASE)에서 온 URL 은 이용자가 지정한 값이 아니므로 stored-ssrf 대상이 아니다.
export async function ping(config: { WEBHOOK_URL: string }): Promise<void> {
  await fetch(config.WEBHOOK_URL, { method: "POST" });
}
