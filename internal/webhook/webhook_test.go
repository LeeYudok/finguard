package webhook

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const mrPayload = `{
  "object_kind": "merge_request",
  "project": {"id": 124, "git_http_url": "https://gitlab.example.com/dok123/sample-app.git"},
  "object_attributes": {"iid": 7, "action": "open", "last_commit": {"id": "abc123"}}
}`

func post(h http.Handler, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(body))
	req.Header.Set("X-Gitlab-Token", token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHandlerRejectsBadToken(t *testing.T) {
	h := Handler("secret", func(MREvent) { t.Fatal("호출되면 안 됨") })
	if rec := post(h, "wrong", mrPayload); rec.Code != http.StatusUnauthorized {
		t.Fatalf("401 기대, got %d", rec.Code)
	}
}

func TestHandlerDispatchesMR(t *testing.T) {
	ch := make(chan MREvent, 1)
	h := Handler("secret", func(ev MREvent) { ch <- ev })
	if rec := post(h, "secret", mrPayload); rec.Code != http.StatusOK {
		t.Fatalf("200 기대, got %d", rec.Code)
	}
	select {
	case ev := <-ch:
		if ev.Project.ID != 124 || ev.ObjectAttributes.IID != 7 || ev.ObjectAttributes.LastCommit.ID != "abc123" {
			t.Fatalf("unexpected event: %+v", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("handle 미호출")
	}
}

func TestHandlerSkips(t *testing.T) {
	cases := []struct {
		name     string
		payload  string
		dispatch bool
	}{
		{
			"코드 변경 없는 update(oldrev 없음)",
			`{"object_kind":"merge_request","project":{"id":1},"object_attributes":{"iid":2,"action":"update"}}`,
			false,
		},
		{
			"코드 푸시 update(oldrev 있음)",
			`{"object_kind":"merge_request","project":{"id":1},"object_attributes":{"iid":2,"action":"update","oldrev":"def456"}}`,
			true,
		},
		{
			"draft MR",
			`{"object_kind":"merge_request","project":{"id":1},"object_attributes":{"iid":2,"action":"open","draft":true}}`,
			false,
		},
		{
			"WIP MR",
			`{"object_kind":"merge_request","project":{"id":1},"object_attributes":{"iid":2,"action":"open","work_in_progress":true}}`,
			false,
		},
		{
			"draft 해제 update(oldrev 없음) — ready 전환은 스캔",
			`{"object_kind":"merge_request","project":{"id":1},"object_attributes":{"iid":2,"action":"update"},"changes":{"draft":{"previous":true,"current":false}}}`,
			true,
		},
		{
			"WIP 해제 update(oldrev 없음) — 구버전 필드도 스캔",
			`{"object_kind":"merge_request","project":{"id":1},"object_attributes":{"iid":2,"action":"update"},"changes":{"work_in_progress":{"previous":true,"current":false}}}`,
			true,
		},
		{
			"draft 설정 update — ready→draft 는 스킵",
			`{"object_kind":"merge_request","project":{"id":1},"object_attributes":{"iid":2,"action":"update"},"changes":{"draft":{"previous":false,"current":true}}}`,
			false,
		},
	}
	for _, tc := range cases {
		ch := make(chan MREvent, 1)
		h := Handler("secret", func(ev MREvent) { ch <- ev })
		if rec := post(h, "secret", tc.payload); rec.Code != http.StatusOK {
			t.Fatalf("%s: 200 기대, got %d", tc.name, rec.Code)
		}
		select {
		case <-ch:
			if !tc.dispatch {
				t.Errorf("%s: handle 호출됨, 원하는 값 스킵", tc.name)
			}
		case <-time.After(200 * time.Millisecond):
			if tc.dispatch {
				t.Errorf("%s: handle 미호출, 원하는 값 호출", tc.name)
			}
		}
	}
}

func TestHandlerInflightDedup(t *testing.T) {
	var calls atomic.Int32
	started := make(chan struct{}, 8)
	release := make(chan struct{})
	h := Handler("secret", func(MREvent) {
		calls.Add(1)
		started <- struct{}{}
		<-release
	})

	if rec := post(h, "secret", mrPayload); rec.Code != http.StatusOK {
		t.Fatalf("1차: 200 기대, got %d", rec.Code)
	}
	<-started

	// 같은 project/iid 진행 중 — 스킵돼야 한다.
	if rec := post(h, "secret", mrPayload); rec.Code != http.StatusOK {
		t.Fatalf("2차: 200 기대, got %d", rec.Code)
	}
	time.Sleep(100 * time.Millisecond)
	if got := calls.Load(); got != 1 {
		t.Fatalf("진행 중 중복: handle 호출 %d회, 원하는 값 1", got)
	}

	// 다른 MR 은 진행 중이어도 스캔한다.
	other := strings.Replace(mrPayload, `"iid": 7`, `"iid": 8`, 1)
	post(h, "secret", other)
	<-started
	if got := calls.Load(); got != 2 {
		t.Fatalf("다른 MR: handle 호출 %d회, 원하는 값 2", got)
	}

	// 완료 후 같은 MR 재이벤트는 다시 스캔한다.
	close(release)
	deadline := time.After(time.Second)
	for calls.Load() < 3 {
		post(h, "secret", mrPayload)
		select {
		case <-deadline:
			t.Fatalf("완료 후 재스캔 안 됨: 호출 %d회", calls.Load())
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func TestHandlerPendingRunsLatest(t *testing.T) {
	events := make(chan MREvent, 4)
	release := make(chan struct{})
	h := Handler("secret", func(ev MREvent) {
		events <- ev
		<-release
	})

	post(h, "secret", mrPayload)
	first := <-events

	// 진행 중 도착한 이벤트 2개 — 마지막 것만 보류되어 완료 후 실행돼야 한다.
	stale := strings.Replace(mrPayload, `"abc123"`, `"stale9"`, 1)
	latest := strings.Replace(mrPayload, `"abc123"`, `"def456"`, 1)
	post(h, "secret", stale)
	post(h, "secret", latest)
	close(release)

	select {
	case second := <-events:
		if first.ObjectAttributes.LastCommit.ID != "abc123" {
			t.Errorf("1차 sha = %s, 원하는 값 abc123", first.ObjectAttributes.LastCommit.ID)
		}
		if got := second.ObjectAttributes.LastCommit.ID; got != "def456" {
			t.Errorf("보류 이벤트 sha = %s, 원하는 값 def456 (최신)", got)
		}
	case <-time.After(time.Second):
		t.Fatal("보류 이벤트가 실행되지 않음")
	}
	select {
	case ev := <-events:
		t.Fatalf("3번째 실행 발생 (sha=%s), 보류는 최신 1개만이어야 함", ev.ObjectAttributes.LastCommit.ID)
	case <-time.After(200 * time.Millisecond):
	}
}

func TestHandlerIgnoresOtherEvents(t *testing.T) {
	h := Handler("secret", func(MREvent) { t.Fatal("호출되면 안 됨") })
	if rec := post(h, "secret", `{"object_kind":"push"}`); rec.Code != http.StatusOK {
		t.Fatalf("push 이벤트도 200, got %d", rec.Code)
	}
	if rec := post(h, "secret", `{"object_kind":"merge_request","object_attributes":{"action":"close"}}`); rec.Code != http.StatusOK {
		t.Fatalf("close 액션도 200, got %d", rec.Code)
	}
}
