package agent

import (
	"context"
	"strings"
	"victory/ai_gateway/v2/pkg/provider"
)

type Tool struct {
}

type Agent struct {
	Provider     Provider
	SystemPrompt string
	Tools        []Tool
	Config       *provider.Config
	Model        string
}

type AgentMessagePart int

const (
	AgentMessageThink AgentMessagePart = iota
	AgentMessageNoThink
)

type AgentMessage struct {
	Type    AgentMessagePart //
	Message string
}

func NewAgent(llm Provider, config *provider.Config, tools []Tool, model string) *Agent {
	return &Agent{
		Provider: llm,
		Tools:    tools,
		Config:   config,
		Model:    model,
	}
}

func (a *Agent) Run(ctx context.Context, message provider.Message, memory Memory) <-chan AgentMessage {
	chunks := make(chan AgentMessage, 10)
	messages := memory.GetMessages()
	go func() {
		defer close(chunks)
		var messageBuffer strings.Builder
		for chunk := range a.Provider.Stream(ctx, a.Model, *a.Config, messages) {
			msg := AgentMessage{Type: AgentMessageThink, Message: chunk.Text}
			if !chunk.IsThink {
				msg.Type = AgentMessageNoThink
				messageBuffer.WriteString(chunk.Text)
			}
			select {
			case chunks <- msg:
			case <-ctx.Done():
			}
		}
		if messageBuffer.Len() > 0 {
			memory.AddMessage(provider.Message{
				Role:    provider.ModelRole,
				Type:    provider.TextMessage,
				Content: messageBuffer.String(),
			})
		}
	}()

	return chunks
}
