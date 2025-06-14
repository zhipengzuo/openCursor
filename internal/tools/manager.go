package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// DefaultToolManager 默认工具管理器实现
type DefaultToolManager struct {
	tools   map[string]Tool
	mu      sync.RWMutex
	workDir string // 工作目录，用于解析相对路径
}

// NewDefaultToolManager 创建新的工具管理器
func NewDefaultToolManager() *DefaultToolManager {
	workDir, _ := os.Getwd()
	return &DefaultToolManager{
		tools:   make(map[string]Tool),
		workDir: workDir,
	}
}

// SetWorkDirectory 设置工作目录
func (tm *DefaultToolManager) SetWorkDirectory(dir string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.workDir = dir
}

// GetWorkDirectory 获取工作目录
func (tm *DefaultToolManager) GetWorkDirectory() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.workDir
}

// RegisterTool 注册工具
func (tm *DefaultToolManager) RegisterTool(name string, tool Tool) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	if _, exists := tm.tools[name]; exists {
		return fmt.Errorf("tool '%s' already registered", name)
	}
	
	tm.tools[name] = tool
	return nil
}

// GetTool 获取工具
func (tm *DefaultToolManager) GetTool(name string) (Tool, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	tool, exists := tm.tools[name]
	return tool, exists
}

// ListTools 列出所有工具的模式
func (tm *DefaultToolManager) ListTools() []ToolSchema {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	schemas := make([]ToolSchema, 0, len(tm.tools))
	for _, tool := range tm.tools {
		schemas = append(schemas, tool.Schema)
	}
	
	return schemas
}

// ExecuteTool 执行工具
func (tm *DefaultToolManager) ExecuteTool(name string, params map[string]interface{}) (*ToolResult, error) {
	tm.mu.RLock()
	tool, exists := tm.tools[name]
	workDir := tm.workDir
	tm.mu.RUnlock()
	
	if !exists {
		return &ToolResult{
			Name:    name,
			Success: false,
			Error:   fmt.Sprintf("tool '%s' not found", name),
		}, nil
	}
	
	// 为工具执行提供工作目录上下文
	if params == nil {
		params = make(map[string]interface{})
	}
	params["__work_dir__"] = workDir
	
	result, err := tool.Function(params)
	if err != nil {
		return &ToolResult{
			Name:    name,
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	
	return &ToolResult{
		Name:    name,
		Result:  result,
		Success: true,
	}, nil
}

// ResolvePath 解析路径，如果是相对路径则基于工作目录解析
func (tm *DefaultToolManager) ResolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	
	tm.mu.RLock()
	workDir := tm.workDir
	tm.mu.RUnlock()
	
	return filepath.Join(workDir, path)
} 