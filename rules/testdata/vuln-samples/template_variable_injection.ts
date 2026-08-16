// 회귀 픽스처 — FIN-INJ-001 (finguard.ts.eval).
//
// 실제 사례: toss/es-toolkit 의 lodash 호환 template() 에서 발견해 제보한
// 코드 주입 취약점(lodash CVE-2021-23337 동종). variable 옵션이 검증 없이
// 컴파일되는 함수 시그니처에 보간되어 new Function 으로 실행된다.
//
//   template('', { variable: 'a = globalThis.__PWNED__ = true' })()  // 임의 코드 실행
//
// finguard 의 ts.eval 룰이 이 new Function 호출을 반드시 검출해야 한다.
// 룰을 손볼 때 이 케이스가 조용히 빠지지 않도록 골든 테스트로 고정한다.

interface TemplateOptions {
  variable?: string;
  imports?: Record<string, unknown>;
}

export function vulnerableTemplate(source: string, options: TemplateOptions) {
  const imports = { ...options.imports };
  const importsKeys = Object.keys(imports);
  const importValues = Object.values(imports);

  // variable 이 검증 없이 함수 헤더에 삽입된다 (취약 지점)
  const compiledFunction = `function(${options.variable || 'obj'}) {
    let __p = '';
    ${options.variable ? source : `with(obj) {\n${source}\n}`}
    return __p;
  }`;

  // FIN-INJ-001 이 검출해야 하는 지점
  return new Function(...importsKeys, `return ${compiledFunction}`)(...importValues);
}
