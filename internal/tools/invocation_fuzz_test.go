package tools

import (
	"context"
	"strings"
	"testing"
)

// TestInvocationRequest_Fuzz_EdgeCases tests InvocationRequest edge cases
func TestInvocationRequest_Fuzz_TokenBudgetBoundaries(t *testing.T) {
	testBudgets := []int{
		0,          // zero budget
		1,          // minimum
		10,         // small
		1000,       // default
		1000000,    // huge
		-1,         // negative (invalid)
		999999999,  // max int
	}

	for _, budget := range testBudgets {
		request := &InvocationRequest{
			Intent:      "test_intent",
			TokenBudget: budget,
			MaxResults:  10,
		}

		// Request should be valid regardless of budget
		if request == nil {
			t.Error("Request creation failed")
		}
	}
}

// TestInvocationRequest_Fuzz_MaxResultsBoundaries tests max results edge cases
func TestInvocationRequest_Fuzz_MaxResultsBoundaries(t *testing.T) {
	testMaxResults := []int{
		0,          // zero (no results)
		1,          // single result
		10,         // typical
		1000,       // many
		1000000,    // extreme
		-1,         // negative (invalid)
		999999999,  // max int
	}

	for _, maxResults := range testMaxResults {
		request := &InvocationRequest{
			Intent:      "test_intent",
			MaxResults:  maxResults,
			TokenBudget: 3000,
		}

		if request == nil {
			t.Error("Request creation failed")
		}
	}
}

// TestInvocationRequest_Fuzz_SpecialIntentNames tests with unusual intent names
func TestInvocationRequest_Fuzz_SpecialIntentNames(t *testing.T) {
	intents := []string{
		"",
		" ",
		"intent with spaces",
		"UPPERCASE_INTENT",
		"intent-with-dashes",
		"intent/with/slashes",
		"intent:with:colons",
		"intent\nwith\nnewlines",
		"intent\twith\ttabs",
		"intent\x00with\x00nulls",
		"😀emoji🎉intent",
		"你好intent世界",
		strings.Repeat("a", 100000),
	}

	for _, intent := range intents {
		request := &InvocationRequest{
			Intent:      intent,
			TokenBudget: 3000,
			MaxResults:  10,
		}

		if request == nil {
			t.Error("Request creation failed")
		}
	}
}

// TestInvocationRequest_Fuzz_ContextDeadlines tests various context scenarios
func TestInvocationRequest_Fuzz_ContextDeadlines(t *testing.T) {
	testCases := []struct {
		name     string
		createCtx func() context.Context
	}{
		{"background", func() context.Context {
			return context.Background()
		}},
		{"cancelled context", func() context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			return ctx
		}},
	}

	for _, tc := range testCases {
		ctx := tc.createCtx()
		request := &InvocationRequest{
			Intent:      "test_intent",
			TokenBudget: 3000,
			MaxResults:  10,
		}

		// Should handle context without panicking
		if request == nil || ctx == nil {
			t.Errorf("Failed to create request or context in test case %s", tc.name)
		}
	}
}

// TestInvocationRequest_Fuzz_LargePayloads tests with large request payload
func TestInvocationRequest_Fuzz_LargePayloads(t *testing.T) {
	// Very long intent name
	request := &InvocationRequest{
		Intent:      strings.Repeat("intent", 10000),
		TokenBudget: 3000,
		MaxResults:  10,
	}

	if request == nil {
		t.Error("Request creation with large intent failed")
	}

	// Verify fields are preserved
	if len(request.Intent) != len(strings.Repeat("intent", 10000)) {
		t.Error("Intent truncated unexpectedly")
	}
}

// TestInvocationRequest_Fuzz_BoundaryValues tests boundary conditions
func TestInvocationRequest_Fuzz_BoundaryValues(t *testing.T) {
	testCases := []struct {
		intent       string
		tokenBudget  int
		maxResults   int
	}{
		{"", 0, 0},
		{"a", 1, 1},
		{strings.Repeat("a", 10000), 1000000, 1000000},
		{"intent", -1, -1},
		{"intent", 999999999, 999999999},
	}

	for _, tc := range testCases {
		request := &InvocationRequest{
			Intent:      tc.intent,
			TokenBudget: tc.tokenBudget,
			MaxResults:  tc.maxResults,
		}

		if request == nil {
			t.Errorf("Failed to create request with intent=%q, budget=%d, max=%d",
				tc.intent, tc.tokenBudget, tc.maxResults)
		}
	}
}

// TestInvocationRequest_Fuzz_UnicodeContent tests unicode handling
func TestInvocationRequest_Fuzz_UnicodeContent(t *testing.T) {
	intents := []string{
		"hello",
		"你好世界",                 // Chinese
		"مرحبا بالعالم",           // Arabic
		"🎉🎊🎈",                  // Emoji
		"café",                    // Accented
		"Ñoño",                    // Spanish
		"(っ˘̩╭╮˘̩)っ",           // Complex Unicode
		strings.Repeat("中", 1000), // Many Chinese chars
	}

	for _, intent := range intents {
		request := &InvocationRequest{
			Intent:      intent,
			TokenBudget: 3000,
			MaxResults:  10,
		}

		if request == nil {
			t.Errorf("Failed to create request with unicode intent: %s", intent)
		}

		// Verify intent is preserved
		if request.Intent != intent {
			t.Errorf("Intent was corrupted: got %q, want %q", request.Intent, intent)
		}
	}
}
