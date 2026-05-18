// Package provider is a unify interface of genai and openai
package provider

import "context"

type ThinkLevel int

const (
	Unspecified ThinkLevel = iota
	Minimal
	Low
	Medium
	High
	Extreme
)

type ThinkConfig struct {
	EnableThink bool
	ThinkLevel  ThinkLevel
}

type ResponseConfig struct {
	JSONResponse bool
	JSONSchema   map[string]any
}

type TranslationOpts struct {
	SourceLang string
	TargetLang string
}

type Tool struct{}

var (
	WebSearch Tool
	MapSearch Tool
)

type Config struct {
	SystemPrompt   *string
	Temperature    *uint8
	ThinkConfig    *ThinkConfig
	ResponseConfig *ResponseConfig
	Translation    *TranslationOpts
	Tools          []Tool
}

type MessageType int

const (
	TextMessage MessageType = iota
	ImageMessage
)

type Role int

const (
	UserRole Role = iota
	ModelRole
)

type Message struct {
	Role    Role
	Type    MessageType
	Content string
}

type TokenUsage struct {
	InputToken  int32
	OutputToken int32
	CachedToken int32
}

type SearchReference struct {
	Title string
	URL   string
}

type Response struct {
	Think *string
	Text  string
}

type GenerateResponse struct {
	Response   Response
	TokenUsage TokenUsage
	SearchRef  []SearchReference
}

type ImageRequest struct{}

type ImageResponse struct{}

type LLMProvider interface {
	Complete(ctx context.Context, model Model, config *Config, messages []Message) (*GenerateResponse, error)
	Stream(ctx context.Context, model Model, config *Config, messages []Message) <-chan StreamChunk
}
