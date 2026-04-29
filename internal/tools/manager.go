package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/szibis/claude-escalate/internal/config"
	"github.com/szibis/claude-escalate/internal/gateway"
)

// ToolManager is the main orchestrator for tool operations
type ToolManager struct {
	registry       *ToolRegistry
	invoker        *Invoker
	selector       *Selector
	errorHandler   *ErrorHandler
	processor      *OutputProcessor
	factory        *gateway.AdapterFactory
}

// NewToolManager creates a new tool manager
func NewToolManager(
	registry *ToolRegistry,
	factory *gateway.AdapterFactory,
) *ToolManager {
	selector := NewSelector()
	invoker := NewInvoker(factory, selector)
	errorHandler := NewErrorHandler(ErrorPolicy{
		FailOpen:       true,
		LogAllErrors:   true,
		AlertThreshold: 5,
	})
	processor := NewOutputProcessor()

	return &ToolManager{
		registry:     registry,
		invoker:      invoker,
		selector:     selector,
		errorHandler: errorHandler,
		processor:    processor,
		factory:      factory,
	}
}

// InvokeForIntent orchestrates tool invocation for a given intent
func (tm *ToolManager) InvokeForIntent(ctx context.Context, intent, input string) (string, error) {
	req := &InvocationRequest{
		Intent:      intent,
		Input:       input,
		MaxDuration: 30 * time.Second,
		MaxResults:  3,
		TokenBudget: 3000,
	}

	// Invoke tools
	results, err := tm.invoker.Invoke(ctx, req)
	if err != nil {
		if tm.errorHandler.policy.FailOpen {
			// Return empty context instead of error
			return "", nil
		}
		return "", err
	}

	// Process results for context injection
	contextStr, tokens, err := tm.processor.ProcessResults(results)
	if err != nil && !tm.errorHandler.policy.FailOpen {
		return "", err
	}

	fmt.Printf("Tool invocation complete: %d tokens used from %d budget\n", tokens, req.TokenBudget)
	return contextStr, nil
}

// RegisterCustomTool registers a new custom tool
func (tm *ToolManager) RegisterCustomTool(ctx context.Context, tool *ToolMetadata) error {
	if err := tm.registry.RegisterTool(ctx, tool); err != nil {
		return err
	}

	// Persist to config
	cfg := config.DefaultConfig()
	loader := config.NewLoader("")
	if loadedCfg, err := loader.Load(); err == nil {
		cfg = loadedCfg
	}

	return tm.registry.SaveToConfig(cfg)
}

// UpdateTool updates an existing tool configuration
func (tm *ToolManager) UpdateTool(ctx context.Context, name string, updates *ToolMetadata) error {
	if err := tm.registry.UpdateTool(ctx, name, updates); err != nil {
		return err
	}

	// Persist to config
	cfg := config.DefaultConfig()
	loader := config.NewLoader("")
	if loadedCfg, err := loader.Load(); err == nil {
		cfg = loadedCfg
	}

	return tm.registry.SaveToConfig(cfg)
}

// RemoveTool removes a tool from the registry and config
func (tm *ToolManager) RemoveTool(name string) error {
	if err := tm.registry.RemoveTool(name); err != nil {
		return err
	}

	// Persist to config
	cfg := config.DefaultConfig()
	loader := config.NewLoader("")
	if loadedCfg, err := loader.Load(); err == nil {
		cfg = loadedCfg
	}

	return tm.registry.SaveToConfig(cfg)
}

// GetToolStatus returns the current status of a tool
func (tm *ToolManager) GetToolStatus(name string) (map[string]interface{}, error) {
	tool, err := tm.registry.GetTool(name)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":                 tool.Name,
		"type":                 tool.Type,
		"status":               tool.HealthStatus,
		"enabled":              tool.Enabled,
		"registered_at":        tool.RegisteredAt,
		"last_health_check":    tool.LastHealthCheck,
		"consecutive_errors":   tool.ConsecutiveErrors,
	}, nil
}

// ListAllTools returns all registered tools with their status
func (tm *ToolManager) ListAllTools() []map[string]interface{} {
	tools := tm.registry.ListTools()
	result := make([]map[string]interface{}, 0, len(tools))

	for _, tool := range tools {
		result = append(result, map[string]interface{}{
			"name":               tool.Name,
			"type":               tool.Type,
			"status":             tool.HealthStatus,
			"enabled":            tool.Enabled,
			"consecutive_errors": tool.ConsecutiveErrors,
		})
	}

	return result
}

// StartHealthMonitoring begins periodic health checks
func (tm *ToolManager) StartHealthMonitoring(ctx context.Context) {
	tm.registry.StartMonitoring(ctx)
}

// GetAuditLog returns recent tool operation logs
func (tm *ToolManager) GetAuditLog(limit int) []AuditEntry {
	return tm.registry.GetAuditLog(limit)
}

// UpdateIntentToolMapping updates which tools are used for an intent
func (tm *ToolManager) UpdateIntentToolMapping(intent string, tools []string) {
	tm.selector.AddIntentMapping(intent, tools)
}

// SetToolWeights updates the preferences for custom vs builtin tools
func (tm *ToolManager) SetToolWeights(customWeight, builtinWeight float64) {
	tm.selector.SetWeights(customWeight, builtinWeight)
}

// LoadFromConfig initializes the manager from a config file
func (tm *ToolManager) LoadFromConfig(cfg *config.Config) error {
	return tm.registry.LoadFromConfig(cfg)
}
