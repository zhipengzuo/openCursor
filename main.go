package main

import (
	"openCursor/cmd"
	"os"
)

// Version 版本信息，构建时通过 ldflags 注入
var Version = "dev"

func main() {
	// 设置版本信息到cmd包
	cmd.SetVersion(Version)
	
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
} 