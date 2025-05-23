package types

type Message struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

type Part struct {
	Text             *string
	FunctionCall     *FunctionCall
	FunctionResponse *FunctionResponse
}

type FunctionCall struct {
	Name string
	Args map[string]any
}

type FunctionResponse struct {
	Name     string
	Response map[string]any
}
