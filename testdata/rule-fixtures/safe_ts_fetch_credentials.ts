// 오탐 방지 대조군 — finguard.ts.hardcoded-secret (#90).
// Fetch 표준 옵션 `credentials` 의 값은 RequestCredentials 열거값(omit|same-origin|include)이라
// 시크릿이 아니라 정책 지시자다. 이 파일은 어떤 룰에도 걸리지 않아야 한다(safe_ 접두 규약).

export async function getJson(path: string): Promise<Response> {
  return fetch(path, { credentials: "same-origin" });
}

export async function postAction(path: string): Promise<Response> {
  return fetch(path, { method: "POST", credentials: "same-origin" });
}

// 열거값 셋 전부 — 어느 것도 시크릿이 아니다.
export const anonymous: RequestInit = { credentials: "omit" };
export const crossSite: RequestInit = { credentials: "include" };

// 홑따옴표·공백 변형도 같은 값이다.
export const singleQuoted: RequestInit = { credentials: 'same-origin' };
export const spaced: RequestInit = { credentials : "same-origin" };

// 스트리밍 구독처럼 다른 옵션과 섞여도 마찬가지다.
export async function subscribe(url: string, signal: AbortSignal): Promise<Response> {
  return fetch(url, {
    credentials: "same-origin",
    signal,
  });
}
