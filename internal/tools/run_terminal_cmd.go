package tools

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// RunTerminalCmdParams run_terminal_cmd工具的参数
type RunTerminalCmdParams struct {
	Command      string `json:"command"`
	IsBackground bool   `json:"is_background"`
	Explanation  string `json:"explanation,omitempty"`
}

// RunTerminalCmdResult run_terminal_cmd工具的返回结果
type RunTerminalCmdResult struct {
	Command      string `json:"command"`
	Output       string `json:"output"`
	Error        string `json:"error,omitempty"`
	ExitCode     int    `json:"exit_code"`
	IsBackground bool   `json:"is_background"`
	PID          int    `json:"pid,omitempty"`
}

// runTerminalCmdFunction 运行终端命令工具函数
func runTerminalCmdFunction(params map[string]interface{}) (interface{}, error) {
	// 解析参数
	command, ok := params["command"].(string)
	if !ok || command == "" {
		return nil, fmt.Errorf("command is required")
	}

	isBackground, _ := params["is_background"].(bool)
	workDir, _ := params["__work_dir__"].(string)

	// 清理命令（移除换行符）
	command = strings.ReplaceAll(command, "\n", " ")
	command = strings.TrimSpace(command)

	result := &RunTerminalCmdResult{
		Command:      command,
		IsBackground: isBackground,
	}

	// 根据操作系统选择shell
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	// 设置工作目录
	if workDir != "" {
		if filepath.IsAbs(workDir) {
			cmd.Dir = workDir
		} else {
			absDir, err := filepath.Abs(workDir)
			if err == nil {
				cmd.Dir = absDir
			}
		}
	}

	if isBackground {
		// 后台运行
		err := cmd.Start()
		if err != nil {
			result.Error = err.Error()
			result.ExitCode = -1
			return result, nil
		}
		
		result.PID = cmd.Process.Pid
		result.Output = fmt.Sprintf("Command started in background with PID: %d", cmd.Process.Pid)
		result.ExitCode = 0
		
		// 启动一个goroutine来等待命令完成
		go func() {
			cmd.Wait()
		}()
	} else {
		// 前台运行
		output, err := cmd.CombinedOutput()
		result.Output = string(output)
		
		if err != nil {
			result.Error = err.Error()
			if exitError, ok := err.(*exec.ExitError); ok {
				result.ExitCode = exitError.ExitCode()
			} else {
				result.ExitCode = -1
			}
		} else {
			result.ExitCode = 0
		}
	}

	return result, nil
}

// NewRunTerminalCmdTool 创建run_terminal_cmd工具
func NewRunTerminalCmdTool() Tool {
	schema := ToolSchema{
		Name:        "run_terminal_cmd",
		Description: "PROPOSE a command to run on behalf of the user.\nIf you have this tool, note that you DO have the ability to run commands directly on the USER's system.\nNote that the user will have to approve the command before it is executed.\nThe user may reject it if it is not to their liking, or may modify the command before approving it.  If they do change it, take those changes into account.\nThe actual command will NOT execute until the user approves it. The user may not approve it immediately. Do NOT assume the command has started running.\nIf the step is WAITING for user approval, it has NOT started running.\nIn using these tools, adhere to the following guidelines:\n1. Based on the contents of the conversation, you will be told if you are in the same shell as a previous step or a different shell.\n2. If in a new shell, you should `cd` to the appropriate directory and do necessary setup in addition to running the command.\n3. If in the same shell, LOOK IN CHAT HISTORY for your current working directory.\n4. For ANY commands that would require user interaction, ASSUME THE USER IS NOT AVAILABLE TO INTERACT and PASS THE NON-INTERACTIVE FLAGS (e.g. --yes for npx).\n5. If the command would use a pager, append ` | cat` to the command.\n6. For commands that are long running/expected to run indefinitely until interruption, please run them in the background. To run jobs in the background, set `is_background` to true rather than changing the details of the command.\n7. Dont include any newlines in the command.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "The terminal command to execute",
				},
				"is_background": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether the command should be run in the background",
				},
				"explanation": map[string]interface{}{
					"type":        "string",
					"description": "One sentence explanation as to why this command needs to be run and how it contributes to the goal.",
				},
			},
			"required": []string{"command", "is_background"},
		},
	}

	return Tool{
		Schema:   schema,
		Function: runTerminalCmdFunction,
	}
} 