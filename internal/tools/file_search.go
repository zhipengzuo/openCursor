package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileSearchParams file_search工具的参数
type FileSearchParams struct {
	Query       string `json:"query"`
	Explanation string `json:"explanation,omitempty"`
}

// FileMatch 文件匹配结果
type FileMatch struct {
	Path  string  `json:"path"`
	Score float64 `json:"score"`
	Match string  `json:"match"`
}

// FileSearchResult file_search工具的返回结果
type FileSearchResult struct {
	Query   string      `json:"query"`
	Matches []FileMatch `json:"matches"`
	Count   int         `json:"count"`
}

// calculateFuzzyScore 计算模糊匹配分数
func calculateFuzzyScore(query, path string) float64 {
	query = strings.ToLower(query)
	path = strings.ToLower(path)
	
	// 基础文件名匹配
	fileName := strings.ToLower(filepath.Base(path))
	
	// 完全匹配得分最高
	if fileName == query {
		return 100.0
	}
	
	// 包含查询字符串
	if strings.Contains(fileName, query) {
		return 80.0 + float64(len(query))/float64(len(fileName))*20.0
	}
	
	// 路径包含查询字符串
	if strings.Contains(path, query) {
		return 60.0 + float64(len(query))/float64(len(path))*20.0
	}
	
	// 计算字符匹配度
	score := 0.0
	queryChars := []rune(query)
	pathChars := []rune(fileName)
	
	// 简单的字符匹配算法
	queryIdx := 0
	for i, char := range pathChars {
		if queryIdx < len(queryChars) && char == queryChars[queryIdx] {
			score += 1.0
			queryIdx++
			
			// 连续匹配奖励
			if queryIdx < len(queryChars) && i+1 < len(pathChars) && 
			   pathChars[i+1] == queryChars[queryIdx] {
				score += 0.5
			}
		}
	}
	
	// 计算匹配比例
	if len(queryChars) > 0 {
		score = (score / float64(len(queryChars))) * 50.0
	}
	
	return score
}

// fileSearchFunction 文件搜索工具函数
func fileSearchFunction(params map[string]interface{}) (interface{}, error) {
	// 解析参数
	query, ok := params["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query is required")
	}

	workDir, _ := params["__work_dir__"].(string)
	searchPath := "."
	if workDir != "" {
		searchPath = workDir
	}

	result := &FileSearchResult{
		Query:   query,
		Matches: []FileMatch{},
	}

	var allFiles []string
	
	// 遍历目录收集所有文件
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续处理其他文件
		}
		
		// 跳过目录和隐藏文件
		if info.IsDir() || strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}
		
		// 跳过一些常见的不需要搜索的文件类型
		ext := strings.ToLower(filepath.Ext(path))
		skipExtensions := []string{".exe", ".dll", ".so", ".dylib", ".o", ".a", 
			".jar", ".war", ".zip", ".tar", ".gz", ".7z", ".rar",
			".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg", ".ico",
			".mp3", ".mp4", ".avi", ".mov", ".wav", ".pdf"}
		
		for _, skipExt := range skipExtensions {
			if ext == skipExt {
				return nil
			}
		}
		
		allFiles = append(allFiles, path)
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	// 计算匹配分数并过滤
	var matches []FileMatch
	for _, file := range allFiles {
		score := calculateFuzzyScore(query, file)
		if score > 0 {
			// 生成匹配描述
			match := generateMatchDescription(query, file)
			
			matches = append(matches, FileMatch{
				Path:  file,
				Score: score,
				Match: match,
			})
		}
	}

	// 按分数排序
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	// 限制结果数量为10个
	if len(matches) > 10 {
		matches = matches[:10]
	}

	result.Matches = matches
	result.Count = len(matches)

	return result, nil
}

// generateMatchDescription 生成匹配描述
func generateMatchDescription(query, path string) string {
	fileName := filepath.Base(path)
	dir := filepath.Dir(path)
	
	query = strings.ToLower(query)
	fileName = strings.ToLower(fileName)
	
	if strings.Contains(fileName, query) {
		return fmt.Sprintf("Filename contains '%s'", query)
	}
	
	if strings.Contains(strings.ToLower(dir), query) {
		return fmt.Sprintf("Directory path contains '%s'", query)
	}
	
	return "Fuzzy match"
}

// NewFileSearchTool 创建file_search工具
func NewFileSearchTool() Tool {
	schema := ToolSchema{
		Name:        "file_search",
		Description: "Fast file search based on fuzzy matching against file path. Use if you know part of the file path but don't know where it's located exactly. Response will be capped to 10 results. Make your query more specific if need to filter results further.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Fuzzy filename to search for",
				},
				"explanation": map[string]interface{}{
					"type":        "string",
					"description": "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
				},
			},
			"required": []string{"query", "explanation"},
		},
	}

	return Tool{
		Schema:   schema,
		Function: fileSearchFunction,
	}
} 