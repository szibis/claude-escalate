package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/szibis/claude-escalate/internal/config"
	"github.com/szibis/claude-escalate/internal/dashboard"
	"github.com/szibis/claude-escalate/internal/metrics"
)

func main() {
	loader := config.NewLoader("")
	collector := metrics.NewMetricsCollector()
	publisher := metrics.NewMetricsPublisher(collector, 1*time.Minute)
	s := dashboard.NewServer("127.0.0.1", 8077, loader, collector, publisher)

	reqBody := `{"name":"test_tool","type":"cli","path":"/usr/bin/test","settings":{}}`
	req := httptest.NewRequest("POST", "/api/tools/add", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.HandleToolsAdd(w, req)

	body, _ := io.ReadAll(w.Body)
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Body: %s\n", body)
}
