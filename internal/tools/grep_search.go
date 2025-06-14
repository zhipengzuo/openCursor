package tools

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// GrepSearchParams grep_search工具的参数
type GrepSearchParams struct {
	Query          string `json:"query"`
	CaseSensitive  bool   `json:"case_sensitive,omitempty"`
	IncludePattern string `json:"include_pattern,omitempty"`
	ExcludePattern string `json:"exclude_pattern,omitempty"`
	Explanation    string `json:"explanation,omitempty"`
}

// GrepMatch 匹配结果
type GrepMatch struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column,omitempty"`
	Content  string `json:"content"`
	Match    string `json:"match"`
}

// GrepSearchResult grep_search工具的返回结果
type GrepSearchResult struct {
	Query          string      `json:"query"`
	Matches        []GrepMatch `json:"matches"`
	TotalMatches   int         `json:"total_matches"`
	MatchedFiles   int         `json:"matched_files"`
	CaseSensitive  bool        `json:"case_sensitive"`
	IncludePattern string      `json:"include_pattern,omitempty"`
	ExcludePattern string      `json:"exclude_pattern,omitempty"`
}

// grepSearchFunction grep搜索工具函数
func grepSearchFunction(params map[string]interface{}) (interface{}, error) {
	// 解析参数
	query, ok := params["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query is required")
	}

	caseSensitive, _ := params["case_sensitive"].(bool)
	includePattern, _ := params["include_pattern"].(string)
	excludePattern, _ := params["exclude_pattern"].(string)
	workDir, _ := params["__work_dir__"].(string)

	// 检查ripgrep是否可用
	_, err := exec.LookPath("rg")
	if err != nil {
		// 如果ripgrep不可用，回退到内置实现
		return fallbackGrepSearch(query, caseSensitive, includePattern, excludePattern, workDir)
	}

	// 构建ripgrep命令
	args := []string{
		"--no-heading",
		"--line-number",
		"--column",
		"--color=never",
		"--max-count=50", // 限制最多50个匹配
	}

	// 大小写敏感选项
	if !caseSensitive {
		args = append(args, "--ignore-case")
	}

	// 包含模式
	if includePattern != "" {
		args = append(args, "--glob", includePattern)
	}

	// 排除模式
	if excludePattern != "" {
		args = append(args, "--glob", "!"+excludePattern)
	}

	// 添加查询模式
	args = append(args, query)

	// 添加搜索路径
	searchPath := "."
	if workDir != "" {
		searchPath = workDir
	}
	args = append(args, searchPath)

	// 执行ripgrep命令
	cmd := exec.Command("rg", args...)
	output, err := cmd.Output()

	result := &GrepSearchResult{
		Query:          query,
		CaseSensitive:  caseSensitive,
		IncludePattern: includePattern,
		ExcludePattern: excludePattern,
		Matches:        []GrepMatch{},
	}

	if err != nil {
		// ripgrep返回非零退出码可能只是表示没有找到匹配项
		if len(output) == 0 {
			return result, nil
		}
	}

	// 解析ripgrep输出
	lines := strings.Split(string(output), "\n")
	fileSet := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// ripgrep输出格式: file:line:column:content
		parts := strings.SplitN(line, ":", 4)
		if len(parts) < 4 {
			continue
		}

		file := parts[0]
		lineNum := 0
		columnNum := 0
		content := parts[3]

		// 解析行号
		if ln, err := parseIntSafe(parts[1]); err == nil {
			lineNum = ln
		}

		// 解析列号
		if cn, err := parseIntSafe(parts[2]); err == nil {
			columnNum = cn
		}

		// 提取匹配的部分
		match := extractMatch(content, query, caseSensitive)

		grepMatch := GrepMatch{
			File:    file,
			Line:    lineNum,
			Column:  columnNum,
			Content: content,
			Match:   match,
		}

		result.Matches = append(result.Matches, grepMatch)
		fileSet[file] = true
	}

	result.TotalMatches = len(result.Matches)
	result.MatchedFiles = len(fileSet)

	return result, nil
}

// fallbackGrepSearch 内置的grep搜索实现（当ripgrep不可用时）
func fallbackGrepSearch(query string, caseSensitive bool, includePattern, excludePattern, workDir string) (*GrepSearchResult, error) {
	result := &GrepSearchResult{
		Query:          query,
		CaseSensitive:  caseSensitive,
		IncludePattern: includePattern,
		ExcludePattern: excludePattern,
		Matches:        []GrepMatch{},
	}

	// 编译正则表达式
	var regexPattern string
	if !caseSensitive {
		regexPattern = "(?i)" + query
	} else {
		regexPattern = query
	}

	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	searchPath := "."
	if workDir != "" {
		searchPath = workDir
	}

	// 使用filepath.Walk遍历文件
	err = filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续处理其他文件
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 检查包含模式
		if includePattern != "" {
			matched, _ := filepath.Match(includePattern, filepath.Base(path))
			if !matched {
				return nil
			}
		}

		// 检查排除模式
		if excludePattern != "" {
			matched, _ := filepath.Match(excludePattern, filepath.Base(path))
			if matched {
				return nil
			}
		}

		// 读取并搜索文件内容
		return searchInFile(path, regex, caseSensitive, result)
	})

	if err != nil {
		return nil, err
	}

	// 计算文件数量
	fileSet := make(map[string]bool)
	for _, match := range result.Matches {
		fileSet[match.File] = true
	}
	result.MatchedFiles = len(fileSet)
	result.TotalMatches = len(result.Matches)

	return result, nil
}

// parseIntSafe 安全地解析整数
func parseIntSafe(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// extractMatch 从内容中提取匹配的部分
func extractMatch(content, query string, caseSensitive bool) string {
	if !caseSensitive {
		content = strings.ToLower(content)
		query = strings.ToLower(query)
	}

	// 简单的字符串匹配
	if idx := strings.Index(content, query); idx >= 0 {
		return content[idx : idx+len(query)]
	}

	return query
}

// searchInFile 在文件中搜索匹配项
func searchInFile(filePath string, regex *regexp.Regexp, caseSensitive bool, result *GrepSearchResult) error {
	file, err := os.Open(filePath)
	if err != nil {
		return nil // 忽略无法打开的文件
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	matchCount := 0

	for scanner.Scan() && matchCount < 50 { // 限制匹配数量
		lineNumber++
		line := scanner.Text()
		
		if regex.MatchString(line) {
			// 找到匹配项
			match := GrepMatch{
				File:     filePath,
				Line:     lineNumber,
				Content:  line,
				Match:    extractMatch(line, result.Query, caseSensitive),
			}
			
			result.Matches = append(result.Matches, match)
			matchCount++
		}
	}

	return scanner.Err()
}

// NewGrepSearchTool 创建grep_search工具
func NewGrepSearchTool() Tool {
	schema := ToolSchema{
		Name:        "grep_search",
		Description: "### Instructions:\nThis is best for finding exact text matches or regex patterns.\nThis is preferred over semantic search when we know the exact symbol/function name/etc. to search in some set of directories/file types.\n\nUse this tool to run fast, exact regex searches over text files using the `ripgrep` engine.\nTo avoid overwhelming output, the results are capped at 50 matches.\nUse the include or exclude patterns to filter the search scope by file type or specific paths.\n\n- Always escape special regex characters: ( ) [ ] { } + * ? ^ $ | . \\\n- Use `\\` to escape any of these characters when they appear in your search string.\n- Do NOT perform fuzzy or semantic matches.\n- Return only a valid regex pattern string.\n\n### Examples:\n| Literal               | Regex Pattern            |\n|-----------------------|--------------------------|\n| function(             | function\\(              |\n| value[index]          | value\\[index\\]         |\n| file.txt               | file\\.txt                |\n| user|admin            | user\\|admin             |\n| path\\to\\file         | path\\\\to\\\\file        |\n| hello world           | hello world              |\n| foo\\(bar\\)          | foo\\\\(bar\\\\)         |",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "The regex pattern to search for",
				},
				"case_sensitive": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether the search should be case sensitive",
				},
				"include_pattern": map[string]interface{}{
					"type":        "string",
					"description": "Glob pattern for files to include (e.g. '*.ts' for TypeScript files)",
				},
				"exclude_pattern": map[string]interface{}{
					"type":        "string",
					"description": "Glob pattern for files to exclude",
				},
				"explanation": map[string]interface{}{
					"type":        "string",
					"description": "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
				},
			},
			"required": []string{"query"},
		},
	}

	return Tool{
		Schema:   schema,
		Function: grepSearchFunction,
	}
} 