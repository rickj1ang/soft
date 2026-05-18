package provider

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type MockRateLimiter struct {
	limitConfig map[string]int
}

var syncMap = sync.Map{}

func NewMockRateLimiter(limitConfig map[string]int) *MockRateLimiter {
	return &MockRateLimiter{limitConfig: limitConfig}
}

func syncMapInit() {
	syncMap.Store("qwen3.5-flash", 0)
	syncMap.Store("deepseek-v4-pro", 0)
}

func NewMockRateLimiterWithDefaults(t *testing.T) *MockRateLimiter {
	t.Helper()
	limiter := map[string]int{
		"qwen3.5-flash":   1,
		"deepseek-v4-pro": 1,
	}
	return NewMockRateLimiter(limiter)
}

func (m *MockRateLimiter) Take(ctx context.Context, model string) error {
	if count, ok := m.limitConfig[model]; ok {
		if value, ok := syncMap.Load(model); ok {
			if value.(int) < count {
				syncMap.Store(model, value.(int)+1)
				return nil
			}
			return errors.New("rate limit exceed")
		} else {
			return errors.New("no config")
		}
	}
	return errors.New("no config")
}

func TestRateLimiter(t *testing.T) {
	syncMapInit()
	rateLimiter := NewMockRateLimiterWithDefaults(t)
	openai := newTestOpenAIProvider(t, "BJ")
	systemPrompt := "you are a helpful assistant, and you will think or analyze in your thinking part"
	config := &Config{
		SystemPrompt: &systemPrompt,
		ThinkConfig: &ThinkConfig{
			EnableThink: true,
			ThinkLevel:  High,
		},
	}
	messages := []Message{
		{
			Type:    TextMessage,
			Content: "please calculate 199 * 19 carefully",
			Role:    UserRole,
		},
	}

	rateProvider := NewRateLimitProvider(openai, rateLimiter)
	var wg sync.WaitGroup
	wg.Go(
		func() {
			t.Log("first call")
			resp, err := rateProvider.Complete(t.Context(), QWEN_35_FLASH, config, messages)
			if err != nil {
				t.Log(err.Error())
			}
			t.Log(resp.Response.Text)
		},
	)
	time.Sleep(time.Second)
	wg.Go(func() {
		t.Log("second call")
		resp, err := rateProvider.Complete(t.Context(), QWEN_35_FLASH, config, messages)
		if err != nil {
			t.Log(err.Error())
		} else {
			t.Log(resp.Response.Text)
		}
	})

	wg.Wait()
}
