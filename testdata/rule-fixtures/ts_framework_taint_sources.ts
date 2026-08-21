// NestJS·nestia 컨트롤러 파라미터 데코레이터를 taint source 로 인식하는지 확인하는 정탐 픽스처 (#29).
// 확장 전에는 pattern-sources 가 $REQ.query/body/params/headers 뿐이라 아래가 전부 미검출이었다.
// 대조군은 safe_ts_framework_taint.ts.
import { Body, Controller, Get, Headers, Param, Post, Query } from "@nestjs/common";
import core from "@nestia/core";
import { exec } from "child_process";
import fs from "fs";

declare const res: any;

@Controller("payments")
export class PaymentController {
  @Post("webhook")
  async register(@Body() input: { url: string }): Promise<unknown> {
    // EXPECT: finguard.ts.ssrf
    return fetch(input.url);
  }

  @Get("receipt")
  receipt(@Query() q: { name: string }): Buffer {
    // EXPECT: finguard.ts.path-traversal
    return fs.readFileSync(q.name);
  }

  @Get("go/:to")
  go(@Param() p: { to: string }): void {
    // EXPECT: finguard.ts.open-redirect
    res.redirect(p.to);
  }

  @Post("report")
  report(@Headers() h: Record<string, string>): void {
    // EXPECT: finguard.ts.os-command-injection
    exec(h["x-report-cmd"]);
  }

  // nestia 의 암호화 본문 — samchon/payments 가 실제로 쓰는 형태
  @Post("encrypted")
  async encrypted(@core.EncryptedBody() body: { target: string }): Promise<unknown> {
    // EXPECT: finguard.ts.ssrf
    return fetch(body.target);
  }

  @Post("typed")
  async typed(@core.TypedBody() body: { target: string }): Promise<unknown> {
    // EXPECT: finguard.ts.ssrf
    return fetch(body.target);
  }
}

// 2차(stored) SSRF — 저장된 콜백 URL 을 그대로 호출한다.
// taint 로는 잡히지 않아 별도 regex 룰이 담당한다.
export async function sendWebhook(history: { webhook_url: string }): Promise<void> {
  // EXPECT: finguard.ts.stored-ssrf
  await fetch(history.webhook_url, { method: "POST" });
}
