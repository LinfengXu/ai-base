package aibase

// Message represents an individual message with its role and content.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (m *Messages) Add(role string, content string) {
	m.Messages = append(m.Messages, Message{role, content})
}

// AddSystemMessage 方法用于向Messages结构体中添加一条系统消息
func (m *Messages) SystemAdd(content string) {
	m.Messages = append(m.Messages, Message{"system", content})
}

// AddUserMessage 方法用于向Messages结构体中添加一条用户消息
func (m *Messages) UserAdd(content string) {
	m.Messages = append(m.Messages, Message{"user", content})
}

// AddAssistantMessage 方法用于向Messages结构体中添加一条助手消息
func (m *Messages) AssistantAdd(content string) {
	m.Messages = append(m.Messages, Message{"assistant", content})
}

type Messages struct {
	Messages []Message `json:"messages"`
}
type QwenLongStreamUseReq struct {
	AiChatCommon
	Messages    []Message `query:"messages" json:"messages"`
	GroupID     uint      `query:"group_id" json:"group_id"`
	ServiceType string    `query:"service_type" json:"service_type"` //   服务类型
}

type AiChatCommon struct {
	MaxTokens   int     `query:"max_tokens" json:"max_tokens"`
	TopP        float64 `query:"top_p" json:"top_p"`
	Temperature float64 `query:"temperature" json:"temperature"`
}

type ResponseToFontend struct {
	Content    string `json:"content"`
	IsFinished bool   `json:"is_finished"`
	SessionId  uint   `json:"session_id"`
	Problem    string `json:"problem"`
}

type OpenaiResponse struct {
	Choices []Choice `json:"choices"`
	Object  string   `json:"object"`
}

type Choice struct {
	Delta        Delta  `json:"delta"`
	FinishReason string `json:"finish_reason"`
	Index        int    `json:"index"`
	Logprobs     string `json:"logprobs"`
}

type Delta struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}
