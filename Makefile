.PHONY: all build run tool clean help

GREEN_PREFIX="\033[32m"
COLOR_SUFFIX="\033[0m"
SKY_BLUE_PREFIX="\033[36m"
BINARY="ace"

all: tool build

build:
	go mod tidy
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o ${BINARY} .

run:
	@go run ./

tool:
	go vet ./...
	go fmt ./...
	goimports -w .
	@echo ${GREEN_PREFIX} "go tool ok" ${COLOR_SUFFIX}

clean:
	@if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

help:
	@echo ${SKY_BLUE_PREFIX} "make - 格式化 Go 代码, 并编译生成二进制文件" ${COLOR_SUFFIX}
	@echo ${SKY_BLUE_PREFIX} "make build - 编译 Go 代码, 生成二进制文件" ${COLOR_SUFFIX}
	@echo ${SKY_BLUE_PREFIX} "make run - 直接运行 Go 代码" ${COLOR_SUFFIX}
	@echo ${SKY_BLUE_PREFIX} "make clean - 移除二进制文件和 vim swap files" ${COLOR_SUFFIX}
	@echo ${SKY_BLUE_PREFIX} "make tool - 运行 Go 工具 'fmt' and 'vet'" ${COLOR_SUFFIX}
