BINARY  := finguard
PKG     := ./cmd/finguard
LDFLAGS := -s -w

.PHONY: build build-linux test clean

# 호스트(개발 머신)용 빌드
build:
	go build -trimpath -ldflags '$(LDFLAGS)' -o bin/$(BINARY) $(PKG)

# 폐쇄망 리눅스 서버 반입용 정적 바이너리
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build -trimpath -ldflags '$(LDFLAGS)' -o bin/$(BINARY)-linux-amd64 $(PKG)

test:
	go test ./...

clean:
	rm -rf bin/
