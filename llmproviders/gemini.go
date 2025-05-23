package llmproviders

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/KennethanCeyer/adk-go/adk"
	adktypes "github.com/KennethanCeyer/adk-go/adk/types"
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

func convertADKToolsToGenaiTools(adkTools []adk.Tool) []*genai.Tool {
	if len(adkTools) == 0 { return nil }
	genaiTools := make([]*genai.Tool, len(adkTools))
	for i, t := range adkTools {
		var paramSchema *genai.Schema
		paramSchemaData := t.Parameters()
		if paramSchemaData != nil {
			if schemaMap, ok := paramSchemaData.(map[string]any); ok {
				paramSchema = &genai.Schema{}
				if schemaTypeStr, typeOk := schemaMap["type"].(string); typeOk {
					switch schemaTypeStr {
					case "object": paramSchema.Type = genai.TypeObject
					case "string": paramSchema.Type = genai.TypeString
					case "integer": paramSchema.Type = genai.TypeInteger
					case "number": paramSchema.Type = genai.TypeNumber
					case "boolean": paramSchema.Type = genai.TypeBoolean
					case "array": paramSchema.Type = genai.TypeArray
						if itemsMap, itemsOk := schemaMap["items"].(map[string]any); itemsOk {
							itemSchema := &genai.Schema{}
							if itemTypeStr, itemTypeOk := itemsMap["type"].(string); itemTypeOk {
								switch itemTypeStr {
								case "string": itemSchema.Type = genai.TypeString; case "integer": itemSchema.Type = genai.TypeInteger; case "number": itemSchema.Type = genai.TypeNumber; default: itemSchema.Type = genai.TypeString
								}
							}
							paramSchema.Items = itemSchema
						}
					default: paramSchema.Type = genai.TypeObject
					}
				} else { paramSchema.Type = genai.TypeObject }

				if props, pOk := schemaMap["properties"].(map[string]any); pOk {
					paramSchema.Properties = make(map[string]*genai.Schema)
					for k, v := range props {
						propSchemaMap, psOk := v.(map[string]any);
						if !psOk { continue }
						propDesc, _ := propSchemaMap["description"].(string)
						var currentPropType genai.Type
						propTypeStr, typeOk := propSchemaMap["type"].(string)
						if !typeOk { currentPropType = genai.TypeString
						} else {
							switch propTypeStr {
							case "string": currentPropType = genai.TypeString; case "integer": currentPropType = genai.TypeInteger; case "number": currentPropType = genai.TypeNumber; case "boolean": currentPropType = genai.TypeBoolean; case "object": currentPropType = genai.TypeObject; case "array": currentPropType = genai.TypeArray; default: currentPropType = genai.TypeString
							}
						}
						paramSchema.Properties[k] = &genai.Schema{Type: currentPropType, Description: propDesc}
					}
				}
				if reqAny, rOk := schemaMap["required"].([]any); rOk {
					for _, req := range reqAny { if rStr, rsOk := req.(string); rsOk { paramSchema.Required = append(paramSchema.Required, rStr) } }
				} else if reqStr, rSOk := schemaMap["required"].([]string); rSOk { paramSchema.Required = reqStr }

			} else if schemaGenai, ok := paramSchemaData.(*genai.Schema); ok { paramSchema = schemaGenai
			} else if paramSchemaData != nil { log.Printf("Warning: Tool %s param schema is unhandled type %T", t.Name(), paramSchemaData); paramSchema = nil }
		}
		genaiTools[i] = &genai.Tool{FunctionDeclarations: []*genai.FunctionDeclaration{{Name: t.Name(), Description: t.Description(), Parameters: paramSchema}}}
	}
	return genaiTools
}

func convertADKMessagesToGenaiContent(messages []adktypes.Message) []*genai.Content {
	var contents []*genai.Content
	for _, msg := range messages {
		content := &genai.Content{Role: msg.Role}
		for _, p := range msg.Parts {
			var genaiPart genai.Part
			if p.Text != nil { genaiPart = genai.Text(*p.Text)
			} else if p.FunctionCall != nil { if msg.Role == "model" { genaiPart = genai.FunctionCall{Name: p.FunctionCall.Name, Args: p.FunctionCall.Args} } else { continue }
			} else if p.FunctionResponse != nil { genaiPart = genai.FunctionResponse{Name: p.FunctionResponse.Name, Response: p.FunctionResponse.Response}
			} else { continue }
			content.Parts = append(content.Parts, genaiPart)
		}
		if len(content.Parts) > 0 { contents = append(contents, content) }
	}
	return contents
}

func convertGenaiCandidateToADKMessage(candidate *genai.Candidate) adktypes.Message {
	msg := adktypes.Message{Role: "model"}
	if candidate == nil { text := "Error: LLM candidate nil."; msg.Parts = []adktypes.Part{{Text: &text}}; return msg }
	if candidate.Content == nil {
		text := "Error: LLM no content."
		if fr := candidate.FinishReason; fr != genai.FinishReasonStop && fr != genai.FinishReasonUnspecified { text = fmt.Sprintf("LLM stop: %s.", fr.String()) }
		msg.Parts = []adktypes.Part{{Text: &text}}; return msg
	}
	for _, p := range candidate.Content.Parts {
		var adkPart adktypes.Part
		switch v := p.(type) {
		case genai.Text: text := string(v); adkPart.Text = &text
		case genai.FunctionCall: adkPart.FunctionCall = &adktypes.FunctionCall{Name: v.Name, Args: v.Args}
		default: unsupportedText := fmt.Sprintf("[LLM Part Type %T]", v); adkPart.Text = &unsupportedText
		}
		msg.Parts = append(msg.Parts, adkPart)
	}
	return msg
}

func (g *GeminiLLMProvider) GenerateContent(
	ctx context.Context, modelName string, systemInstruction string, tools []adk.Tool,
	history []adktypes.Message, latestMessage adktypes.Message,
) (adktypes.Message, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(g.apiKey))
	if err != nil { return adktypes.Message{}, fmt.Errorf("genai client: %w", err) }
	defer client.Close()

	model := client.GenerativeModel(modelName)
	if systemInstruction != "" { model.SystemInstruction = &genai.Content{Parts: []genai.Part{genai.Text(systemInstruction)}} }
	if len(tools) > 0 { model.Tools = convertADKToolsToGenaiTools(tools) }

	chatSession := model.StartChat()
	if len(history) > 0 { chatSession.History = convertADKMessagesToGenaiContent(history) }
	
	latestGenaiContents := convertADKMessagesToGenaiContent([]adktypes.Message{latestMessage})
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
		if err == iterator.Done { break }
		if err != nil { return adktypes.Message{}, fmt.Errorf("stream LLM: %w", err) }
		if aggResp.Candidates == nil && aggResp.PromptFeedback == nil { aggResp.PromptFeedback = resp.PromptFeedback }
		if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
			for _, sp := range resp.Candidates[0].Content.Parts {
				if fc, ok := sp.(genai.FunctionCall); ok { aggParts = []genai.Part{fc}; break
				} else if t, ok := sp.(genai.Text); ok {
					if len(aggParts) > 0 { if lt, lOk := aggParts[len(aggParts)-1].(genai.Text); lOk { aggParts[len(aggParts)-1] = genai.Text(string(lt) + string(t)); continue } }
					aggParts = append(aggParts, t)
				} else { aggParts = append(aggParts, sp) }
			}
			aggResp.Candidates = []*genai.Candidate{{FinishReason: resp.Candidates[0].FinishReason}}
		}
	}
	if len(aggResp.Candidates) > 0 {
		if aggResp.Candidates[0].Content == nil { aggResp.Candidates[0].Content = &genai.Content{} }
		aggResp.Candidates[0].Content.Parts = aggParts; aggResp.Candidates[0].Content.Role = "model"
	} else if len(aggParts) > 0 { aggResp.Candidates = []*genai.Candidate{{Content: &genai.Content{Role: "model", Parts: aggParts}}} }

	if len(aggResp.Candidates) == 0 {
		bt := "LLM empty/no candidates."
	
		if pf := aggResp.PromptFeedback; pf != nil && pf.BlockReason != genai.BlockReasonUnspecified { bt = fmt.Sprintf("LLM blocked. R: %s. M: %s", pf.BlockReason.String(), pf.BlockReason.String()) }
		log.Println("Warning:", bt); return adktypes.Message{Role: "model", Parts: []adktypes.Part{{Text: &bt}}}, nil
	}
	return convertGenaiCandidateToADKMessage(aggResp.Candidates[0]), nil
}
