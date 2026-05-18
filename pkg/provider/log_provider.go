package provider

import (
	"context"
	"time"
)

type Logger interface {
	// save to DB or just log to stdio is up to the implementation
	Log(ctx context.Context, model Model, elapsed time.Duration, userPrompt string, resp GenerateResponse, err error) error
}

type LogProvider struct {
	inner  LLMProvider
	logger Logger
}

func NewLogProvider(inner LLMProvider, logger Logger) *LogProvider {
	return &LogProvider{
		inner:  inner,
		logger: logger,
	}
}

func (r *LogProvider) Complete(ctx context.Context, model Model, config *Config, messages []Message) (*GenerateResponse, error) {
	if len(messages) < 1 {
		return nil, ErrEmptyMessages
	}

	start := time.Now()
	resp, err := r.inner.Complete(ctx, model, config, messages)
	elapsed := time.Since(start)

	if resp != nil {
		r.logger.Log(ctx, model, elapsed, messages[len(messages)-1].Content, *resp, err)
	}

	return resp, nil
}

func (r *LogProvider) Stream(ctx context.Context, model Model, config *Config, messages []Message) <-chan StreamChunk {
	return r.inner.Stream(ctx, model, config, messages)
}
