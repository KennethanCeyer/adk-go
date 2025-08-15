package models

import "github.com/KennethanCeyer/adk-go/models/types"

// Content is a generic container for data passed between components.
// It is similar to types.Message but used in different contexts like
// system instructions or LLM responses.
type Content struct {
	Parts []types.Part
	Role  string
}

// GenerateContentConfig holds configuration for content generation by the LLM.
type GenerateContentConfig struct {
	SystemInstruction *Content
	// Other potential fields: Temperature, TopP, TopK, MaxOutputTokens, etc.
}

// LlmRequest represents a request to be sent to the LLM, primarily for configuration.
type LlmRequest struct {
	Config *GenerateContentConfig
}

// LlmResponse represents a response received from the LLM.
type LlmResponse struct {
	Content *Content
}
