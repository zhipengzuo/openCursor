package tools

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SearchReplaceParams search_replace工具的参数
type SearchReplaceParams struct {
	FilePath  string `json:"file_path"`
	OldString string `json:"old_string"`
	NewString string `json:"new_string"`
}

// SearchReplaceResult search_replace工具的返回结果
type SearchReplaceResult struct {
	FilePath     string `json:"file_path"`
	OldString    string `json:"old_string"`
	NewString    string `json:"new_string"`
	Replaced     bool   `json:"replaced"`
	LineNumber   int    `json:"line_number,omitempty"`
	OriginalLine string `json:"original_line,omitempty"`
	NewLine      string `json:"new_line,omitempty"`
	Message      string `json:"message"`
}

// searchReplaceFunction 搜索替换工具函数
func searchReplaceFunction(params map[string]interface{}) (interface{}, error) {
	// 解析参数
	filePath, ok := params["file_path"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("file_path is required")
	}

	oldString, ok := params["old_string"].(string)
	if !ok || oldString == "" {
		return nil, fmt.Errorf("old_string is required")
	}

	newString, ok := params["new_string"].(string)
	if !ok {
		return nil, fmt.Errorf("new_string is required")
	}

	workDir, _ := params["__work_dir__"].(string)

	// 解析文件路径
	var targetPath string
	if filepath.IsAbs(filePath) {
		targetPath = filePath
	} else {
		if workDir != "" {
			targetPath = filepath.Join(workDir, filePath)
		} else {
			targetPath = filePath
		}
	}

	result := &SearchReplaceResult{
		FilePath:  targetPath,
		OldString: oldString,
		NewString: newString,
		Replaced:  false,
	}

	// 检查文件是否存在
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		result.Message = fmt.Sprintf("File not found: %s", targetPath)
		return result, nil
	}

	// 读取文件内容
	file, err := os.Open(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var lines []string
	var foundLine int = -1
	var originalLine string

	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		lines = append(lines, line)

		// 查找第一个匹配的行
		if foundLine == -1 && strings.Contains(line, oldString) {
			foundLine = lineNumber
			originalLine = line
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 如果没有找到匹配项
	if foundLine == -1 {
		result.Message = "Old string not found in file"
		return result, nil
	}

	// 执行替换（只替换第一个匹配项）
	lines[foundLine-1] = strings.Replace(lines[foundLine-1], oldString, newString, 1)
	newLine := lines[foundLine-1]

	// 写回文件
	err = writeLinesToFile(targetPath, lines)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	result.Replaced = true
	result.LineNumber = foundLine
	result.OriginalLine = originalLine
	result.NewLine = newLine
	result.Message = fmt.Sprintf("Successfully replaced text on line %d", foundLine)

	return result, nil
}

// writeLinesToFile 将行写入文件
func writeLinesToFile(filePath string, lines []string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for i, line := range lines {
		if i > 0 {
			writer.WriteString("\n")
		}
		writer.WriteString(line)
	}

	// 确保文件以换行符结尾（如果原文件有的话）
	if len(lines) > 0 {
		writer.WriteString("\n")
	}

	return nil
}

// NewSearchReplaceTool 创建search_replace工具
func NewSearchReplaceTool() Tool {
	schema := ToolSchema{
		Name:        "search_replace",
		Description: "Use this tool to propose a search and replace operation on an existing file.\n\nThe tool will replace ONE occurrence of old_string with new_string in the specified file.\n\nCRITICAL REQUIREMENTS FOR USING THIS TOOL:\n\n1. UNIQUENESS: The old_string MUST uniquely identify the specific instance you want to change. This means:\n   - Include AT LEAST 3-5 lines of context BEFORE the change point\n   - Include AT LEAST 3-5 lines of context AFTER the change point\n   - Include all whitespace, indentation, and surrounding code exactly as it appears in the file\n\n2. SINGLE INSTANCE: This tool can only change ONE instance at a time. If you need to change multiple instances:\n   - Make separate calls to this tool for each instance\n   - Each call must uniquely identify its specific instance using extensive context\n\n3. VERIFICATION: Before using this tool:\n   - If multiple instances exist, gather enough context to uniquely identify each one\n   - Plan separate tool calls for each instance",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "The path to the file you want to search and replace in. You can use either a relative path in the workspace or an absolute path. If an absolute path is provided, it will be preserved as is.",
				},
				"old_string": map[string]interface{}{
					"type":        "string",
					"description": "The text to replace (must be unique within the file, and must match the file contents exactly, including all whitespace and indentation)",
				},
				"new_string": map[string]interface{}{
					"type":        "string",
					"description": "The edited text to replace the old_string (must be different from the old_string)",
				},
			},
			"required": []string{"file_path", "old_string", "new_string"},
		},
	}

	return Tool{
		Schema:   schema,
		Function: searchReplaceFunction,
	}
} 