package tools

import (
	"strings"
	"testing"
)

// TestToolRegistry_Fuzz_EdgeCases tests ToolRegistry creation edge cases
func TestToolRegistry_Fuzz_StoragePathEdgeCases(t *testing.T) {
	testPaths := []string{
		"",
		"/",
		"/tmp/config.yaml",
		"/tmp/config with spaces.yaml",
		"/tmp/config\nwith\nnewlines.yaml",
		"/path/with/../traversal/config.yaml",
		strings.Repeat("/a", 1000) + "/config.yaml",
	}

	for _, path := range testPaths {
		// Should not panic
		reg := NewToolRegistry(path)
		if reg == nil {
			t.Errorf("Failed to create ToolRegistry with path: %q", path)
		}

		// Verify path is stored
		if reg.configPath != path {
			t.Errorf("Path mismatch: stored %q, want %q", reg.configPath, path)
		}
	}
}

// TestToolRegistry_Fuzz_ListToolsEdgeCases tests ListTools with various registry states
func TestToolRegistry_Fuzz_ListToolsEdgeCases(t *testing.T) {
	reg := NewToolRegistry("/tmp/test.yaml")

	// Empty registry
	tools := reg.ListTools()
	if tools == nil {
		t.Error("ListTools should return empty slice, not nil")
	}

	if len(tools) != 0 {
		t.Errorf("Empty registry should have 0 tools, got %d", len(tools))
	}
}

// TestToolRegistry_Fuzz_GetToolEdgeCases tests GetTool with various inputs
func TestToolRegistry_Fuzz_GetToolEdgeCases(t *testing.T) {
	reg := NewToolRegistry("/tmp/test.yaml")

	testNames := []string{
		"",
		" ",
		"nonexistent",
		"tool\x00null",
		"tool\nwith\nnewline",
		strings.Repeat("a", 10000),
	}

	for _, name := range testNames {
		tool, err := reg.GetTool(name)

		// GetTool may or may not find the tool, but shouldn't panic
		if tool != nil && name != tool.Name {
			t.Errorf("Tool name mismatch: expected %q, got %q", name, tool.Name)
		}

		_ = err // err may be set or not
	}
}

// TestToolRegistry_Fuzz_HealthCheckBehavior tests health check initialization
func TestToolRegistry_Fuzz_HealthCheckBehavior(t *testing.T) {
	reg := NewToolRegistry("/tmp/test.yaml")

	// Should initialize with sensible defaults
	if !reg.healthCheckOnAdd {
		t.Error("Health check should be enabled by default")
	}

	if !reg.validationOnAdd {
		t.Error("Validation should be enabled by default")
	}

	if !reg.monitoringEnabled {
		t.Error("Monitoring should be enabled by default")
	}
}

// TestToolRegistry_Fuzz_AuditTrailInitialization tests audit trail setup
func TestToolRegistry_Fuzz_AuditTrailInitialization(t *testing.T) {
	reg := NewToolRegistry("/tmp/test.yaml")

	// Audit trail should be initialized
	if !reg.auditTrail {
		t.Error("Audit trail should be enabled by default")
	}

	// GetAuditLog should return valid slice (takes limit parameter)
	log := reg.GetAuditLog(100)
	if log == nil {
		t.Error("GetAuditLog should return slice, not nil")
	}

	// Initially empty
	if len(log) != 0 {
		t.Errorf("New registry should have empty audit log, got %d entries", len(log))
	}
}

// TestToolRegistry_Fuzz_ConfigPathVariations tests configuration path edge cases
func TestToolRegistry_Fuzz_ConfigPathVariations(t *testing.T) {
	paths := []string{
		"",
		"/",
		"/tmp/config.yaml",
		"/tmp/config with spaces.yaml",
		"/tmp/config\nwith\nnewlines.yaml",
		"/very/" + strings.Repeat("long/", 500) + "config.yaml",
		"~/home/relative/path",
		"./local/relative/path",
	}

	for _, path := range paths {
		reg := NewToolRegistry(path)

		// Should handle all paths without panic
		_ = reg.ListTools()
		log := reg.GetAuditLog(100)
		if log == nil {
			t.Error("GetAuditLog returned nil")
		}
	}
}

// TestToolRegistry_Fuzz_UnicodeConfigPath tests unicode in config paths
func TestToolRegistry_Fuzz_UnicodeConfigPath(t *testing.T) {
	paths := []string{
		"/tmp/你好.yaml",
		"/tmp/مرحبا.yaml",
		"/tmp/🎉config.yaml",
		"/tmp/café/config.yaml",
	}

	for _, path := range paths {
		reg := NewToolRegistry(path)
		if reg == nil {
			t.Errorf("Failed to create registry with path: %q", path)
		}

		// Verify path preserved with unicode intact
		if reg.configPath != path {
			t.Errorf("Unicode path not preserved: got %q, want %q", reg.configPath, path)
		}
	}
}

// TestToolRegistry_Fuzz_RemoveToolEdgeCases tests RemoveTool with various inputs
func TestToolRegistry_Fuzz_RemoveToolEdgeCases(t *testing.T) {
	reg := NewToolRegistry("/tmp/test.yaml")

	// Remove from empty registry
	reg.RemoveTool("nonexistent")
	reg.RemoveTool("")
	reg.RemoveTool(" ")

	// Should not panic
	if len(reg.ListTools()) != 0 {
		t.Error("Registry should still be empty")
	}

	// Remove again
	reg.RemoveTool("nonexistent")

	// Should still be fine
	_ = reg.ListTools()
}

// TestToolRegistry_Fuzz_MetadataFieldVariations tests ToolMetadata variations
func TestToolRegistry_Fuzz_MetadataFieldVariations(t *testing.T) {
	// Create registries with various path formats
	registries := []*ToolRegistry{
		NewToolRegistry(""),
		NewToolRegistry("/"),
		NewToolRegistry("/tmp"),
		NewToolRegistry("/tmp/config.yaml"),
		NewToolRegistry(strings.Repeat("a", 10000) + ".yaml"),
	}

	for _, reg := range registries {
		if reg == nil {
			t.Error("Failed to create registry")
			continue
		}

		// List tools should work
		tools := reg.ListTools()
		if tools == nil {
			t.Error("ListTools returned nil")
		}

		// GetAuditLog should work
		log := reg.GetAuditLog(100)
		if log == nil {
			t.Error("GetAuditLog returned nil")
		}
	}
}
