BINARY  := finguard
PKG     := ./cmd/finguard
LDFLAGS := -s -w

.PHONY: build build-linux test test-integration clean

# 호스트(개발 머신)용 빌드
build:
	go build -trimpath -ldflags '$(LDFLAGS)' -o bin/$(BINARY) $(PKG)

# 폐쇄망 리눅스 서버 반입용 정적 바이너리
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build -trimpath -ldflags '$(LDFLAGS)' -o bin/$(BINARY)-linux-amd64 $(PKG)

test:
	go test -race ./...

# 실제 semgrep 을 돌려 룰이 취약 패턴을 검출하는지 확인하는 통합 테스트.
# semgrep 바이너리가 필요하다(없으면 각 테스트가 skip).
test-integration:
	go test -race -tags semgrep_integration ./...

clean:
	rm -rf bin/
