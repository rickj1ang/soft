package memory

import (
	"fmt"

	"github.com/rickj1ang/soft/v2/pkg/provider"
)

type Memory struct {
	sessionID string
	messages  []provider.Message
}

func NewMemory(id string) *Memory {
	return &Memory{
		sessionID: id,
		messages:  make([]provider.Message, 0),
	}
}

func (m *Memory) AddMessage(msg provider.Message) {
	m.messages = append(m.messages, msg)
}

func (m *Memory) GetMessages() []provider.Message {
	return m.messages
}

func (m *Memory) SaveMessages() {
	fmt.Println("save to DB, now skip")
}
