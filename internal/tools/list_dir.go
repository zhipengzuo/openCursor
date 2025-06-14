package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ListDirParams list_dir工具的参数
type ListDirParams struct {
	RelativeWorkspacePath string `json:"relative_workspace_path"`
	Explanation           string `json:"explanation,omitempty"`
}

// FileInfo 文件信息
type FileInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"` // "file" or "dir"
	Size     int64  `json:"size,omitempty"`
	SizeStr  string `json:"size_str,omitempty"`
	ItemCount string `json:"item_count,omitempty"`
}

// ListDirResult list_dir工具的返回结果
type ListDirResult struct {
	Path  string     `json:"path"`
	Items []FileInfo `json:"items"`
	Count int        `json:"count"`
}

// formatSize 格式化文件大小
func formatSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%dB", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1fKB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.1fMB", float64(size)/(1024*1024))
	} else {
		return fmt.Sprintf("%.1fGB", float64(size)/(1024*1024*1024))
	}
}

// countDirItems 计算目录中的项目数量
func countDirItems(dirPath string) string {
	items, err := os.ReadDir(dirPath)
	if err != nil {
		return "? items"
	}
	count := len(items)
	if count == 1 {
		return "1 item"
	}
	return fmt.Sprintf("%d items", count)
}

// listDirFunction 列出目录内容工具函数
func listDirFunction(params map[string]interface{}) (interface{}, error) {
	// 解析参数
	relativePath, ok := params["relative_workspace_path"].(string)
	if !ok {
		return nil, fmt.Errorf("relative_workspace_path is required")
	}

	workDir, _ := params["__work_dir__"].(string)

	// 构建绝对路径
	var targetPath string
	if filepath.IsAbs(relativePath) {
		targetPath = relativePath
	} else {
		if workDir != "" {
			targetPath = filepath.Join(workDir, relativePath)
		} else {
			targetPath = relativePath
		}
	}

	// 检查目录是否存在
	info, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory not found: %s", targetPath)
		}
		return nil, fmt.Errorf("failed to access directory: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", targetPath)
	}

	// 读取目录内容
	entries, err := os.ReadDir(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// 构建结果
	var items []FileInfo
	for _, entry := range entries {
		fileInfo := FileInfo{
			Name: entry.Name(),
		}

		if entry.IsDir() {
			fileInfo.Type = "dir"
			// 计算子目录项目数量
			subDirPath := filepath.Join(targetPath, entry.Name())
			fileInfo.ItemCount = countDirItems(subDirPath)
		} else {
			fileInfo.Type = "file"
			// 获取文件大小
			if info, err := entry.Info(); err == nil {
				fileInfo.Size = info.Size()
				fileInfo.SizeStr = formatSize(info.Size())
			}
		}

		items = append(items, fileInfo)
	}

	// 排序：目录在前，文件在后，各自按名称排序
	sort.Slice(items, func(i, j int) bool {
		if items[i].Type != items[j].Type {
			return items[i].Type == "dir" // 目录排在前面
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	result := &ListDirResult{
		Path:  targetPath,
		Items: items,
		Count: len(items),
	}

	return result, nil
}

// NewListDirTool 创建list_dir工具
func NewListDirTool() Tool {
	schema := ToolSchema{
		Name:        "list_dir",
		Description: "List the contents of a directory. The quick tool to use for discovery, before using more targeted tools like semantic search or file reading. Useful to try to understand the file structure before diving deeper into specific files. Can be used to explore the codebase.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"relative_workspace_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to list contents of, relative to the workspace root.",
				},
				"explanation": map[string]interface{}{
					"type":        "string",
					"description": "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
				},
			},
			"required": []string{"relative_workspace_path"},
		},
	}

	return Tool{
		Schema:   schema,
		Function: listDirFunction,
	}
} 