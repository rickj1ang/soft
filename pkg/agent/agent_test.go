package agent

import (
	"context"
	"testing"
	"time"

	"github.com/rickj1ang/soft/v2/pkg/provider"
)

type MockProvider struct {
}

func (m *MockProvider) Stream(ctx context.Context, model string, config provider.Config, messages []provider.Message) <-chan provider.StreamChunk {
	chunk := make(chan provider.StreamChunk, 10)
	go func() {
		for range 10 {
			select {
			case <-ctx.Done():
				return
			default:
				chunk <- provider.StreamChunk{Text: "laladidilala"}
				time.Sleep(time.Second)
			}
		}
		close(chunk)
	}()
	return chunk
}

type MockMemory struct {
	messages []provider.Message
}

func (m MockMemory) GetMessages() []provider.Message {
	return m.messages
}

func (m MockMemory) AddMessage(message provider.Message) {
	m.messages = append(m.messages, message)
}

func TestAgentStream(t *testing.T) {
	llm := &MockProvider{}
	memory := MockMemory{}
	agent := NewAgent(llm, &provider.Config{}, nil, "test")
	msg := provider.Message{
		Type:    provider.TextMessage,
		Role:    provider.UserRole,
		Content: "hi",
	}
	for chunk := range agent.Run(t.Context(), msg, memory) {
		t.Log(chunk.Message)
	}
}
