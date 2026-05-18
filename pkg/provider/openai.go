package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
)

type OpenAIProvider struct {
	client *openai.Client
}

func NewOpenAIProvider(apiKey, baseURL string) *OpenAIProvider {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)

	return &OpenAIProvider{client: &client}
}

func extractPartsFromMessage(msg Message) []openai.ChatCompletionContentPartUnionParam {
	parts := make([]openai.ChatCompletionContentPartUnionParam, 0, 1)
	switch msg.Type {
	case TextMessage:
		parts = append(parts, openai.TextContentPart(msg.Content))
	case ImageMessage:
		mimeType := DetectMimeFromBase64(msg.Content)
		dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, msg.Content)
		part := openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
			URL: dataURI,
		})
		parts = append(parts, part)
	}
	return parts
}

func reqToOpenAIParams(config *Config, model Model, historyMessages []Message) *openai.ChatCompletionNewParams {
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(historyMessages)+1)
	messages = append(messages, openai.SystemMessage(*config.SystemPrompt))
	for _, msg := range historyMessages {
		switch msg.Role {
		case UserRole:
			parts := extractPartsFromMessage(msg)
			messages = append(messages, openai.UserMessage(parts))
		case ModelRole:
			messages = append(messages, openai.AssistantMessage(msg.Content))
		}
	}

	params := openai.ChatCompletionNewParams{
		Model:    model.String(),
		Messages: messages,
	}

	if config.ThinkConfig != nil {
		params.SetExtraFields(map[string]any{
			"enable_thinking": config.ThinkConfig.EnableThink,
		})
	}

	if config.Translation != nil {
		translationOptions := map[string]any{
			"translation_options": map[string]string{
				"sourceLang": config.Translation.SourceLang,
				"targetLang": config.Translation.TargetLang,
			},
		}
		params.SetExtraFields(translationOptions)
	}

	buildOpenAITools(&params, config.Tools)

	if config.Temperature != nil {
		params.Temperature = openai.Float(intToFloat64(*config.Temperature))
	}

	if config.ResponseConfig != nil && config.ResponseConfig.JSONResponse {
		params.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{JSONSchema: shared.ResponseFormatJSONSchemaJSONSchemaParam{Schema: config.ResponseConfig.JSONSchema}},
		}
	}

	return &params
}

func buildOpenAITools(params *openai.ChatCompletionNewParams, tools []Tool) {
	for _, tool := range tools {
		switch tool {
		case WebSearch:
			params.SetExtraFields(map[string]any{
				"enable_search": true,
			})
		}
	}
}

func (p *OpenAIProvider) Complete(ctx context.Context, model Model, config *Config, messages []Message) (*GenerateResponse, error) {
	params := reqToOpenAIParams(config, model, messages)

	resp, err := p.client.Chat.Completions.New(ctx, *params)
	if err != nil {
		return nil, err
	}

	return openaiRespToResponse(resp), nil
}

func (p *OpenAIProvider) Stream(ctx context.Context, model Model, config *Config, messages []Message) <-chan StreamChunk {
	chunks := make(chan StreamChunk, 10)

	params := reqToOpenAIParams(config, model, messages)
	stream := p.client.Chat.Completions.NewStreaming(ctx, *params)
	go func() {
		defer close(chunks)
		for stream.Next() {
			evt := stream.Current()
			if len(evt.Choices) > 0 {
				isThink, thinkContent := extractThinkOAStream(evt.Choices[0].Delta)
				chunk := StreamChunk{IsThink: isThink, Text: evt.Choices[0].Delta.Content}
				if isThink && thinkContent != nil {
					chunk.Text = *thinkContent
				}
				select {
				case chunks <- chunk:
				case <-ctx.Done():
					return
				}
			}
		}

		if err := stream.Err(); err != nil {
			select {
			case chunks <- StreamChunk{Err: err}:
			case <-ctx.Done():
			}
		}
	}()

	return chunks
}

func openaiRespToResponse(resp *openai.ChatCompletion) *GenerateResponse {
	response := &GenerateResponse{}

	msg := resp.Choices[0].Message
	response.Response.Text = msg.Content

	extractThinkOA(&response.Response, &msg)

	extractTokenUsageOA(&response.TokenUsage, &resp.Usage)

	response.SearchRef = extractGroundingURLsOA(resp)

	return response
}

func extractThinkOAStream(delta openai.ChatCompletionChunkChoiceDelta) (bool, *string) {
	if field, exists := delta.JSON.ExtraFields["reasoning_content"]; exists {
		var thinkContent string
		err := json.Unmarshal([]byte(field.Raw()), &thinkContent)
		if err == nil && thinkContent != "" {
			return true, &thinkContent
		}
	}
	return false, nil
}

func extractThinkOA(customResp *Response, msg *openai.ChatCompletionMessage) {
	var extra struct {
		ReasoningContent string `json:"reasoning_content"`
	}

	if err := json.Unmarshal([]byte(msg.RawJSON()), &extra); err != nil {
		log.Printf("Failed to parse reasoning content: %v", err)
	}
	customResp.Think = &extra.ReasoningContent
}

func extractTokenUsageOA(customUsage *TokenUsage, resp *openai.CompletionUsage) {
	if resp == nil {
		customUsage.CachedToken = 0
		customUsage.InputToken = 0
		customUsage.OutputToken = 0
	} else {
		customUsage.CachedToken = int32(resp.PromptTokensDetails.CachedTokens)
		customUsage.InputToken = int32(resp.PromptTokens)
		customUsage.OutputToken = int32(resp.CompletionTokens)
	}
}

func extractGroundingURLsOA(resp *openai.ChatCompletion) []SearchReference {
	_ = resp
	return nil
}

func intToFloat64(temperature uint8) float64 {
	return float64(temperature) / 10.0
}
