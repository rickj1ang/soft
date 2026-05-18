package provider

import (
	"context"

	"google.golang.org/genai"
)

type GenAIProvider struct {
	client *genai.Client
}

func NewGenAIProvider(ctx context.Context, project, location string) (*GenAIProvider, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  project,
		Location: location,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		return nil, err
	}
	return &GenAIProvider{client: client}, nil
}

func appendContentBaseOnType(msg Message, contents []*genai.Content, role genai.Role) []*genai.Content {
	switch msg.Type {
	case TextMessage:
		contents = append(contents, genai.NewContentFromText(msg.Content, role))
	case ImageMessage:
		imgType := DetectMimeFromBase64(msg.Content)
		imgData, err := DecodeImageBase64(msg.Content)
		if err != nil {
			return contents
		}
		contents = append(contents, genai.NewContentFromBytes(imgData, imgType, role))
	}
	return contents
}

func userMessageToGenAIMessage(messages []Message) ([]*genai.Content, error) {
	contents := []*genai.Content{}
	for _, msg := range messages {
		switch msg.Role {
		case UserRole:
			contents = appendContentBaseOnType(msg, contents, genai.RoleUser)
		case ModelRole:
			contents = appendContentBaseOnType(msg, contents, genai.RoleModel)
		}
	}
	return contents, nil
}

func (p *GenAIProvider) Complete(ctx context.Context, model Model, config *Config, messages []Message) (*GenerateResponse, error) {
	contents, err := userMessageToGenAIMessage(messages)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Models.GenerateContent(ctx, model.String(), contents, reqToGenAIConfig(config))
	if err != nil {
		return nil, err
	}
	if resp.Text() == "" {
		return nil, ErrEmptyResponse
	}

	response := genaiRespToResponse(resp)
	return response, nil
}

type StreamChunk struct {
	IsThink bool
	Text    string
	Err     error
}

func (p *GenAIProvider) Stream(ctx context.Context, model Model, config *Config, messages []Message) <-chan StreamChunk {
	chunks := make(chan StreamChunk, 10)

	contents, err := userMessageToGenAIMessage(messages)
	if err != nil {
		select {
		case chunks <- StreamChunk{Err: err}:
		case <-ctx.Done():
		}
		return chunks
	}

	gaConfig := reqToGenAIConfig(config)

	go func() {
		defer close(chunks)

		for result, err := range p.client.Models.GenerateContentStream(ctx, model.String(), contents, gaConfig) {
			if err != nil {
				select {
				case chunks <- StreamChunk{Err: err}:
				case <-ctx.Done():
				}
				return
			}
			if result == nil {
				continue
			}

			text := result.Text()

			isThink, thinkText := extractThinkGA(result)
			respChunk := StreamChunk{
				IsThink: isThink,
				Text:    text,
			}
			if isThink {
				respChunk.Text = *thinkText
			}
			select {
			case chunks <- respChunk:
			case <-ctx.Done():
				return
			}
		}
	}()
	return chunks
}

func genaiRespToResponse(resp *genai.GenerateContentResponse) *GenerateResponse {
	response := &GenerateResponse{}

	response.Response.Text = resp.Text()

	_, response.Response.Think = extractThinkGA(resp)

	extractTokenUsageGA(&response.TokenUsage, resp)

	response.SearchRef = extractGroundingURLsGA(resp)

	return response
}

func extractThinkGA(genaiResp *genai.GenerateContentResponse) (bool, *string) {
	if len(genaiResp.Candidates) == 0 || genaiResp.Candidates[0] == nil {
		return false, nil
	}
	cand := genaiResp.Candidates[0]
	if cand.Content == nil {
		return false, nil
	}
	for _, part := range cand.Content.Parts {
		if part != nil && part.Thought && part.Text != "" {
			return true, &part.Text
		}
	}
	return false, nil
}

func extractTokenUsageGA(customUsage *TokenUsage, genaiResp *genai.GenerateContentResponse) {
	if genaiResp.UsageMetadata == nil {
		customUsage.CachedToken = 0
		customUsage.InputToken = 0
		customUsage.OutputToken = 0
	} else {
		customUsage.InputToken = genaiResp.UsageMetadata.PromptTokenCount
		customUsage.CachedToken = genaiResp.UsageMetadata.CachedContentTokenCount
		customUsage.OutputToken = genaiResp.UsageMetadata.CandidatesTokenCount
	}
}

func extractGroundingURLsGA(genaiResp *genai.GenerateContentResponse) []SearchReference {
	var result []SearchReference
	if genaiResp == nil || len(genaiResp.Candidates) == 0 {
		return result
	}
	dedup := make(map[string]struct{})
	for _, c := range genaiResp.Candidates {
		if c == nil || c.GroundingMetadata == nil {
			continue
		}
		for _, ch := range c.GroundingMetadata.GroundingChunks {
			if ch == nil || ch.Web == nil {
				continue
			}
			uri := ch.Web.URI
			title := ch.Web.Title
			if uri == "" {
				continue
			}
			if _, ok := dedup[uri]; ok {
				continue
			}
			dedup[uri] = struct{}{}
			result = append(result, SearchReference{Title: title, URL: uri})
			if len(result) >= 5 {
				return result
			}
		}
	}
	return result
}

func reqToGenAIConfig(req *Config) *genai.GenerateContentConfig {
	config := &genai.GenerateContentConfig{}
	if req.SystemPrompt != nil {
		config.SystemInstruction = genai.NewContentFromText(*req.SystemPrompt, genai.RoleUser)
	}

	if req.Temperature != nil {
		config.Temperature = intToFloat32(*req.Temperature)
	}

	if req.ThinkConfig != nil {
		config.ThinkingConfig = &genai.ThinkingConfig{
			IncludeThoughts: req.ThinkConfig.EnableThink,
			ThinkingLevel:   genai.ThinkingLevelUnspecified,
		}
		switch req.ThinkConfig.ThinkLevel {
		case Minimal:
			config.ThinkingConfig.ThinkingLevel = genai.ThinkingLevelMinimal
		case Low:
			config.ThinkingConfig.ThinkingLevel = genai.ThinkingLevelLow
		case Medium:
			config.ThinkingConfig.ThinkingLevel = genai.ThinkingLevelMedium
		case High:
			config.ThinkingConfig.ThinkingLevel = genai.ThinkingLevelHigh
		case Extreme:
			config.ThinkingConfig.ThinkingLevel = genai.ThinkingLevelHigh
		default:
			config.ThinkingConfig.ThinkingLevel = genai.ThinkingLevelUnspecified
		}
	} else {
		config.ThinkingConfig = &genai.ThinkingConfig{
			IncludeThoughts: false,
		}
	}

	if req.ResponseConfig != nil {
		if req.ResponseConfig.JSONResponse {
			config.ResponseMIMEType = "application/json"
			config.ResponseJsonSchema = req.ResponseConfig.JSONSchema
		}
	}

	config.Tools = buildGenAITools(req.Tools)

	return config
}

func buildGenAITools(tools []Tool) []*genai.Tool {
	genAITools := []*genai.Tool{}
	for _, tool := range tools {
		switch tool {
		case WebSearch:
			genAITools = append(genAITools, &genai.Tool{GoogleSearch: &genai.GoogleSearch{}})
		case MapSearch:
			genAITools = append(genAITools, &genai.Tool{GoogleMaps: &genai.GoogleMaps{}})
		}
	}
	return genAITools
}

func intToFloat32(temperature uint8) *float32 {
	temp := float32(temperature) / 10.0
	return &temp
}
