// #66 회귀 픽스처 — semgrep TypeScript 파서의 `export namespace` AST 갭을 고정한다.
//
// (A) export namespace 블록과 (B) top-level 블록은 **같은 미끼 코드**다.
//   - (B) 는 AST 룰이 발화하므로 EXPECT 마커가 붙어 있다. 미끼가 실제로 탐지
//     가능하다는 대조군(control)이다 — 대조군 없이 (A) 의 "0건" 을 해석하면
//     미끼가 원래 안 잡히는 것인지 갭 때문인지 구분할 수 없다.
//   - (A) 는 semgrep 1.169.0 에서 무음이라 마커가 없다.
//
// **마커 부재가 곧 기대값이다.** TestFixtureExpectationsMatchScan 은 양방향
// (마커 있는데 미검출 = 미탐 회귀 / 마커 없는데 검출 = 오탐 회귀)이므로,
// semgrep 업그레이드로 갭이 해소되면 (A) 의 검출이 "마커 없는데 검출됐다" 로
// 테스트를 실패시킨다. 갭 해소가 조용히 지나가지 않고 드러난다.
//
// 그때 할 일: (A) 에 마커를 달고, finguard.ts.namespace-analysis-gap 룰과
// #27·#29 가 넣은 정규식 우회(sql-injection (A) 갈래 · stored-ssrf)를 재검토한다.
//
// 정규식 모드 룰은 이 갭의 영향을 받지 않는다 — (A) 안의 $executeRawUnsafe 가
// sql-injection (A) 갈래로 잡히는 것이 그 증거다.

// EXPECT: finguard.ts.namespace-analysis-gap
export namespace GapDemo {
  // AST 룰 미끼 — semgrep 1.169.0 에서는 아래 전부 무음이다(마커 없음).
  export function osCmd(req: any, cp: any) {
    cp.exec(req.query.cmd);
  }

  export function openRedirect(req: any, res: any) {
    res.redirect(req.query.next);
  }

  export function ssrfCall(req: any) {
    return fetch(req.query.url);
  }

  export function xssSink(el: any, req: any) {
    el.innerHTML = req.body.html;
  }

  export function pathTrav(req: any, fs: any) {
    return fs.readFileSync(req.query.p);
  }

  export function zipSlip(entry: any, fs: any) {
    fs.writeFileSync(entry.entryName, "x");
  }

  export function sqlAstBranch(db: any, id: string) {
    return db.query(`SELECT * FROM payments WHERE id = ${id}`);
  }

  // 정규식 갈래는 블록 안에서도 발화한다.
  export function sqlRegexBranch(prisma: any, script: string) {
    // EXPECT: finguard.ts.sql-injection
    return prisma.$executeRawUnsafe(script);
  }
}

// (B) 대조군 — 같은 코드가 top-level 이면 AST 룰이 정상 발화한다.

export function osCmdTop(req: any, cp: any) {
  // EXPECT: finguard.ts.os-command-injection
  cp.exec(req.query.cmd);
}

export function openRedirectTop(req: any, res: any) {
  // EXPECT: finguard.ts.open-redirect
  res.redirect(req.query.next);
}

export function ssrfCallTop(req: any) {
  // EXPECT: finguard.ts.ssrf
  return fetch(req.query.url);
}

export function xssSinkTop(el: any, req: any) {
  // EXPECT: finguard.ts.xss-innerhtml
  el.innerHTML = req.body.html;
}

export function pathTravTop(req: any, fs: any) {
  // EXPECT: finguard.ts.path-traversal
  return fs.readFileSync(req.query.p);
}

export function zipSlipTop(entry: any, fs: any) {
  // EXPECT: finguard.ts.zip-slip
  fs.writeFileSync(entry.entryName, "x");
}

export function sqlAstBranchTop(db: any, id: string) {
  // EXPECT: finguard.ts.sql-injection
  return db.query(`SELECT * FROM payments WHERE id = ${id}`);
}
