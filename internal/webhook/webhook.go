// Package webhook 은 GitLab MR 이벤트를 수신·검증한다.
// 실패해도 GitLab 재전송 폭주를 막기 위해 200 을 반환하고 로그만 남긴다.
package webhook

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

// MREvent 는 GitLab merge_request 이벤트 payload 중 finguard 이 쓰는 필드만 담는다.
type MREvent struct {
	ObjectKind string `json:"object_kind"`
	User       struct {
		Username string `json:"username"`
	} `json:"user"`
	Project struct {
		ID         int    `json:"id"`
		GitHTTPURL string `json:"git_http_url"`
		WebURL     string `json:"web_url"`
	} `json:"project"`
	ObjectAttributes struct {
		IID    int    `json:"iid"`
		Action string `json:"action"`
		// Oldrev 는 update 이벤트에서 코드 푸시일 때만 채워진다.
		// 라벨/제목/설명 변경 update 에는 없다 → 이걸로 재스캔 여부를 판정한다.
		Oldrev         string `json:"oldrev"`
		WorkInProgress bool   `json:"work_in_progress"`
		Draft          bool   `json:"draft"`
		LastCommit     struct {
			ID string `json:"id"`
		} `json:"last_commit"`
	} `json:"object_attributes"`
	// Changes 는 update 이벤트에서 바뀐 속성의 전/후 값을 담는다.
	// draft 해제(ready 전환)를 oldrev 없는 update 와 구분하는 데 쓴다.
	Changes struct {
		Draft struct {
			Previous bool `json:"previous"`
			Current  bool `json:"current"`
		} `json:"draft"`
		WorkInProgress struct {
			Previous bool `json:"previous"`
			Current  bool `json:"current"`
		} `json:"work_in_progress"`
	} `json:"changes"`
}

// draftCleared 는 이번 update 가 draft/WIP 해제(ready 전환)인지 판정한다.
// draft 상태에서 쌓인 커밋은 이 시점 전까지 스캔된 적이 없으므로 스캔 대상이다.
func draftCleared(ev MREvent) bool {
	c := ev.Changes
	return (c.Draft.Previous && !c.Draft.Current) ||
		(c.WorkInProgress.Previous && !c.WorkInProgress.Current)
}

// skipReason 은 처리하지 않을 이벤트면 사유 문자열을, 처리 대상이면 "" 를 반환한다.
func skipReason(ev MREvent) string {
	if ev.ObjectKind != "merge_request" {
		return "merge_request 아님"
	}
	switch a := ev.ObjectAttributes.Action; a {
	case "open", "reopen":
	case "update":
		if ev.ObjectAttributes.Oldrev == "" && !draftCleared(ev) {
			return "코드 변경 없는 update(라벨/제목 등)"
		}
	default:
		return "무관 action: " + a
	}
	if ev.ObjectAttributes.WorkInProgress || ev.ObjectAttributes.Draft {
		return "draft/WIP MR"
	}
	return ""
}

// Handler 는 X-Gitlab-Token 검증 후 스캔 대상 MR 이벤트만 handle 로 넘긴다.
// handle 은 별도 고루틴에서 돌고, 응답은 항상 즉시 200 이다(토큰 불일치만 401).
// 같은 project/iid 스캔이 진행 중이면 최신 이벤트 1개를 보관했다가 현재 스캔이
// 끝난 뒤 이어서 실행한다 — 연타 푸시로 훅이 겹칠 때 이중 스캔·코멘트 레이스를
// 막으면서, 마지막 푸시가 스캔에서 누락되지 않게 한다.
func Handler(secret string, handle func(MREvent)) http.Handler {
	var (
		mu       sync.Mutex
		inflight = map[string]bool{}
		pending  = map[string]MREvent{}
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusOK)
			return
		}
		got := r.Header.Get("X-Gitlab-Token")
		if subtle.ConstantTimeCompare([]byte(got), []byte(secret)) != 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			log.Printf("webhook body 읽기 실패: %v", err)
			w.WriteHeader(http.StatusOK)
			return
		}
		var ev MREvent
		if err := json.Unmarshal(body, &ev); err != nil {
			log.Printf("webhook payload 파싱 실패: %v", err)
			w.WriteHeader(http.StatusOK)
			return
		}
		if reason := skipReason(ev); reason != "" {
			log.Printf("webhook 스킵: %s (project=%d mr=%d user=%s)",
				reason, ev.Project.ID, ev.ObjectAttributes.IID, ev.User.Username)
			w.WriteHeader(http.StatusOK)
			return
		}
		key := fmt.Sprintf("%d/%d", ev.Project.ID, ev.ObjectAttributes.IID)
		mu.Lock()
		if inflight[key] {
			pending[key] = ev
			mu.Unlock()
			log.Printf("webhook 보류: 스캔 진행 중, 완료 후 이어서 실행 (project/mr=%s)", key)
			w.WriteHeader(http.StatusOK)
			return
		}
		inflight[key] = true
		mu.Unlock()
		go func(ev MREvent) {
			for {
				handle(ev)
				mu.Lock()
				next, ok := pending[key]
				if !ok {
					delete(inflight, key)
					mu.Unlock()
					return
				}
				delete(pending, key)
				mu.Unlock()
				log.Printf("보류 이벤트 이어서 실행 (project/mr=%s)", key)
				ev = next
			}
		}(ev)
		w.WriteHeader(http.StatusOK)
	})
}
