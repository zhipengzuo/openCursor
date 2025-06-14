package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteFileParams write_file工具的参数
type WriteFileParams struct {
	TargetFile  string `json:"target_file"`
	Content     string `json:"content"`
	Overwrite   bool   `json:"overwrite,omitempty"`
	Explanation string `json:"explanation,omitempty"`
}

// WriteFileResult write_file工具的返回结果
type WriteFileResult struct {
	TargetFile   string `json:"target_file"`
	Written      bool   `json:"written"`
	Created      bool   `json:"created"`
	BytesWritten int    `json:"bytes_written"`
	Message      string `json:"message"`
	FileExists   bool   `json:"file_exists"`
}

// writeFileFunction 写入文件工具函数
func writeFileFunction(params map[string]interface{}) (interface{}, error) {
	// 解析参数
	targetFile, ok := params["target_file"].(string)
	if !ok || targetFile == "" {
		return nil, fmt.Errorf("target_file is required")
	}

	content, ok := params["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content is required")
	}

	overwrite, _ := params["overwrite"].(bool)
	workDir, _ := params["__work_dir__"].(string)

	// 解析文件路径
	var filePath string
	if filepath.IsAbs(targetFile) {
		filePath = targetFile
	} else {
		if workDir != "" {
			filePath = filepath.Join(workDir, targetFile)
		} else {
			filePath = targetFile
		}
	}

	result := &WriteFileResult{
		TargetFile: filePath,
		Written:    false,
		Created:    false,
	}

	// 检查文件是否存在
	fileExists := false
	if _, err := os.Stat(filePath); err == nil {
		fileExists = true
		result.FileExists = true
	}

	// 如果文件存在且不允许覆盖
	if fileExists && !overwrite {
		result.Message = "File exists and overwrite is not enabled"
		return result, nil
	}

	// 执行安全检查
	if err := performWriteSecurityChecks(filePath); err != nil {
		result.Message = fmt.Sprintf("Security check failed: %v", err)
		return result, nil
	}

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		result.Message = fmt.Sprintf("Failed to create directory: %v", err)
		return result, nil
	}

	// 写入文件
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to write file: %v", err)
		return result, nil
	}

	result.Written = true
	result.Created = !fileExists
	result.BytesWritten = len(content)

	if result.Created {
		result.Message = fmt.Sprintf("File created successfully with %d bytes", result.BytesWritten)
	} else {
		result.Message = fmt.Sprintf("File overwritten successfully with %d bytes", result.BytesWritten)
	}

	return result, nil
}

// performWriteSecurityChecks 执行写入安全检查
func performWriteSecurityChecks(filePath string) error {
	// 检查是否为系统重要目录
	dangerousPaths := []string{
		"/etc",
		"/bin",
		"/sbin",
		"/usr/bin",
		"/usr/sbin",
		"/boot",
		"/sys",
		"/proc",
		"/dev",
		"C:\\Windows",
		"C:\\Program Files",
		"C:\\Program Files (x86)",
		"C:\\System32",
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	for _, dangerousPath := range dangerousPaths {
		if strings.HasPrefix(absPath, dangerousPath) {
			return fmt.Errorf("cannot write files to system directory: %s", dangerousPath)
		}
	}

	// 检查文件扩展名
	dangerousExtensions := []string{
		".exe", ".dll", ".sys", ".bat", ".cmd", ".com", ".scr",
		".pif", ".application", ".gadget", ".msi", ".msp", ".msc",
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	for _, dangerousExt := range dangerousExtensions {
		if ext == dangerousExt {
			return fmt.Errorf("cannot create potentially dangerous file type: %s", ext)
		}
	}

	// 检查是否为隐藏的系统文件
	fileName := filepath.Base(filePath)
	systemFiles := []string{
		"boot.ini", "ntldr", "bootmgr", "pagefile.sys", "hiberfil.sys",
		"autoexec.bat", "config.sys",
	}

	for _, systemFile := range systemFiles {
		if strings.ToLower(fileName) == systemFile {
			return fmt.Errorf("cannot create system file: %s", systemFile)
		}
	}

	// 检查文件名是否包含危险字符
	dangerousChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range dangerousChars {
		if strings.Contains(fileName, char) {
			return fmt.Errorf("filename contains dangerous character: %s", char)
		}
	}

	return nil
}

// NewWriteFileTool 创建write_file工具
func NewWriteFileTool() Tool {
	schema := ToolSchema{
		Name:        "write_file",
		Description: "Write content to a file. If the file doesn't exist, it will be created. If the file exists, it will be overwritten only if the overwrite parameter is set to true. The tool includes safety checks to prevent writing to system directories or creating dangerous file types.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"target_file": map[string]interface{}{
					"type":        "string",
					"description": "The path of the file to write to. You can use either a relative path in the workspace or an absolute path. If an absolute path is provided, it will be preserved as is.",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "The content to write to the file.",
				},
				"overwrite": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether to overwrite the file if it already exists. Defaults to false.",
				},
				"explanation": map[string]interface{}{
					"type":        "string",
					"description": "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
				},
			},
			"required": []string{"target_file", "content"},
		},
	}

	return Tool{
		Schema:   schema,
		Function: writeFileFunction,
	}
} 