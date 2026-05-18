package provider

import (
	"context"
	"os"
	"testing"
)

func newTestGenAIProvider(t *testing.T) *GenAIProvider {
	t.Helper()
	project := os.Getenv("GOOGLE_CLOUD_PROJECT")
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")

	genai, err := NewGenAIProvider(context.Background(), project, location)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}
	return genai
}

func TestGenerateText(t *testing.T) {
	genai := newTestGenAIProvider(t)
	systemPrompt := "you are a helpful assistant, and you will think or analyze in your thinking part and just answer what user want in a lean and concise manner"
	config := &Config{
		SystemPrompt: &systemPrompt,
		ThinkConfig: &ThinkConfig{
			EnableThink: true,
		},
	}
	contents := []Message{{
		Type:    TextMessage,
		Content: "please calculate 199 * 19 carefully",
		Role:    UserRole,
	}}
	resp, err := genai.Complete(context.TODO(), GEMINI_31_PRO, config, contents)
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

func TestGenerateJSON(t *testing.T) {
	genai := newTestGenAIProvider(t)
	systemPrompt := "you are a helpful assistant, and you will think or analyze in your thinking part and just answer what user want in a lean and concise manner"
	config := &Config{
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
	contents := []Message{{
		Type:    TextMessage,
		Content: "please calculate 199 * 19 carefully",
		Role:    UserRole,
	}}
	resp, err := genai.Complete(context.TODO(), GEMINI_31_PRO, config, contents)
	if err != nil {
		t.Fatalf("failed to generate text: %v", err)
	}
	if resp == nil {
		t.Fatalf("expected non-nil response")
	}
	t.Logf("response text: %v", resp.Response.Text)
	t.Logf("token usage: %v", resp.TokenUsage)
}

func TestWebSearch(t *testing.T) {
	genai := newTestGenAIProvider(t)
	systemPrompt := "you are a helpful assistant, and you will think or analyze in your thinking part and just answer what user want in a lean and concise manner"
	config := &Config{
		SystemPrompt: &systemPrompt,
		Tools:        []Tool{WebSearch},
	}
	contents := []Message{{
		Type:    TextMessage,
		Content: "what is the price of BTC now",
		Role:    UserRole,
	}}
	resp, err := genai.Complete(context.TODO(), GEMINI_31_FLASH_LITE, config, contents)
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

func TestGenAIStream(t *testing.T) {
	genai := newTestGenAIProvider(t)
	systemPrompt := "you are a helpful assistant, and you will think or analyze in your thinking part"
	config := &Config{
		SystemPrompt: &systemPrompt,
		ThinkConfig: &ThinkConfig{
			EnableThink: true,
			ThinkLevel:  High,
		},
	}
	message := []Message{{
		Type:    TextMessage,
		Content: "please calculate 199 * 19 carefully",
		Role:    UserRole,
	}}
	chunks := genai.Stream(context.TODO(), GEMINI_31_PRO, config, message)
	for chunk := range chunks {
		if chunk.Err != nil {
			t.Fatalf("stream chunk error: %v", chunk.Err)
		}
		t.Logf("think part: %v", chunk.IsThink)
		t.Logf("stream chunk: %v", chunk.Text)
	}
}
