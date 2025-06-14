package tools

// ToolFunction 工具函数类型
type ToolFunction func(params map[string]interface{}) (interface{}, error)

// ToolSchema 工具模式定义
type ToolSchema struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"input_schema"`
}

// Tool 工具定义
type Tool struct {
	Schema   ToolSchema
	Function ToolFunction
}

// ToolCall 工具调用请求
type ToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResult 工具调用结果
type ToolResult struct {
	Name    string      `json:"name"`
	Result  interface{} `json:"result"`
	Error   string      `json:"error,omitempty"`
	Success bool        `json:"success"`
}

// ToolManager 工具管理器接口
type ToolManager interface {
	RegisterTool(name string, tool Tool) error
	GetTool(name string) (Tool, bool)
	ListTools() []ToolSchema
	ExecuteTool(name string, params map[string]interface{}) (*ToolResult, error)
} 