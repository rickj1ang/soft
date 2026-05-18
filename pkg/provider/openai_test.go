package provider

import (
	"context"
	"os"
	"testing"
)

func newTestOpenAIProvider(t *testing.T, region string) *OpenAIProvider {
	t.Helper()
	api_key := "APIKEY" + "_" + region
	base_url := "BASEURL" + "_" + region

	apiKey := os.Getenv(api_key)
	baseURL := os.Getenv(base_url)
	openai := NewOpenAIProvider(apiKey, baseURL)

	return openai
}

func TestOpenAIGenerateText(t *testing.T) {
	openai := newTestOpenAIProvider(t, "BJ")
	systemPrompt := "you are a helpful assistant, and you will think or analyze in your thinking part and just answer what user want in a lean and concise manner"
	config := &Config{
		SystemPrompt: &systemPrompt,
		ThinkConfig: &ThinkConfig{
			EnableThink: true,
		},
	}
	messages := []Message{
		{
			Type:    TextMessage,
			Content: "please calculate 199 * 19 carefully",
			Role:    UserRole,
		},
	}
	resp, err := openai.Complete(context.TODO(), QWEN_36_FLASH, config, messages)
	if err != nil {
		t.Fatalf("failed to generate text: %v", err)
	}
	if resp == nil {
		t.Fatalf("expected non-nil response")
	}
	t.Logf("response think: %s", *resp.Response.Think)
	t.Logf("response text: %v", resp.Response.Text)
	t.Logf("token usage: %v", resp.TokenUsage)
}

func TestOpenAIGenerateJSON(t *testing.T) {
	openai := newTestOpenAIProvider(t, "BJ")
	systemPrompt := "you are a helpful assistant, and you will think or analyze in your thinking part and just answer what user want in a lean and concise manner"
	req := &Config{
		SystemPrompt: &systemPrompt,
		ResponseConfig: &ResponseConfig{
			JSONResponse: true,
			JSONSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"result": map[string]any{
						"type": "string",
					},
				},
			},
		},
	}
	messages := []Message{
		{
			Type:    TextMessage,
			Content: "please calculate 199 * 19 carefully, return in json format",
			Role:    UserRole,
		},
	}
	resp, err := openai.Complete(context.TODO(), QWEN_36_FLASH, req, messages)
	if err != nil {
		t.Fatalf("failed to generate text: %v", err)
	}
	if resp == nil {
		t.Fatalf("expected non-nil response")
	}
	t.Logf("response text: %v", resp.Response.Text)
	t.Logf("token usage: %v", resp.TokenUsage)
}

func TestOpenAIWebSearch(t *testing.T) {
	genai := newTestGenAIProvider(t)
	systemPrompt := "you are a helpful assistant, and you will think or analyze in your thinking part and just answer what user want in a lean and concise manner"
	req := &Config{
		SystemPrompt: &systemPrompt,
		Tools:        []Tool{WebSearch},
	}
	messages := []Message{
		{
			Type:    TextMessage,
			Content: "what is the price of BTC now",
			Role:    UserRole,
		},
	}
	resp, err := genai.Complete(context.TODO(), GEMINI_31_FLASH_LITE, req, messages)
	if err != nil {
		t.Fatalf("failed to generate text: %v", err)
	}
	if resp == nil {
		t.Fatalf("expected non-nil response")
	}
	t.Logf("response text: %v", resp.Response.Text)
	t.Logf("token usage: %v", resp.TokenUsage)
	for _, ref := range resp.SearchRef {
		t.Logf("search reference: url=%v, title=%v", ref.URL, ref.Title)
	}
}

func TestOpenAIStream(t *testing.T) {
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
	chunks := openai.Stream(context.TODO(), QWEN_35_PLUS, config, messages)
	for chunk := range chunks {
		if chunk.Err != nil {
			t.Fatalf("stream chunk error: %v", chunk.Err)
		}
		t.Logf("think part: %v", chunk.IsThink)
		t.Logf("stream chunk: %v", chunk.Text)
	}
}
