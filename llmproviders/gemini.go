package llmproviders

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
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

// parseSchema recursively converts a map-based schema definition into a genai.Schema.
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
				// Ensure args are of type map[string]any for genai.FunctionCall
				argsMap, ok := p.FunctionCall.Args.(map[string]any)
				if !ok && p.FunctionCall.Args != nil {
					log.Printf("Warning: FunctionCall args for '%s' is not map[string]any, but %T. Attempting to convert.", p.FunctionCall.Name, p.FunctionCall.Args)
					// Attempt a reflection-based conversion if not the right type, though this is brittle.
					// A better solution is to ensure the caller provides the correct type.
					v := reflect.ValueOf(p.FunctionCall.Args)
					if v.Kind() == reflect.Map {
						argsMap = make(map[string]any)
						for _, key := range v.MapKeys() {
							argsMap[key.String()] = v.MapIndex(key).Interface()
						}
						ok = true
					}
				}
				if ok && adkMessage.Role == "model" {
					genaiPart = genai.FunctionCall{Name: p.FunctionCall.Name, Args: argsMap}
				} else {
					continue
				}
			} else if p.FunctionResponse != nil { genaiPart = genai.FunctionResponse{Name: p.FunctionResponse.Name, Response: p.FunctionResponse.Response}
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
			if argsMap, ok := v.Args.(map[string]any); ok {
				adkPart.FunctionCall = &modelstypes.FunctionCall{Name: v.Name, Args: argsMap}
			} else {
				log.Printf("Warning: Received FunctionCall with non-map args from LLM: %T", v.Args)
				continue
			}
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
	var aggResp genai.GenerateContentResponse; var aggParts []genai.Part
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("stream LLM: %w", err)
		}
		if aggResp.Candidates == nil && aggResp.PromptFeedback == nil { aggResp.PromptFeedback = resp.PromptFeedback }
		if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
			for _, sp := range resp.Candidates[0].Content.Parts {
				if fc, ok := sp.(genai.FunctionCall); ok { aggParts = []genai.Part{fc}; break
				} else if t, ok := sp.(genai.Text); ok {
					if len(aggParts) > 0 {
						if lt, lOk := aggParts[len(aggParts)-1].(genai.Text); lOk {
							aggParts[len(aggParts)-1] = genai.Text(strings.Join([]string{string(lt), string(t)}, ""))
							continue
						}
					}
					aggParts = append(aggParts, t)
				} else { aggParts = append(aggParts, sp) }
			}
			aggResp.Candidates = []*genai.Candidate{{FinishReason: resp.Candidates[0].FinishReason}}
		}
	}
	if len(aggResp.Candidates) == 0 { aggResp.Candidates = []*genai.Candidate{{}} }
	if aggResp.Candidates[0].Content == nil { aggResp.Candidates[0].Content = &genai.Content{} }
	aggResp.Candidates[0].Content.Parts = aggParts
	aggResp.Candidates[0].Content.Role = "model"

	return convertGenaiCandidateToADKMessage(aggResp.Candidates[0]), nil
}
