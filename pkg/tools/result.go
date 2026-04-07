package tools

type ToolResult struct {
	ForLLM  string `json:"for_llm"`
	ForUser string `json:"for_user,omitempty"`
	Silent  bool   `json:"silent"`
	IsError bool   `json:"is_error"`
	Async   bool   `json:"async"`
	Err     error  `json:"-"`
}

func SuccessResult(forLLM, forUser string) *ToolResult {
	return &ToolResult{ForLLM: forLLM, ForUser: forUser}
}

func ErrorResult(msg string) *ToolResult {
	return &ToolResult{
		ForLLM:  "Error: " + msg,
		ForUser: "Sorry, I encountered an error: " + msg,
		IsError: true,
	}
}

func SilentResult(forLLM string) *ToolResult {
	return &ToolResult{ForLLM: forLLM, Silent: true}
}
