package cmd

import (
	"openCursor/internal/client"
	"openCursor/internal/tools"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// 版本信息
var version = "dev"

// SetVersion 设置版本号
func SetVersion(v string) {
	version = v
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "openCursor [query]",
	Short: "A CLI tool to interact with DeepSeek API",
	Long: `openCursor is a command line tool that allows you to interact with DeepSeek AI models.
You can send queries and receive streaming responses with tool calling support.

Environment Variables:
  OPENAI_API_KEY    API key for authentication (required)
  MODEL             Model name to use (default: "deepseek-chat")  
  BASE_URL          API base URL (default: "https://api.deepseek.com/v1")

Examples:
  export OPENAI_API_KEY="your-api-key"
  export MODEL="deepseek-chat"
  export BASE_URL="https://api.deepseek.com/v1"
  
  openCursor "Hello, how are you?"
  openCursor "Please help me write a Python function"
  openCursor "List files in current directory"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		
		// 获取环境变量
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			fmt.Fprintf(os.Stderr, "Error: OPENAI_API_KEY environment variable is required.\n")
			os.Exit(1)
		}
		
		model := os.Getenv("MODEL")
		if model == "" {
			model = "deepseek-chat" // 默认模型
		}
		
		baseURL := os.Getenv("BASE_URL")
		if baseURL == "" {
			baseURL = "https://api.deepseek.com/v1" // 默认URL
		}
		
		// 初始化工具管理器
		if err := tools.RegisterDefaultTools(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to register tools: %v\n", err)
			os.Exit(1)
		}
		
		// 设置工作目录为当前目录
		workDir, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to get current directory: %v\n", err)
			os.Exit(1)
		}
		tools.SetDefaultWorkDirectory(workDir)
		
		// 创建DeepSeek客户端
		aiClient := client.NewClient(apiKey, baseURL, model)
		aiClient.SetToolManager(tools.GetDefaultManager())
		
		// 发送查询并处理流式响应（支持工具调用）
		if err := aiClient.StreamQueryWithTools(query); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of openCursor",
	Long:  `Display the current version of openCursor.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("openCursor version %s\n", version)
	},
}

func init() {
	// 添加version子命令
	rootCmd.AddCommand(versionCmd)
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
} 