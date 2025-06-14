.PHONY: build clean install test help build-all release

# 版本信息
VERSION ?= dev
BINARY_NAME = openCursor
LDFLAGS = -ldflags="-s -w -X main.Version=$(VERSION)"

# 默认目标
help:
	@echo "Available targets:"
	@echo "  build      - 构建二进制文件"
	@echo "  build-all  - 构建所有平台的二进制文件"
	@echo "  clean      - 清理构建文件"
	@echo "  install    - 安装到 /usr/local/bin"
	@echo "  test       - 运行测试"
	@echo "  release    - 创建发布包"
	@echo "  help       - 显示此帮助信息"

# 构建当前平台的二进制文件
build:
	@echo "构建 $(BINARY_NAME)..."
	@go build $(LDFLAGS) -o $(BINARY_NAME) .

# 构建所有平台的二进制文件
build-all:
	@echo "构建所有平台的二进制文件..."
	@mkdir -p dist
	
	# macOS Intel
	@echo "构建 macOS Intel..."
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .
	
	# macOS Apple Silicon
	@echo "构建 macOS Apple Silicon..."
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .
	
	# Linux x86_64
	@echo "构建 Linux x86_64..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .
	
	# Linux ARM64
	@echo "构建 Linux ARM64..."
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 .
	
	# Windows x86_64
	@echo "构建 Windows x86_64..."
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .
	
	@echo "所有二进制文件已生成到 dist/ 目录"

# 清理构建文件
clean:
	@echo "清理构建文件..."
	@rm -f $(BINARY_NAME)
	@rm -rf dist/

# 安装到系统目录
install: build
	@echo "安装 $(BINARY_NAME) 到 /usr/local/bin..."
	@sudo cp $(BINARY_NAME) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "安装完成！"

# 运行测试
test:
	@echo "运行测试..."
	@go test ./...

# 创建发布包
release: build-all
	@echo "创建发布包..."
	@cd dist && \
	for file in *; do \
		if [ -f "$$file" ]; then \
			echo "创建 $$file.tar.gz"; \
			tar -czf "$$file.tar.gz" "$$file"; \
		fi \
	done
	@echo "发布包已创建到 dist/ 目录"

# 运行示例（需要设置API密钥）
example: build
	@echo "运行示例（请确保设置了 OPENAI_API_KEY 环境变量）"
	@./$(BINARY_NAME) "你好，请介绍一下你自己"

# 显示版本信息
version: build
	@./$(BINARY_NAME) version 