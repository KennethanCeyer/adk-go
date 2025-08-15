package events

import (
	"github.com/KennethanCeyer/adk-go/models"
)

type Event struct {
	LlmResponse *models.LlmResponse
	Content     *models.Content
}
