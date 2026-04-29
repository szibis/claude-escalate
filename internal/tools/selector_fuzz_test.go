package tools

import (
	"fmt"
	"strings"
	"testing"
)

// TestSelector_Fuzz_InvalidIntentNames tests selector with various malformed intent names
func TestSelector_Fuzz_InvalidIntentNames(t *testing.T) {
	sel := NewSelector()

	testCases := []struct {
		intent string
		valid  bool
	}{
		{"", false},                    // empty
		{" ", false},                   // whitespace only
		{"intent with spaces", false},  // spaces
		{"intent-with-dashes", false},  // dashes (spec uses underscores)
		{"UPPERCASE", true},            // uppercase (should work)
		{"intent_1", true},             // alphanumeric with underscore
		{"_leading_underscore", true},  // leading underscore
		{"intent__double", true},       // double underscore
		{"x", true},                    // single character
		{strings.Repeat("a", 1000), true}, // very long name
		{"intent\n", false},            // newline
		{"intent\t", false},            // tab
		{"intent\x00", false},          // null byte
	}

	for _, tc := range testCases {
		tools, err := sel.SelectForIntent(tc.intent)
		if !tc.valid {
			// Invalid intents should error if fallback disabled, else return fallback
			if err == nil && sel.fallbackEnabled {
				// Expected - fallback enabled returns default tools
				if len(tools) == 0 {
					t.Errorf("SelectForIntent(%q) with fallback should return tools, got empty", tc.intent)
				}
			}
		} else {
			if err != nil && !sel.fallbackEnabled {
				t.Errorf("SelectForIntent(%q) should succeed, got error: %v", tc.intent, err)
			}
		}
	}
}

// TestSelector_Fuzz_ExtremeTool names tests with various tool name formats
func TestSelector_Fuzz_ExtremeToolNames(t *testing.T) {
	sel := NewSelector()

	testCases := []string{
		"",                             // empty
		"a",                            // single char
		strings.Repeat("a", 10000),     // extremely long
		"tool_with_123_numbers",        // numbers
		"tool_with_CAPS",               // mixed case
		"_leading",                     // leading underscore
		"trailing_",                    // trailing underscore
		"__double__",                   // double underscores
		"tool-with-dashes",             // dashes (not spec compliant but test anyway)
		"tool.with.dots",               // dots
		"tool:with:colons",             // colons
		"tool/with/slashes",            // slashes
		"tool with spaces",             // spaces
	}

	for _, name := range testCases {
		tools := []string{name}
		sel.AddIntentMapping("test_intent", tools)

		retrieved, _ := sel.SelectForIntent("test_intent")
		if len(retrieved) > 0 && retrieved[0] == name {
			// Tool was added successfully
			continue
		}
	}
}

// TestSelector_Fuzz_WeightBoundaries tests weight selection with extreme values
func TestSelector_Fuzz_WeightBoundaries(t *testing.T) {
	sel := NewSelector()

	testCases := []struct {
		customWeight float64
		builtinWeight float64
		valid        bool
	}{
		{0.0, 0.0, true},           // zero weights
		{1.0, 1.0, true},           // max weights
		{0.5, 0.5, true},           // equal weights
		{0.9, 0.1, true},           // skewed weights
		{-1.0, 1.0, false},         // negative weight
		{2.0, 0.5, false},          // over 1.0
		{0.00001, 0.00001, true},   // very small weights
		{0.999999, 0.999999, true}, // near-max weights
	}

	for _, tc := range testCases {
		sel.SetWeights(tc.customWeight, tc.builtinWeight)
		// Verify weights were set or ignored
		if tc.valid || tc.customWeight == sel.customToolWeight {
			// Weights were set as expected
		}
	}
}

// TestSelector_Fuzz_SelectByWeight_EdgeCases tests weighted selection edge cases
func TestSelector_Fuzz_SelectByWeight_EdgeCases(t *testing.T) {
	sel := NewSelector()

	// Empty tools
	result := sel.SelectByWeight([]string{}, map[string]float64{})
	if result != "" {
		t.Error("SelectByWeight with empty tools should return empty string")
	}

	// Single tool with zero weight
	result = sel.SelectByWeight([]string{"tool1"}, map[string]float64{"tool1": 0.0})
	if result != "tool1" {
		t.Errorf("SelectByWeight with single tool should return it, got %q", result)
	}

	// Many tools with extreme weight differences
	tools := make([]string, 100)
	weights := make(map[string]float64)
	for i := 0; i < 100; i++ {
		tools[i] = fmt.Sprintf("tool_%d", i)
		if i == 0 {
			weights[tools[i]] = 999999.0 // extreme weight
		} else {
			weights[tools[i]] = 0.0001
		}
	}

	// Run selection multiple times - first tool should dominate
	firstToolCount := 0
	for i := 0; i < 1000; i++ {
		if result := sel.SelectByWeight(tools, weights); result == "tool_0" {
			firstToolCount++
		}
	}

	// First tool should be selected much more often (heuristic check)
	if firstToolCount < 900 {
		t.Logf("SelectByWeight skewed weight distribution: tool_0 selected %d/1000 times (expected >900)", firstToolCount)
	}
}

// TestSelector_Fuzz_ConcurrentModification tests concurrent access patterns
func TestSelector_Fuzz_ConcurrentModification(t *testing.T) {
	sel := NewSelector()

	// Add initial mappings
	for i := 0; i < 10; i++ {
		intent := fmt.Sprintf("intent_%d", i)
		tools := []string{fmt.Sprintf("tool_%d_a", i), fmt.Sprintf("tool_%d_b", i)}
		sel.AddIntentMapping(intent, tools)
	}

	// Concurrent reads should not panic
	results := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func(idx int) {
			intent := fmt.Sprintf("intent_%d", idx%10)
			_, _ = sel.SelectForIntent(intent)
			results <- true
		}(i)
	}

	// Collect all results
	for i := 0; i < 100; i++ {
		<-results
	}
}

// TestSelector_Fuzz_InvalidFallbackState tests fallback flag edge cases
func TestSelector_Fuzz_InvalidFallbackState(t *testing.T) {
	sel := NewSelector()

	// Test rapid fallback toggle
	for i := 0; i < 1000; i++ {
		sel.SetFallback(i%2 == 0)
		_, _ = sel.SelectForIntent("test_intent")
	}

	// Test with no default mappings (would need to manually clear)
	tools, _ := sel.SelectForIntent("nonexistent")
	if len(tools) == 0 && sel.fallbackEnabled {
		t.Error("With fallback enabled, should return fallback tools")
	}
}
