package financial_analyst

import (
	"fmt"

	"github.com/KennethanCeyer/adk-go/agents"
	agentinterfaces "github.com/KennethanCeyer/adk-go/agents/interfaces"
	"github.com/KennethanCeyer/adk-go/examples"
	"github.com/KennethanCeyer/adk-go/llmproviders"
	"github.com/KennethanCeyer/adk-go/models/types"
	"github.com/KennethanCeyer/adk-go/tools"
	"github.com/KennethanCeyer/adk-go/tools/example"
)

const financialAnalystInstruction = "You are a helpful financial analyst. When the conversation starts, introduce yourself and what you can do. For example: 'Hello, I am a financial analyst agent. I can provide the latest stock price and company news for a given ticker symbol. Which company are you interested in?'. To create a report, you must use your tools to gather the latest stock price and company news. Use the `get_stock_price` tool for prices and the `get_company_news` tool for news. Synthesize the information from these tools into a concise report for the user."

// NewFinancialAnalystAgent creates a new financial analyst agent.
// This factory pattern is a good practice for creating modular and testable agents.
func NewFinancialAnalystAgent() (agentinterfaces.LlmAgent, error) {
	provider, err := llmproviders.NewGeminiLLMProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini llm provider: %w", err)
	}

	instruction := financialAnalystInstruction
	agent := agents.NewBaseLlmAgent(
		"financial_analyst",
		"An agent that provides stock prices and company news.",
		"gemini-2.5-flash",
		&types.Message{Parts: []types.Part{{Text: &instruction}}},
		provider,
		[]tools.Tool{
			example.NewStockPriceTool(),
			example.NewCompanyNewsTool(),
		},
	)
	return agent, nil
}

func init() {
	agent, err := NewFinancialAnalystAgent()
	if err != nil {
		examples.RegisterAgent("financial_analyst", nil, err)
		return
	}
	examples.RegisterAgent("financial_analyst", agent, nil)
}
