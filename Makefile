.PHONY: default test
all: default test

APPNAME=pump

default:
	go mod tidy&&go fmt ./...&&revive .&&goimports -w .&&golangci-lint run --enable-all&&go install -ldflags="-s -w" ./...

install:
	go install -ldflags="-s -w" ./...

package: install
	upx ~/go/bin/$(APPNAME)
	mv ~/go/bin/$(APPNAME) ~/go/bin/$(APPNAME)-darwin-amd64
	gzip ~/go/bin/$(APPNAME)-darwin-amd64

test:
	go test ./...

# https://hub.docker.com/_/golang
# docker run --rm -v "$PWD":/usr/src/myapp -v "$HOME/dockergo":/go -w /usr/src/myapp golang make docker
# docker run --rm -it -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang bash
# 静态连接 glibc
docker:
	docker run --rm -v "$$PWD":/usr/src/myapp -v "$$HOME/dockergo":/go -w /usr/src/myapp golang make dockerinstall
	upx ~/dockergo/bin/$(APPNAME)
	mv ~/dockergo/bin/$(APPNAME)  ~/dockergo/bin/$(APPNAME)-amd64-glibc2.28
	gzip ~/dockergo/bin/$(APPNAME)-amd64-glibc2.28

dockerinstall:
	go install -v -x -a -ldflags '-extldflags "-static" -s -w' ./...