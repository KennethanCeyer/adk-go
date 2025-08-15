package events

import (
	"github.com/KennethanCeyer/adk-go/models"
)

// Event represents a generic event within the ADK flow.
// It can encapsulate different types of data, such as LLM responses or content.
type Event struct {
	LlmResponse *models.LlmResponse
	Content     *models.Content
}
