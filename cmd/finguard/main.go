// finguard — 소스코드 취약점 점검 봇.
//
//	finguard scan  --dir <경로> [--format report|rdjsonl]   로컬 디렉터리 점검
//	finguard serve --addr :8480                              MR webhook 서버
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/leeyudok/finguard/internal/gitlab"
	"github.com/leeyudok/finguard/internal/mapping"
	"github.com/leeyudok/finguard/internal/rdjson"
	"github.com/leeyudok/finguard/internal/repoconfig"
	"github.com/leeyudok/finguard/internal/runner"
	"github.com/leeyudok/finguard/internal/scanner"
	"github.com/leeyudok/finguard/internal/webhook"

	"bytes"
	"net/http"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	log.SetPrefix("finguard ")
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "사용법: finguard <scan|serve> [flags]")
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "scan":
		err = cmdScan(os.Args[2:])
	case "serve":
		err = cmdServe(os.Args[2:])
	default:
		err = fmt.Errorf("알 수 없는 서브커맨드: %s", os.Args[1])
	}
	if err != nil {
		log.Fatal(err)
	}
}

func defaultPath(rel string) string {
	exe, err := os.Executable()
	if err != nil {
		return rel
	}
	return filepath.Join(filepath.Dir(exe), rel)
}

func cmdScan(args []string) error {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	dir := fs.String("dir", "", "점검할 로컬 소스 디렉터리")
	rules := fs.String("rules", defaultPath("rules"), "Semgrep 커스텀 룰 경로")
	mapPath := fs.String("mapping", defaultPath("mapping/rules.yaml"), "룰ID 매핑 테이블")
	format := fs.String("format", "report", "출력 형식: report | rdjsonl")
	semgrepBin := fs.String("semgrep", "semgrep", "semgrep 바이너리")
	fs.Parse(args)

	if *dir == "" {
		return fmt.Errorf("--dir 은 필수")
	}
	tbl, err := mapping.Load(*mapPath)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	findings, err := scanner.CLI{Bin: *semgrepBin}.Scan(ctx, *rules, *dir)
	if err != nil {
		return err
	}
	switch *format {
	case "rdjsonl":
		written, skipped, err := rdjson.Emit(os.Stdout, findings, tbl)
		if err != nil {
			return err
		}
		log.Printf("rdjsonl %d건 출력, 매핑 없음 %d건 생략", written, skipped)
	case "report":
		printReport(findings, tbl)
	default:
		return fmt.Errorf("알 수 없는 format: %s", *format)
	}
	return nil
}

// indent 는 여러 줄 문자열 각 줄 앞에 prefix 를 붙인다.
func indent(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		if l != "" {
			lines[i] = prefix + l
		}
	}
	return strings.Join(lines, "\n")
}

// printReport 는 로컬 점검용 사람이 읽는 리포트를 출력한다.
// 코멘트와 동일하게 매핑 테이블 값만 근거로 쓴다.
func printReport(findings []scanner.Finding, tbl mapping.Table) {
	bySeverity := map[string][]string{}
	skipped := 0
	for _, f := range findings {
		r, ok := tbl.Lookup(f.RuleID)
		if !ok {
			skipped++
			continue
		}
		basis := r.Basis
		if r.KisaItem != "" {
			basis += " / 금보원 평가항목 " + r.KisaItem
		}
		var b strings.Builder
		fmt.Fprintf(&b, "  %s:%d  [%s] %s (%s)\n      근거: %s",
			f.Path, f.StartLine, r.Code, r.Title, r.CWE, basis)
		if ex := strings.TrimSpace(r.Explain); ex != "" {
			fmt.Fprintf(&b, "\n      %s", ex)
		}
		if fix := strings.TrimSpace(r.FixExample); fix != "" {
			fmt.Fprintf(&b, "\n      수정 예시:\n%s", indent(fix, "      "))
		}
		bySeverity[r.Severity] = append(bySeverity[r.Severity], b.String())
	}
	total := 0
	for _, sev := range []string{"ERROR", "WARNING", "INFO"} {
		lines := bySeverity[sev]
		if len(lines) == 0 {
			continue
		}
		sort.Strings(lines)
		fmt.Printf("\n== %s (%d건) ==\n", sev, len(lines))
		for _, l := range lines {
			fmt.Println(l)
		}
		total += len(lines)
	}
	fmt.Printf("\n총 %d건 (매핑 없어 생략 %d건)\n", total, skipped)
}

func cmdServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", ":8480", "listen 주소")
	rules := fs.String("rules", defaultPath("rules"), "Semgrep 커스텀 룰 경로")
	mapPath := fs.String("mapping", defaultPath("mapping/rules.yaml"), "룰ID 매핑 테이블")
	semgrepBin := fs.String("semgrep", "semgrep", "semgrep 바이너리")
	fs.Parse(args)

	secret := os.Getenv("FINGUARD_WEBHOOK_SECRET")
	if secret == "" {
		return fmt.Errorf("FINGUARD_WEBHOOK_SECRET 이 비어 있음")
	}
	token := os.Getenv("REVIEWDOG_GITLAB_API_TOKEN")
	apiURL := os.Getenv("CI_API_V4_URL")
	if token == "" || apiURL == "" {
		return fmt.Errorf("REVIEWDOG_GITLAB_API_TOKEN / CI_API_V4_URL 이 비어 있음")
	}
	tbl, err := mapping.Load(*mapPath)
	if err != nil {
		return err
	}
	sg := scanner.CLI{Bin: *semgrepBin}
	rd := runner.CLI{}
	blockOnGlobal := repoconfig.GlobalBlockOn()
	log.Printf("차단 게이트 전역 기본: %v (FINGUARD_BLOCK_ON)", blockOnGlobal)

	// 동시 스캔 상한 — semgrep 이 무제한으로 뜨는 것을 막는다.
	maxScans := 2
	if v := os.Getenv("FINGUARD_MAX_SCANS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxScans = n
		}
	}
	sem := make(chan struct{}, maxScans)

	handle := func(ev webhook.MREvent) {
		sem <- struct{}{}
		defer func() { <-sem }()
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()
		sha := ev.ObjectAttributes.LastCommit.ID
		log.Printf("MR 점검 시작: project=%d mr=%d sha=%s", ev.Project.ID, ev.ObjectAttributes.IID, sha)

		// payload 는 위조 가능 — clone URL 이 신뢰 GitLab 을 가리킬 때만 진행한다.
		// (아니면 토큰이 임의 호스트로 전송될 수 있다.)
		if err := gitlab.ValidateCloneHost(apiURL, ev.Project.GitHTTPURL); err != nil {
			log.Printf("clone URL 거부: %v (project=%d)", err, ev.Project.ID)
			return
		}

		// 내부 오류를 MR 에서 보이게 하는 실패 보고 — 게이트 판정과 무관한 오류다.
		fail := func(what string, err error) {
			log.Printf("%s: %v", what, err)
			if serr := gitlab.SetCommitStatus(ctx, apiURL, token, ev.Project.ID, sha,
				"failed", what+" — 재푸시 시 재시도"); serr != nil {
				log.Printf("failed status 게시 실패: %v", serr)
			}
		}

		if err := gitlab.SetCommitStatus(ctx, apiURL, token, ev.Project.ID, sha,
			"running", "취약점 점검 중"); err != nil {
			log.Printf("running status 게시 실패(점검은 계속): %v", err)
		}

		work, err := os.MkdirTemp("", "finguard-*")
		if err != nil {
			fail("작업 디렉터리 생성 실패", err)
			return
		}
		defer os.RemoveAll(work)

		src := filepath.Join(work, "src")
		if err := gitlab.FetchSource(ctx, ev.Project.GitHTTPURL, token, ev.ObjectAttributes.IID, sha, src); err != nil {
			fail("소스 확보 실패", err)
			return
		}
		cfg, found, err := repoconfig.Load(src)
		if err != nil {
			// 설정 파일 오류가 점검 자체를 막지 않도록 전역 기본으로 진행한다.
			log.Printf(".finguard.yml 무시(전역 기본 적용): %v", err)
			cfg = repoconfig.Config{}
		} else if found {
			log.Printf(".finguard.yml 적용: ignore=%d block_on 오버라이드=%v",
				len(cfg.Ignore), cfg.BlockOn != nil)
		}

		findings, err := sg.Scan(ctx, *rules, src)
		if err != nil {
			fail("스캔 실패", err)
			return
		}
		kept := findings[:0:0]
		ignored := 0
		for _, f := range findings {
			if cfg.Ignored(f.Path) {
				ignored++
				continue
			}
			kept = append(kept, f)
		}
		if ignored > 0 {
			log.Printf("ignore 패턴으로 %d건 제외", ignored)
		}
		var buf bytes.Buffer
		written, skipped, err := rdjson.Emit(&buf, kept, tbl)
		if err != nil {
			fail("rdjsonl 생성 실패", err)
			return
		}
		log.Printf("rdjsonl %d건, 매핑 없음 %d건 생략", written, skipped)
		if written > 0 {
			env := runner.Env{
				APIToken:  token,
				APIV4URL:  apiURL,
				ProjectID: strconv.Itoa(ev.Project.ID),
				MRIID:     strconv.Itoa(ev.ObjectAttributes.IID),
				CommitSHA: sha,
			}
			if err := rd.Post(ctx, &buf, env); err != nil {
				fail("reviewdog 실패", err)
				return
			}
		}

		// 스캔이 정상 완료된 경우에만 commit status 를 게시한다.
		var sevs []string
		for _, f := range kept {
			if r, ok := tbl.Lookup(f.RuleID); ok {
				sevs = append(sevs, r.Severity)
			}
		}
		blockOn := repoconfig.EffectiveBlockOn(blockOnGlobal, cfg)
		state, desc := repoconfig.Gate(sevs, blockOn)
		if err := gitlab.SetCommitStatus(ctx, apiURL, token, ev.Project.ID, sha, state, desc); err != nil {
			log.Printf("commit status 게시 실패: %v", err)
			return
		}
		log.Printf("MR 점검 완료: project=%d mr=%d status=%s (%s)",
			ev.Project.ID, ev.ObjectAttributes.IID, state, desc)
	}

	mux := http.NewServeMux()
	mux.Handle("/webhook", webhook.Handler(secret, handle))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	log.Printf("listen %s (동시 스캔 상한 %d)", *addr, maxScans)
	srv := &http.Server{
		Addr:              *addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}
	return srv.ListenAndServe()
}
