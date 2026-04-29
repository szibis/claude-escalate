package tools

import (
	"encoding/json"
	"fmt"
	"strings"
)

// OutputProcessor handles tool output formatting and context injection
type OutputProcessor struct {
	maxContextTokens  int
	stripANSI         bool
	parseJSON         bool
	truncateOnLimit   bool
	tokenBudget       int
	perToolLimit      int
}

// NewOutputProcessor creates a new output processor
func NewOutputProcessor() *OutputProcessor {
	return &OutputProcessor{
		maxContextTokens: 2000,
		stripANSI:        true,
		parseJSON:        true,
		truncateOnLimit:  true,
		tokenBudget:      3000,
		perToolLimit:     1000,
	}
}

// ProcessOutput processes and formats tool output for context injection
func (op *OutputProcessor) ProcessOutput(toolName string, output interface{}, metadata map[string]interface{}) (string, int, error) {
	// Step 1: Convert to string
	outputStr := op.toString(output)

	// Step 2: Clean output
	outputStr = op.cleanOutput(outputStr)

	// Step 3: Estimate tokens (rough estimate: 1 token ≈ 4 chars)
	estimatedTokens := len(outputStr) / 4
	if estimatedTokens > op.perToolLimit {
		if !op.truncateOnLimit {
			return "", 0, fmt.Errorf("output exceeds limit: %d tokens > %d", estimatedTokens, op.perToolLimit)
		}
		// Truncate to limit
		maxChars := op.perToolLimit * 4
		if len(outputStr) > maxChars {
			outputStr = outputStr[:maxChars] + "\n[output truncated...]"
		}
		estimatedTokens = op.perToolLimit
	}

	// Step 4: Format for injection
	formatted := op.formatForContext(toolName, outputStr, metadata, estimatedTokens)

	return formatted, estimatedTokens, nil
}

// toString converts output to string
func (op *OutputProcessor) toString(output interface{}) string {
	switch v := output.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case map[string]interface{}:
		if op.parseJSON {
			data, _ := json.MarshalIndent(v, "", "  ")
			return string(data)
		}
		return fmt.Sprintf("%v", v)
	default:
		if op.parseJSON {
			data, _ := json.MarshalIndent(v, "", "  ")
			return string(data)
		}
		return fmt.Sprintf("%v", v)
	}
}

// cleanOutput removes ANSI codes and control characters
func (op *OutputProcessor) cleanOutput(output string) string {
	if op.stripANSI {
		output = stripANSICodes(output)
	}
	output = removeControlChars(output)
	return strings.TrimSpace(output)
}

// formatForContext injects output into markdown format for context
func (op *OutputProcessor) formatForContext(toolName string, output string, metadata map[string]interface{}, tokens int) string {
	var sb strings.Builder

	// Header with tool name
	sb.WriteString(fmt.Sprintf("```tool:%s\n", toolName))

	// Optional metadata
	if len(metadata) > 0 {
		if execTime, ok := metadata["execution_time_ms"].(int64); ok {
			sb.WriteString(fmt.Sprintf("# Execution: %dms\n", execTime))
		}
		if cached, ok := metadata["cached"].(bool); ok && cached {
			sb.WriteString("# (cached result)\n")
		}
	}

	// Output
	sb.WriteString(output)

	// Footer with token info
	sb.WriteString(fmt.Sprintf("\n# (~%d tokens)\n", tokens))
	sb.WriteString("```\n")

	return sb.String()
}

// ProcessResults processes multiple tool results and injects into context
func (op *OutputProcessor) ProcessResults(results []InvocationResult) (string, int, error) {
	if len(results) == 0 {
		return "", 0, nil
	}

	var sb strings.Builder
	totalTokens := 0

	sb.WriteString("## Tool Results\n\n")

	for _, result := range results {
		if !result.Success {
			sb.WriteString(fmt.Sprintf("### ❌ %s\nError: %s\n\n", result.ToolName, result.Error))
			continue
		}

		// Process each successful result
		metadata := map[string]interface{}{
			"execution_time_ms": result.DurationMs,
			"cached":            result.Cached,
		}

		formatted, tokens, err := op.ProcessOutput(result.ToolName, result.Output, metadata)
		if err != nil {
			sb.WriteString(fmt.Sprintf("### ⚠️ %s\nWarning: %v\n\n", result.ToolName, err))
			continue
		}

		sb.WriteString(fmt.Sprintf("### ✅ %s\n%s\n", result.ToolName, formatted))
		totalTokens += tokens

		// Check budget
		if totalTokens > op.tokenBudget {
			sb.WriteString(fmt.Sprintf("\n⚠️ Token budget exceeded (%d > %d). Stopping tool output injection.\n", totalTokens, op.tokenBudget))
			break
		}
	}

	return sb.String(), totalTokens, nil
}

// stripANSICodes removes ANSI color and formatting codes
func stripANSICodes(s string) string {
	ansiRegex := "\\x1b\\[[0-9;]*m"
	// Simple implementation: remove common ANSI codes
	s = strings.ReplaceAll(s, "\x1b[0m", "")
	s = strings.ReplaceAll(s, "\x1b[1m", "")
	s = strings.ReplaceAll(s, "\x1b[2m", "")
	s = strings.ReplaceAll(s, "\x1b[3m", "")
	s = strings.ReplaceAll(s, "\x1b[4m", "")
	s = strings.ReplaceAll(s, "\x1b[7m", "")
	// Remove color codes (30-37 foreground, 40-47 background)
	for i := 30; i <= 47; i++ {
		s = strings.ReplaceAll(s, fmt.Sprintf("\x1b[%dm", i), "")
	}
	_ = ansiRegex // Suppress unused variable warning
	return s
}

// removeControlChars removes non-printable control characters
func removeControlChars(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if r >= 32 || r == '\n' || r == '\r' || r == '\t' {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
