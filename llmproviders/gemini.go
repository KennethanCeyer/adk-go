package llmproviders

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
	"github.com/KennethanCeyer/adk-go/tools"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type GeminiLLMProvider struct {
	apiKey string
}

func NewGeminiLLMProvider() (*GeminiLLMProvider, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}
	return &GeminiLLMProvider{apiKey: apiKey}, nil
}

func parseSchema(schemaMap map[string]any) *genai.Schema {
	if schemaMap == nil {
		return nil
	}

	schema := &genai.Schema{}

	if typeStr, ok := schemaMap["type"].(string); ok {
		switch typeStr {
		case "object":
			schema.Type = genai.TypeObject
		case "string":
			schema.Type = genai.TypeString
		case "integer":
			schema.Type = genai.TypeInteger
		case "number":
			schema.Type = genai.TypeNumber
		case "boolean":
			schema.Type = genai.TypeBoolean
		case "array":
			schema.Type = genai.TypeArray
			if itemsMap, itemsOk := schemaMap["items"].(map[string]any); itemsOk {
				schema.Items = parseSchema(itemsMap)
			}
		}
	}

	if props, ok := schemaMap["properties"].(map[string]any); ok && schema.Type == genai.TypeObject {
		schema.Properties = make(map[string]*genai.Schema)
		for key, val := range props {
			if propMap, propOk := val.(map[string]any); propOk {
				propSchema := parseSchema(propMap)
				if desc, descOk := propMap["description"].(string); descOk {
					propSchema.Description = desc
				}
				schema.Properties[key] = propSchema
			}
		}
	}

	if req, ok := schemaMap["required"].([]any); ok {
		for _, r := range req {
			if rStr, rOk := r.(string); rOk {
				schema.Required = append(schema.Required, rStr)
			}
		}
	} else if req, ok := schemaMap["required"].([]string); ok {
		schema.Required = req
	}

	return schema
}

func convertADKToolsToGenaiTools(adkTools []tools.Tool) []*genai.Tool {
	if len(adkTools) == 0 { return nil }
	genaiTools := make([]*genai.Tool, len(adkTools))
	for i, t := range adkTools {
		var paramSchema *genai.Schema
		paramSchemaData := t.Parameters()
		if paramSchemaData != nil {
			if schemaGenai, ok := paramSchemaData.(*genai.Schema); ok {
				paramSchema = schemaGenai
			} else if schemaMap, ok := paramSchemaData.(map[string]any); ok {
				paramSchema = parseSchema(schemaMap)
			} else if paramSchemaData != nil {
				log.Printf("Warning: Tool '%s' parameter schema is of unhandled type %T", t.Name(), paramSchemaData)
			}
		}
		genaiTools[i] = &genai.Tool{FunctionDeclarations: []*genai.FunctionDeclaration{{Name: t.Name(), Description: t.Description(), Parameters: paramSchema}}}
	}
	return genaiTools
}

func convertADKMessagesToGenaiContent(messages []modelstypes.Message) []*genai.Content {
	var genaiContents []*genai.Content
	for _, adkMessage := range messages {
		// Gemini API roles are "user" and "model". "function" role from ADK needs mapping.
		role := adkMessage.Role
		if role == "function" {
			role = "user" // Gemini API expects function responses to have the "user" role.
		}
		content := &genai.Content{Role: role}
		for _, p := range adkMessage.Parts {
			var genaiPart genai.Part
			if p.Text != nil { genaiPart = genai.Text(*p.Text)
			} else if p.FunctionCall != nil {
				// The ADK FunctionCall.Args is already map[string]any, so no assertion is needed.
				if adkMessage.Role == "model" {
					genaiPart = genai.FunctionCall{Name: p.FunctionCall.Name, Args: p.FunctionCall.Args}
				} else {
					continue
				}
			} else if p.FunctionResponse != nil {
				if respMap, ok := p.FunctionResponse.Response.(map[string]any); ok {
					genaiPart = genai.FunctionResponse{Name: p.FunctionResponse.Name, Response: respMap}
				} else {
					log.Printf("Warning: FunctionResponse.Response for tool '%s' is not a map[string]any, but %T. Skipping part.", p.FunctionResponse.Name, p.FunctionResponse.Response)
					continue
				}
			} else { continue }
			content.Parts = append(content.Parts, genaiPart)
		}
		if len(content.Parts) > 0 {
			genaiContents = append(genaiContents, content)
		}
	}
	return genaiContents
}

func convertGenaiCandidateToADKMessage(candidate *genai.Candidate) *modelstypes.Message {
	adkMessage := &modelstypes.Message{Role: "model"}
	if candidate == nil { text := "Error: LLM candidate was nil."; adkMessage.Parts = []modelstypes.Part{{Text: &text}}; return adkMessage }
	if candidate.Content == nil {
		text := "Error: LLM no content."
		if fr := candidate.FinishReason; fr != genai.FinishReasonStop && fr != genai.FinishReasonUnspecified { text = fmt.Sprintf("LLM stop: %s.", fr.String()) }
		adkMessage.Parts = []modelstypes.Part{{Text: &text}}; return adkMessage
	}
	for _, p := range candidate.Content.Parts {
		var adkPart modelstypes.Part
		switch v := p.(type) {
		case genai.Text: text := string(v); adkPart.Text = &text
		case genai.FunctionCall:
			// The genai.FunctionCall.Args is map[string]any, so no assertion is needed.
			adkPart.FunctionCall = &modelstypes.FunctionCall{Name: v.Name, Args: v.Args}
		default: unsupportedText := fmt.Sprintf("[LLM Part Type %T]", v); adkPart.Text = &unsupportedText
		}
		adkMessage.Parts = append(adkMessage.Parts, adkPart)
	}
	return adkMessage
}

func (g *GeminiLLMProvider) GenerateContent(
	ctx context.Context,
	modelName string,
	systemInstruction *modelstypes.Message,
	tools []tools.Tool,
	history []modelstypes.Message,
	latestMessage modelstypes.Message,
) (*modelstypes.Message, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(g.apiKey))
	if err != nil { return nil, fmt.Errorf("genai client: %w", err) }
	defer client.Close()

	model := client.GenerativeModel(modelName)
	if systemInstruction != nil && len(systemInstruction.Parts) > 0 && systemInstruction.Parts[0].Text != nil {
		model.SystemInstruction = &genai.Content{Parts: []genai.Part{genai.Text(*systemInstruction.Parts[0].Text)}}
	}
	if len(tools) > 0 { model.Tools = convertADKToolsToGenaiTools(tools) }

	chatSession := model.StartChat()
	if len(history) > 0 { chatSession.History = convertADKMessagesToGenaiContent(history) }
	
	latestGenaiContents := convertADKMessagesToGenaiContent([]modelstypes.Message{latestMessage})
	var latestPartsToSend []genai.Part
    if len(latestGenaiContents) > 0 && len(latestGenaiContents[0].Parts) > 0 {
        latestPartsToSend = latestGenaiContents[0].Parts
    } else {
        log.Println("Warning: latestMessage converted to empty parts for LLM. Sending an empty text part.");
        latestPartsToSend = []genai.Part{genai.Text("")} // Send a minimal valid part
    }

	iter := chatSession.SendMessageStream(ctx, latestPartsToSend...)
	var aggregatedParts []genai.Part
	var finalCandidate *genai.Candidate

	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed during LLM stream: %w", err)
		}

		// The last response in the stream contains the final state (e.g., FinishReason).
		if len(resp.Candidates) > 0 {
			finalCandidate = resp.Candidates[0]
			if finalCandidate.Content != nil {
				aggregatedParts = append(aggregatedParts, finalCandidate.Content.Parts...)
			}
		}
	}

	consolidatedParts := consolidateTextParts(aggregatedParts)

	if finalCandidate == nil {
		finalCandidate = &genai.Candidate{} // Create a blank candidate if stream was empty.
	}
	finalCandidate.Content = &genai.Content{
		Parts: consolidatedParts,
		Role:  "model",
	}

	return convertGenaiCandidateToADKMessage(finalCandidate), nil
}

func consolidateTextParts(parts []genai.Part) []genai.Part {
	if len(parts) == 0 {
		return nil
	}
	var result []genai.Part
	var textBuffer strings.Builder

	for _, part := range parts {
		if t, ok := part.(genai.Text); ok {
			textBuffer.WriteString(string(t))
		} else {
			// If there's pending text, write it out before the non-text part.
			if textBuffer.Len() > 0 {
				result = append(result, genai.Text(textBuffer.String()))
				textBuffer.Reset()
			}
			result = append(result, part)
		}
	}
	if textBuffer.Len() > 0 {
		result = append(result, genai.Text(textBuffer.String()))
	}
	return result
}
