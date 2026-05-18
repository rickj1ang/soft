package provider

import "context"

type RateLimiter interface {
	Take(context.Context, string) error
}

type RateLimitProvider struct {
	inner   LLMProvider
	limiter RateLimiter
}

func NewRateLimitProvider(inner LLMProvider, limiter RateLimiter) *RateLimitProvider {
	return &RateLimitProvider{
		inner:   inner,
		limiter: limiter,
	}
}

func (r *RateLimitProvider) Complete(ctx context.Context, model Model, config *Config, messages []Message) (*GenerateResponse, error) {
	if err := r.limiter.Take(ctx, model.String()); err != nil {
		return nil, err
	}
	return r.inner.Complete(ctx, model, config, messages)
}

func (r *RateLimitProvider) Stream(ctx context.Context, model Model, config *Config, messages []Message) <-chan StreamChunk {
	if err := r.limiter.Take(ctx, model.String()); err != nil {
		ch := make(chan StreamChunk, 1)
		go func() {
			ch <- StreamChunk{Err: err}
			close(ch)
		}()
		return ch
	}
	return r.inner.Stream(ctx, model, config, messages)
}
