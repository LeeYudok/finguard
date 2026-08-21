// #72 FP1 — 라인/블록 주석·문자열 안의 죽은 eval 은 검출하지 않는다.
// 아래 "살아있는" 호출 4건은 정탐이므로 반드시 계속 검출돼야 한다.

export function runCallbacks(callback: string, data: unknown) {
  // EXPECT: finguard.ts.eval
  eval(callback);

  // EXPECT: finguard.ts.eval
  const fn = new Function("return 1");

  // EXPECT: finguard.ts.eval
  window.eval(callback);

  // 같은 줄에서 닫힌 블록 주석 뒤의 실행 코드도 정탐이다.
  // EXPECT: finguard.ts.eval
  const guarded = data /* note */ && eval(callback);

  // 여기부터는 실행되지 않는 텍스트 — 검출되면 오탐 회귀다.
  // eval(signData.signNorBack)(data);
  // new Function("signData.signNorBack()")(error);
    //   eval(signData.signCloseBack)();
  /* eval(deadBlockComment) */
  /**
   * eval(docComment)
   */
  const label = "eval(insideStringLiteral)";
  const single = 'new Function("insideStringLiteral")';
  void fn; // eval(trailingLineComment)

  return [guarded, label, single];
}
