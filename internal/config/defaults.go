package config

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/szibis/claude-escalate/internal/discovery"
)

// DefaultConfig returns a base configuration with auto-detected tools
func DefaultConfig() *Config {
	cfg := &Config{
		Gateway: GatewayConfig{
			Port:            9000,
			Host:            "0.0.0.0",
			SecurityLayer:   true,
			ShutdownTimeout: 30,
			MaxRequestSize:  10485760, // 10MB
			DataDir:         expandHome("~/.claude/data/escalate"),
		},
		Optimizations: OptimizationsConfig{
			RTK: RTKConfig{
				Enabled:             isToolAvailable("rtk"),
				CommandProxySavings: 99.4,
				CacheSavings:        true,
			},
			MCP: MCPConfig{
				Enabled: true,
				Tools:   detectAndCreateTools(),
			},
			SemanticCache: SemanticCacheConfig{
				Enabled:             true,
				EmbeddingModel:      "text-embedding-3-small",
				SimilarityThreshold: 0.85,
				HitRateTarget:       0.90,
				FalsePositiveLimit:  0.01,
				MaxCacheSize:        500,
			},
			KnowledgeGraph: KnowledgeGraphConfig{
				Enabled:         true,
				IndexLocalCode:  true,
				IndexWebContent: true,
				CacheLookups:    true,
				DBPath:          expandHome("~/.claude/data/escalate/kg.db"),
			},
			InputOptimization: InputOptimizationConfig{
				Enabled:                  true,
				StripUnusedTools:         true,
				CompressParameters:       true,
				DeduplicateExactRequests: true,
			},
			OutputOptimization: OutputOptimizationConfig{
				Enabled:             true,
				ResponseCompression: true,
				FieldFiltering:      true,
				DeltaDetection:      true,
			},
			BatchAPI: BatchAPIConfig{
				Enabled: true,
			},
		},
		Security: SecurityConfig{
			Enabled:                  true,
			MaxRatePerMinute:         1000,
			BlockSuspiciousPatterns:  true,
			RequireSecureConnection:  false,
			AllowedOrigins:           []string{"localhost", "127.0.0.1"},
		},
		Metrics: MetricsConfig{
			Enabled:           true,
			RetentionDays:     30,
			SamplingInterval:  1,
		},
	}

	return cfg
}

// detectAndCreateTools auto-detects installed tools and creates MCPTool entries
func detectAndCreateTools() []MCPTool {
	tools := discovery.DetectTools()
	var mcpTools []MCPTool

	// RTK
	if tools.RTKPath != "" {
		mcpTools = append(mcpTools, MCPTool{
			Type: "cli",
			Name: "rtk",
			Settings: map[string]interface{}{
				"path":        tools.RTKPath,
				"description": "Real Token Killer - Command output optimization (99.4% savings)",
				"enabled":     true,
			},
		})
	}

	// Scrapling
	if tools.ScraplingPath != "" {
		mcpTools = append(mcpTools, MCPTool{
			Type: "mcp",
			Name: "scrapling",
			Settings: map[string]interface{}{
				"path":        tools.ScraplingPath,
				"description": "Web scraping and content extraction (85-94% savings)",
				"enabled":     true,
			},
		})
	}

	// Git
	if tools.GitPath != "" {
		mcpTools = append(mcpTools, MCPTool{
			Type: "cli",
			Name: "git",
			Settings: map[string]interface{}{
				"path":        tools.GitPath,
				"description": "Version control and diff operations",
				"enabled":     true,
			},
		})
	}

	// LSP Servers
	for lang, path := range tools.LSPServers {
		mcpTools = append(mcpTools, MCPTool{
			Type: "lsp",
			Name: lang + "-lsp",
			Settings: map[string]interface{}{
				"path":        path,
				"language":    lang,
				"description": "Language server for " + lang + " (code navigation)",
				"enabled":     true,
			},
		})
	}

	return mcpTools
}

// isToolAvailable checks if a tool is available in PATH
func isToolAvailable(toolName string) bool {
	_, err := exec.LookPath(toolName)
	return err == nil
}

// expandHome expands ~ to home directory
func expandHome(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[1:])
		}
	}
	return path
}
