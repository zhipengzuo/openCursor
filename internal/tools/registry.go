package tools

import (
	"fmt"
)

// Registry 工具注册器
type Registry struct {
	manager ToolManager
}

// NewRegistry 创建新的工具注册器
func NewRegistry() *Registry {
	return &Registry{
		manager: NewDefaultToolManager(),
	}
}

// GetManager 获取工具管理器
func (r *Registry) GetManager() ToolManager {
	return r.manager
}

// SetWorkDirectory 设置工作目录
func (r *Registry) SetWorkDirectory(dir string) {
	if tm, ok := r.manager.(*DefaultToolManager); ok {
		tm.SetWorkDirectory(dir)
	}
}

// RegisterAllTools 注册所有工具
func (r *Registry) RegisterAllTools() error {
	// 注册 read_file 工具
	if err := r.manager.RegisterTool("read_file", NewReadFileTool()); err != nil {
		return fmt.Errorf("failed to register read_file tool: %w", err)
	}

	// 注册 run_terminal_cmd 工具
	if err := r.manager.RegisterTool("run_terminal_cmd", NewRunTerminalCmdTool()); err != nil {
		return fmt.Errorf("failed to register run_terminal_cmd tool: %w", err)
	}

	// 注册 list_dir 工具
	if err := r.manager.RegisterTool("list_dir", NewListDirTool()); err != nil {
		return fmt.Errorf("failed to register list_dir tool: %w", err)
	}

	// 注册 grep_search 工具
	if err := r.manager.RegisterTool("grep_search", NewGrepSearchTool()); err != nil {
		return fmt.Errorf("failed to register grep_search tool: %w", err)
	}

	// 注册 search_replace 工具
	if err := r.manager.RegisterTool("search_replace", NewSearchReplaceTool()); err != nil {
		return fmt.Errorf("failed to register search_replace tool: %w", err)
	}

	// 注册 file_search 工具
	if err := r.manager.RegisterTool("file_search", NewFileSearchTool()); err != nil {
		return fmt.Errorf("failed to register file_search tool: %w", err)
	}

	// 注册 delete_file 工具
	if err := r.manager.RegisterTool("delete_file", NewDeleteFileTool()); err != nil {
		return fmt.Errorf("failed to register delete_file tool: %w", err)
	}

	// 注册 write_file 工具
	if err := r.manager.RegisterTool("write_file", NewWriteFileTool()); err != nil {
		return fmt.Errorf("failed to register write_file tool: %w", err)
	}

	return nil
}

// DefaultRegistry 默认的全局工具注册器
var DefaultRegistry = NewRegistry()

// RegisterDefaultTools 注册默认工具到全局注册器
func RegisterDefaultTools() error {
	return DefaultRegistry.RegisterAllTools()
}

// GetDefaultManager 获取默认工具管理器
func GetDefaultManager() ToolManager {
	return DefaultRegistry.GetManager()
}

// SetDefaultWorkDirectory 设置默认工作目录
func SetDefaultWorkDirectory(dir string) {
	DefaultRegistry.SetWorkDirectory(dir)
} 