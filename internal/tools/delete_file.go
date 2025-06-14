package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DeleteFileParams delete_file工具的参数
type DeleteFileParams struct {
	TargetFile  string `json:"target_file"`
	Explanation string `json:"explanation,omitempty"`
}

// DeleteFileResult delete_file工具的返回结果
type DeleteFileResult struct {
	TargetFile string `json:"target_file"`
	Deleted    bool   `json:"deleted"`
	Message    string `json:"message"`
	FileInfo   string `json:"file_info,omitempty"`
}

// deleteFileFunction 删除文件工具函数
func deleteFileFunction(params map[string]interface{}) (interface{}, error) {
	// 解析参数
	targetFile, ok := params["target_file"].(string)
	if !ok || targetFile == "" {
		return nil, fmt.Errorf("target_file is required")
	}

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

	result := &DeleteFileResult{
		TargetFile: filePath,
		Deleted:    false,
	}

	// 检查文件是否存在
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		result.Message = "File does not exist"
		return result, nil
	}

	if err != nil {
		result.Message = fmt.Sprintf("Failed to access file: %v", err)
		return result, nil
	}

	// 记录文件信息
	if info.IsDir() {
		result.FileInfo = fmt.Sprintf("Directory with %d bytes", info.Size())
		result.Message = "Cannot delete directories with this tool"
		return result, nil
	} else {
		result.FileInfo = fmt.Sprintf("File with %d bytes", info.Size())
	}

	// 执行安全检查
	if err := performSecurityChecks(filePath); err != nil {
		result.Message = fmt.Sprintf("Security check failed: %v", err)
		return result, nil
	}

	// 尝试删除文件
	err = os.Remove(filePath)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to delete file: %v", err)
		return result, nil
	}

	result.Deleted = true
	result.Message = "File successfully deleted"

	return result, nil
}

// performSecurityChecks 执行安全检查
func performSecurityChecks(filePath string) error {
	// 检查是否为系统重要文件
	dangerousPaths := []string{
		"/etc",
		"/bin",
		"/sbin",
		"/usr/bin",
		"/usr/sbin",
		"/boot",
		"/sys",
		"/proc",
		"C:\\Windows",
		"C:\\Program Files",
		"C:\\Program Files (x86)",
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	for _, dangerousPath := range dangerousPaths {
		if strings.HasPrefix(absPath, dangerousPath) {
			return fmt.Errorf("cannot delete files in system directory: %s", dangerousPath)
		}
	}

	// 检查文件扩展名
	dangerousExtensions := []string{
		".exe", ".dll", ".sys", ".bat", ".cmd", ".com", ".scr",
		".pif", ".application", ".gadget", ".msi", ".msp", ".msc",
	}

	ext := filepath.Ext(filePath)
	for _, dangerousExt := range dangerousExtensions {
		if ext == dangerousExt {
			return fmt.Errorf("cannot delete potentially dangerous file type: %s", ext)
		}
	}

	// 检查是否为隐藏的系统文件
	fileName := filepath.Base(filePath)
	systemFiles := []string{
		"boot.ini", "ntldr", "bootmgr", "pagefile.sys", "hiberfil.sys",
		".DS_Store", "Thumbs.db", "desktop.ini",
	}

	for _, systemFile := range systemFiles {
		if fileName == systemFile {
			return fmt.Errorf("cannot delete system file: %s", systemFile)
		}
	}

	return nil
}

// NewDeleteFileTool 创建delete_file工具
func NewDeleteFileTool() Tool {
	schema := ToolSchema{
		Name:        "delete_file",
		Description: "Deletes a file at the specified path. The operation will fail gracefully if:\n    - The file doesn't exist\n    - The operation is rejected for security reasons\n    - The file cannot be deleted",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"target_file": map[string]interface{}{
					"type":        "string",
					"description": "The path of the file to delete, relative to the workspace root.",
				},
				"explanation": map[string]interface{}{
					"type":        "string",
					"description": "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
				},
			},
			"required": []string{"target_file"},
		},
	}

	return Tool{
		Schema:   schema,
		Function: deleteFileFunction,
	}
} 