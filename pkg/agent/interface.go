package agent

import (
	"context"

	"github.com/rickj1ang/soft/v2/pkg/provider"
)

type Provider interface {
	Stream(context.Context, string, provider.Config, []provider.Message) <-chan provider.StreamChunk
}

type Memory interface {
	GetMessages() []provider.Message
	AddMessage(provider.Message)
}
