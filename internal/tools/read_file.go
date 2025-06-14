package tools

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadFileParams read_file工具的参数
type ReadFileParams struct {
	TargetFile                   string `json:"target_file"`
	ShouldReadEntireFile         bool   `json:"should_read_entire_file"`
	StartLineOneIndexed          int    `json:"start_line_one_indexed"`
	EndLineOneIndexedInclusive   int    `json:"end_line_one_indexed_inclusive"`
	Explanation                  string `json:"explanation,omitempty"`
}

// ReadFileResult read_file工具的返回结果
type ReadFileResult struct {
	Content           string `json:"content"`
	TotalLines        int    `json:"total_lines"`
	StartLine         int    `json:"start_line,omitempty"`
	EndLine           int    `json:"end_line,omitempty"`
	FilePath          string `json:"file_path"`
	LinesNotShown     string `json:"lines_not_shown,omitempty"`
	ReadEntireFile    bool   `json:"read_entire_file"`
}

// readFileFunction 读取文件工具函数
func readFileFunction(params map[string]interface{}) (interface{}, error) {
	// 解析参数
	targetFile, ok := params["target_file"].(string)
	if !ok || targetFile == "" {
		return nil, fmt.Errorf("target_file is required")
	}

	shouldReadEntireFile, _ := params["should_read_entire_file"].(bool)
	
	// 处理startLine参数，支持多种数值类型
	var startLine int
	if val, ok := params["start_line_one_indexed"]; ok {
		switch v := val.(type) {
		case float64:
			startLine = int(v)
		case int:
			startLine = v
		case int64:
			startLine = int(v)
		}
	}
	
	// 处理endLine参数，支持多种数值类型
	var endLine int
	if val, ok := params["end_line_one_indexed_inclusive"]; ok {
		switch v := val.(type) {
		case float64:
			endLine = int(v)
		case int:
			endLine = v
		case int64:
			endLine = int(v)
		}
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

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 读取所有行
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	totalLines := len(lines)

	result := &ReadFileResult{
		FilePath:       filePath,
		TotalLines:     totalLines,
		ReadEntireFile: shouldReadEntireFile,
	}

	if shouldReadEntireFile {
		// 读取整个文件
		result.Content = strings.Join(lines, "\n")
		result.StartLine = 1
		result.EndLine = totalLines
	} else {
		// 读取指定行范围
		startLineInt := startLine
		endLineInt := endLine

		// 验证行号范围
		if startLineInt < 1 {
			startLineInt = 1
		}
		if endLineInt < startLineInt {
			return nil, fmt.Errorf("end_line (%d) must be >= start_line (%d)", endLineInt, startLineInt)
		}
		if startLineInt > totalLines {
			return nil, fmt.Errorf("start_line (%d) exceeds total lines (%d)", startLineInt, totalLines)
		}

		// 调整结束行号
		if endLineInt > totalLines {
			endLineInt = totalLines
		}

		// 验证行数限制（最多250行，最少200行）
		lineCount := endLineInt - startLineInt + 1
		if lineCount > 250 {
			return nil, fmt.Errorf("cannot read more than 250 lines at once (requested: %d)", lineCount)
		}
		if lineCount < 200 && totalLines >= 200 && endLineInt < totalLines {
			// 如果请求的行数少于200行且文件总行数>=200，建议读取更多行
			suggestedEnd := startLineInt + 199
			if suggestedEnd > totalLines {
				suggestedEnd = totalLines
			}
			return nil, fmt.Errorf("minimum 200 lines required when file has >= 200 lines. Consider reading lines %d-%d", startLineInt, suggestedEnd)
		}

		// 提取指定行范围 (转换为0-based索引)
		startIdx := startLineInt - 1
		endIdx := endLineInt - 1

		selectedLines := lines[startIdx : endIdx+1]
		result.Content = strings.Join(selectedLines, "\n")
		result.StartLine = startLineInt
		result.EndLine = endLineInt

		// 生成未显示行数的摘要
		var notShownParts []string
		if startLineInt > 1 {
			notShownParts = append(notShownParts, fmt.Sprintf("Lines 1-%d not shown", startLineInt-1))
		}
		if endLineInt < totalLines {
			notShownParts = append(notShownParts, fmt.Sprintf("Lines %d-%d not shown", endLineInt+1, totalLines))
		}
		if len(notShownParts) > 0 {
			result.LinesNotShown = strings.Join(notShownParts, "; ")
		}
	}

	return result, nil
}

// NewReadFileTool 创建read_file工具
func NewReadFileTool() Tool {
	schema := ToolSchema{
		Name: "read_file",
		Description: "Read the contents of a file. The output of this tool call will be the 1-indexed file contents from start_line_one_indexed to end_line_one_indexed_inclusive, together with a summary of the lines outside start_line_one_indexed and end_line_one_indexed_inclusive.\nNote that this call can view at most 250 lines at a time and 200 lines minimum.\n\nWhen using this tool to gather information, it's your responsibility to ensure you have the COMPLETE context. Specifically, each time you call this command you should:\n1) Assess if the contents you viewed are sufficient to proceed with your task.\n2) Take note of where there are lines not shown.\n3) If the file contents you have viewed are insufficient, and you suspect they may be in lines not shown, proactively call the tool again to view those lines.\n4) When in doubt, call this tool again to gather more information. Remember that partial file views may miss critical dependencies, imports, or functionality.\n\nIn some cases, if reading a range of lines is not enough, you may choose to read the entire file.\nReading entire files is often wasteful and slow, especially for large files (i.e. more than a few hundred lines). So you should use this option sparingly.\nReading the entire file is not allowed in most cases. You are only allowed to read the entire file if it has been edited or manually attached to the conversation by the user.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"target_file": map[string]interface{}{
					"type":        "string",
					"description": "The path of the file to read. You can use either a relative path in the workspace or an absolute path. If an absolute path is provided, it will be preserved as is.",
				},
				"should_read_entire_file": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether to read the entire file. Defaults to false.",
				},
				"start_line_one_indexed": map[string]interface{}{
					"type":        "integer",
					"description": "The one-indexed line number to start reading from (inclusive).",
				},
				"end_line_one_indexed_inclusive": map[string]interface{}{
					"type":        "integer",
					"description": "The one-indexed line number to end reading at (inclusive).",
				},
				"explanation": map[string]interface{}{
					"type":        "string",
					"description": "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
				},
			},
			"required": []string{
				"target_file",
				"should_read_entire_file",
				"start_line_one_indexed",
				"end_line_one_indexed_inclusive",
			},
		},
	}

	return Tool{
		Schema:   schema,
		Function: readFileFunction,
	}
} 