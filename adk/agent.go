package adk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	adktypes "github.com/KennethanCeyer/adk-go/adk/types"
)

type Tool interface {
	Name() string
	Description() string
	Parameters() any
	Execute(ctx context.Context, args any) (any, error)
}

type LLMProvider interface {
	GenerateContent(
		ctx context.Context,
		modelName string,
		systemInstruction string,
		tools []Tool,
		history []adktypes.Message,
		latestMessage adktypes.Message,
	) (adktypes.Message, error)
}

type Agent interface {
	Name() string
	SystemInstruction() string
	Tools() []Tool
	ModelName() string
	LLMProvider() LLMProvider
	Process(ctx context.Context, history []adktypes.Message, latestMessage adktypes.Message) (adktypes.Message, error)
}

// PrettyPrint formats and prints any given data structure with indentation for readability.
// Useful for debugging complex structs like agent responses.
func PrettyPrint(data interface{}) {
	var prettyJSON bytes.Buffer
	encoder := json.NewEncoder(&prettyJSON)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(data)
	if err != nil {
		fmt.Printf("Error pretty printing data: %v\n", err)
		// Fallback to default printing
		fmt.Printf("%+v\n", data)
		return
	}
	fmt.Println(prettyJSON.String())
}
