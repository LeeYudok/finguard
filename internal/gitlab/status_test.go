package gitlab

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetCommitStatus(t *testing.T) {
	var gotPath, gotToken, gotState, gotName, gotDesc string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotToken = r.Header.Get("PRIVATE-TOKEN")
		if err := r.ParseForm(); err != nil {
			t.Errorf("form 파싱 실패: %v", err)
		}
		gotState = r.FormValue("state")
		gotName = r.FormValue("name")
		gotDesc = r.FormValue("description")
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	err := SetCommitStatus(context.Background(), srv.URL, "tok", 124, "abc123", "failed", "차단 대상 지적 2건 (ERROR)")
	if err != nil {
		t.Fatalf("err = %v, 원하는 값 nil", err)
	}
	if gotPath != "/projects/124/statuses/abc123" {
		t.Errorf("path = %s", gotPath)
	}
	if gotToken != "tok" {
		t.Errorf("PRIVATE-TOKEN = %s, 원하는 값 tok", gotToken)
	}
	if gotState != "failed" || gotName != "finguard" || gotDesc == "" {
		t.Errorf("form: state=%s name=%s desc=%s", gotState, gotName, gotDesc)
	}
}

func TestSetCommitStatusHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	err := SetCommitStatus(context.Background(), srv.URL, "tok", 1, "sha", "success", "d")
	if err == nil {
		t.Fatal("err = nil, 원하는 값 403 에러")
	}
}
