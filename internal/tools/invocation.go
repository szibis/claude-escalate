package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/szibis/claude-escalate/internal/gateway"
)

// InvocationRequest represents a request to invoke tools
type InvocationRequest struct {
	Intent       string                 `json:"intent"`
	Input        string                 `json:"input"`
	Context      map[string]interface{} `json:"context,omitempty"`
	MaxDuration  time.Duration          `json:"max_duration"`
	MaxResults   int                    `json:"max_results"`
	TokenBudget  int                    `json:"token_budget"`
}

// InvocationResult represents the result of tool invocation
type InvocationResult struct {
	ToolName    string                 `json:"tool_name"`
	Success     bool                   `json:"success"`
	Output      interface{}            `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	DurationMs  int64                  `json:"duration_ms"`
	TokensUsed  int                    `json:"tokens_used"`
	Cached      bool                   `json:"cached"`
}

// Invoker handles tool invocation
type Invoker struct {
	factory    *gateway.AdapterFactory
	selector   *Selector
	cache      map[string]InvocationResult
	maxResults int
	timeout    time.Duration
}

// NewInvoker creates a new tool invoker
func NewInvoker(factory *gateway.AdapterFactory, selector *Selector) *Invoker {
	return &Invoker{
		factory:    factory,
		selector:   selector,
		cache:      make(map[string]InvocationResult),
		maxResults: 3,
		timeout:    30 * time.Second,
	}
}

// Invoke invokes selected tools based on intent
func (i *Invoker) Invoke(ctx context.Context, req *InvocationRequest) ([]InvocationResult, error) {
	if req.MaxDuration > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.MaxDuration)
		defer cancel()
	}

	// Select tools for this intent
	selectedTools, err := i.selector.SelectForIntent(req.Intent)
	if err != nil {
		return nil, fmt.Errorf("selecting tools: %w", err)
	}

	if len(selectedTools) == 0 {
		return []InvocationResult{}, nil
	}

	// Limit number of results
	if req.MaxResults > 0 && len(selectedTools) > req.MaxResults {
		selectedTools = selectedTools[:req.MaxResults]
	}

	// Invoke selected tools in parallel (up to max 2 concurrent)
	results := make([]InvocationResult, 0)
	for idx, toolName := range selectedTools {
		if idx >= 2 {
			break // Max parallel: 2 tools
		}

		result := i.invokeOne(ctx, toolName, req)
		results = append(results, result)
	}

	return results, nil
}

// invokeOne invokes a single tool and returns result
func (i *Invoker) invokeOne(ctx context.Context, toolName string, req *InvocationRequest) InvocationResult {
	start := time.Now()

	// Check cache
	if cached, ok := i.cache[toolName]; ok {
		cached.Cached = true
		cached.DurationMs = time.Since(start).Milliseconds()
		return cached
	}

	// Get adapter
	adapter, err := i.factory.GetAdapter(toolName)
	if err != nil {
		return InvocationResult{
			ToolName:   toolName,
			Success:    false,
			Error:      fmt.Sprintf("adapter not found: %v", err),
			DurationMs: time.Since(start).Milliseconds(),
		}
	}

	// Create tool request
	toolReq := &gateway.ToolRequest{
		Tool:   toolName,
		Input:  req.Input,
		Params: req.Context,
	}

	// Execute with timeout
	execCtx, cancel := context.WithTimeout(ctx, i.timeout)
	defer cancel()

	resp, err := adapter.Execute(execCtx, toolReq)
	if err != nil {
		return InvocationResult{
			ToolName:   toolName,
			Success:    false,
			Error:      err.Error(),
			DurationMs: time.Since(start).Milliseconds(),
		}
	}

	result := InvocationResult{
		ToolName:   toolName,
		Success:    resp.Success,
		Output:     resp.Data,
		Error:      resp.Error,
		DurationMs: time.Since(start).Milliseconds(),
		TokensUsed: 0, // TODO: Extract from response metadata
	}

	// Cache successful results
	if resp.Success {
		i.cache[toolName] = result
	}

	return result
}
