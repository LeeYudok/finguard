// 대조 픽스처 — 취약하지 않은 안전한 구현.
// eval 도 new Function 도 쓰지 않으므로 FIN-INJ-001 이 검출하면 안 된다.

interface TemplateOptions {
  variable?: string;
}

export function safeTemplate(source: string, _options: TemplateOptions) {
  // 정적 치환만 — 동적 코드 컴파일 없음
  return (data: Record<string, string>) =>
    source.replace(/\{\{(\w+)\}\}/g, (_m, key) => data[key] ?? '');
}
