package example

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/KennethanCeyer/adk-go/tools"
	"github.com/google/generative-ai-go/genai"
)

type StockPriceTool struct{}

func NewStockPriceTool() tools.Tool {
	return &StockPriceTool{}
}

func (t *StockPriceTool) Name() string {
	return "get_stock_price"
}

func (t *StockPriceTool) Description() string {
	return "Fetches the current stock price for a given ticker symbol."
}

func (t *StockPriceTool) Parameters() any {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"ticker_symbol": {
				Type:        genai.TypeString,
				Description: "The stock ticker symbol, e.g., 'GOOGL' for Google.",
			},
		},
		Required: []string{"ticker_symbol"},
	}
}

func (t *StockPriceTool) Execute(ctx context.Context, args any) (any, error) {
	argsMap, ok := args.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("get_stock_price: invalid arguments format, expected map[string]any, got %T", args)
	}
	ticker, ok := argsMap["ticker_symbol"].(string)
	if !ok {
		return nil, fmt.Errorf("get_stock_price: invalid or missing 'ticker_symbol' argument")
	}

	// In a real application, you would call an external API here.
	// For this example, we'll return a mock price.
	price := 100.0 + rand.Float64()*(500.0-100.0) // Random price between 100 and 500
	return map[string]any{
		"ticker_symbol": ticker,
		"price":         fmt.Sprintf("%.2f", price),
		"currency":      "USD",
	}, nil
}

type CompanyNewsTool struct{}

func NewCompanyNewsTool() tools.Tool {
	return &CompanyNewsTool{}
}

func (t *CompanyNewsTool) Name() string {
	return "get_company_news"
}

func (t *CompanyNewsTool) Description() string {
	return "Fetches recent news headlines for a given company ticker symbol."
}

func (t *CompanyNewsTool) Parameters() any {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"ticker_symbol": {
				Type:        genai.TypeString,
				Description: "The stock ticker symbol, e.g., 'GOOGL' for Google.",
			},
		},
		Required: []string{"ticker_symbol"},
	}
}

func (t *CompanyNewsTool) Execute(ctx context.Context, args any) (any, error) {
	argsMap, ok := args.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("get_company_news: invalid arguments format, expected map[string]any, got %T", args)
	}
	ticker, ok := argsMap["ticker_symbol"].(string)
	if !ok {
		return nil, fmt.Errorf("get_company_news: invalid or missing 'ticker_symbol' argument")
	}

	headlines := []string{
		fmt.Sprintf("%s announces record Q3 earnings, stock surges.", ticker),
		fmt.Sprintf("New product launch from %s receives positive reviews.", ticker),
		fmt.Sprintf("Analysts upgrade %s to 'Buy' following innovation showcase.", ticker),
	}

	headlinesAsAny := make([]any, len(headlines))
	for i, h := range headlines {
		headlinesAsAny[i] = h
	}

	return map[string]any{
		"ticker_symbol": ticker,
		"headlines":     headlinesAsAny,
		"retrieved_at":  time.Now().Format(time.RFC3339),
	}, nil
}
