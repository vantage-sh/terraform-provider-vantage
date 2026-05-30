// Mock Vantage API server for testing business metric reordering scenarios.
// Intentionally returns cost_report_tokens_with_metadata in a different order
// than submitted to reproduce the reordering bug.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var reverseOrder = flag.Bool("reverse", true, "Reverse the order of cost_report_tokens in responses")
var shuffleOrder = flag.Bool("shuffle", false, "Shuffle the order of cost_report_tokens (rotate by 1)")
var port = flag.Int("port", 9090, "Port to listen on")

type CostReportToken struct {
	CostReportToken string   `json:"cost_report_token"`
	UnitScale       string   `json:"unit_scale"`
	LabelFilter     []string `json:"label_filter"`
}

type BusinessMetricValue struct {
	Amount float64 `json:"amount"`
	Date   string  `json:"date"`
	Label  string  `json:"label,omitempty"`
}

type BusinessMetric struct {
	Token                        string            `json:"token"`
	Title                        string            `json:"title"`
	CostReportTokensWithMetadata []CostReportToken `json:"cost_report_tokens_with_metadata"`
	CreatedByToken               *string           `json:"created_by_token,omitempty"`
	ImportType                   *string           `json:"import_type"`
	IntegrationToken             *string           `json:"integration_token"`
	Values                       []BusinessMetricValue `json:"values,omitempty"`
	ForecastedValues             []BusinessMetricValue `json:"forecasted_values,omitempty"`
}

type CreateBusinessMetric struct {
	Title                        string            `json:"title"`
	CostReportTokensWithMetadata []CostReportToken `json:"cost_report_tokens_with_metadata"`
	Values                       []BusinessMetricValue `json:"values,omitempty"`
	ForecastedValues             []BusinessMetricValue `json:"forecasted_values,omitempty"`
}

type BusinessMetricResponse struct {
	Token                        string            `json:"token"`
	Title                        string            `json:"title"`
	CostReportTokensWithMetadata []CostReportToken `json:"cost_report_tokens_with_metadata"`
	CreatedByToken               *string           `json:"created_by_token"`
	ImportType                   *string           `json:"import_type"`
	IntegrationToken             *string           `json:"integration_token"`
}

type BusinessMetricsResponse struct {
	BusinessMetrics []BusinessMetricResponse `json:"business_metrics"`
	Links           Links                    `json:"links"`
}

type Links struct {
	Next *string `json:"next"`
	Prev *string `json:"prev"`
}

var (
	mu       sync.Mutex
	metrics  = make(map[string]*BusinessMetric)
	counter  int
)

func newToken() string {
	counter++
	return fmt.Sprintf("bsnss_mtrc_%04d", counter)
}

func reorderTokens(tokens []CostReportToken) []CostReportToken {
	if len(tokens) == 0 {
		return tokens
	}

	result := make([]CostReportToken, len(tokens))

	if *reverseOrder {
		// Reverse the slice
		for i, t := range tokens {
			result[len(tokens)-1-i] = t
		}
		log.Printf("  Reordering: reversed %d cost_report_tokens", len(tokens))
	} else if *shuffleOrder {
		// Rotate by 1 (last becomes first)
		result[0] = tokens[len(tokens)-1]
		copy(result[1:], tokens[:len(tokens)-1])
		log.Printf("  Reordering: rotated %d cost_report_tokens", len(tokens))
	} else {
		copy(result, tokens)
	}
	return result
}

func toResponse(m *BusinessMetric) BusinessMetricResponse {
	importType := "csv"
	tokens := reorderTokens(m.CostReportTokensWithMetadata)

	// Normalize label_filter: ensure it's never nil (use empty slice)
	for i := range tokens {
		if tokens[i].LabelFilter == nil {
			tokens[i].LabelFilter = []string{}
		}
	}

	return BusinessMetricResponse{
		Token:                        m.Token,
		Title:                        m.Title,
		CostReportTokensWithMetadata: tokens,
		CreatedByToken:               m.CreatedByToken,
		ImportType:                   &importType,
		IntegrationToken:             m.IntegrationToken,
	}
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func handleBusinessMetrics(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	log.Printf(">>> %s %s", r.Method, r.URL.Path)

	switch r.Method {
	case http.MethodPost:
		var create CreateBusinessMetric
		if err := json.NewDecoder(r.Body).Decode(&create); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		token := newToken()
		importType := "csv"
		m := &BusinessMetric{
			Token:                        token,
			Title:                        create.Title,
			CostReportTokensWithMetadata: create.CostReportTokensWithMetadata,
			ImportType:                   &importType,
			Values:                       create.Values,
			ForecastedValues:             create.ForecastedValues,
		}
		if m.CostReportTokensWithMetadata == nil {
			m.CostReportTokensWithMetadata = []CostReportToken{}
		}

		metrics[token] = m

		log.Printf("  Created %s: %q with %d cost_report_tokens, %d values",
			token, create.Title, len(m.CostReportTokensWithMetadata), len(m.Values))
		for i, t := range m.CostReportTokensWithMetadata {
			log.Printf("    [%d] %s (scale=%s, filter=%v)", i, t.CostReportToken, t.UnitScale, t.LabelFilter)
		}

		writeJSON(w, http.StatusCreated, toResponse(m))

	case http.MethodGet:
		// List all business metrics
		var list []BusinessMetricResponse
		for _, m := range metrics {
			list = append(list, toResponse(m))
		}
		if list == nil {
			list = []BusinessMetricResponse{}
		}
		writeJSON(w, http.StatusOK, BusinessMetricsResponse{
			BusinessMetrics: list,
			Links:           Links{},
		})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleBusinessMetric(w http.ResponseWriter, r *http.Request, token string) {
	mu.Lock()
	defer mu.Unlock()

	log.Printf(">>> %s %s", r.Method, r.URL.Path)

	switch r.Method {
	case http.MethodGet:
		m, ok := metrics[token]
		if !ok {
			http.Error(w, `{"errors":["not found"]}`, http.StatusNotFound)
			return
		}
		log.Printf("  GET %s: %q with %d cost_report_tokens", token, m.Title, len(m.CostReportTokensWithMetadata))
		resp := toResponse(m)
		log.Printf("  Returning in order:")
		for i, t := range resp.CostReportTokensWithMetadata {
			log.Printf("    [%d] %s (scale=%s)", i, t.CostReportToken, t.UnitScale)
		}
		writeJSON(w, http.StatusOK, resp)

	case http.MethodPut:
		m, ok := metrics[token]
		if !ok {
			http.Error(w, `{"errors":["not found"]}`, http.StatusNotFound)
			return
		}

		var update CreateBusinessMetric
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		m.Title = update.Title
		if update.CostReportTokensWithMetadata != nil {
			m.CostReportTokensWithMetadata = update.CostReportTokensWithMetadata
		}
		if update.Values != nil {
			m.Values = update.Values
		}
		if update.ForecastedValues != nil {
			m.ForecastedValues = update.ForecastedValues
		}

		log.Printf("  Updated %s: %q with %d cost_report_tokens",
			token, m.Title, len(m.CostReportTokensWithMetadata))
		for i, t := range m.CostReportTokensWithMetadata {
			log.Printf("    [%d] %s (scale=%s, filter=%v)", i, t.CostReportToken, t.UnitScale, t.LabelFilter)
		}

		writeJSON(w, http.StatusOK, toResponse(m))

	case http.MethodDelete:
		_, ok := metrics[token]
		if !ok {
			http.Error(w, `{"errors":["not found"]}`, http.StatusNotFound)
			return
		}
		delete(metrics, token)
		log.Printf("  Deleted %s", token)
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func router(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Remove /v2 prefix
	path = strings.TrimPrefix(path, "/v2")

	// Route: /business_metrics
	if path == "/business_metrics" {
		handleBusinessMetrics(w, r)
		return
	}

	// Route: /business_metrics/{token}
	if strings.HasPrefix(path, "/business_metrics/") {
		parts := strings.Split(strings.TrimPrefix(path, "/business_metrics/"), "/")
		token := parts[0]

		// Sub-routes for values/forecasted_values (not used by provider in CRUD)
		if len(parts) > 1 {
			// For completeness, return empty values
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"values": []interface{}{},
				"links":  map[string]interface{}{},
			})
			return
		}

		handleBusinessMetric(w, r, token)
		return
	}

	// Workspaces (needed by some data sources)
	if path == "/workspaces" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"workspaces": []map[string]interface{}{
				{"token": "wrkspc_test", "name": "Test Workspace"},
			},
			"links": map[string]interface{}{},
		})
		return
	}

	// Cost reports (minimal stub)
	if path == "/cost_reports" || strings.HasPrefix(path, "/cost_reports/") {
		handleCostReports(w, r, path)
		return
	}

	log.Printf("404: %s %s", r.Method, r.URL.Path)
	http.Error(w, `{"errors":["not found"]}`, http.StatusNotFound)
}

// Cost report stubs
var (
	costReports    = make(map[string]map[string]interface{})
	reportCounter  int
)

func handleCostReports(w http.ResponseWriter, r *http.Request, path string) {
	if path == "/cost_reports" {
		switch r.Method {
		case http.MethodPost:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			mu.Lock()
			reportCounter++
			token := fmt.Sprintf("rprt_%04d", reportCounter)
			report := map[string]interface{}{
				"token":            token,
				"title":            body["title"],
				"workspace_token":  body["workspace_token"],
				"date_interval":    body["date_interval"],
				"filter":           body["filter"],
				"created_at":       time.Now().Format(time.RFC3339),
				"updated_at":       time.Now().Format(time.RFC3339),
				"groupings":        []string{},
				"business_metric_tokens_with_metadata": []interface{}{},
			}
			costReports[token] = report
			mu.Unlock()
			log.Printf("  Created cost_report %s: %q", token, body["title"])
			writeJSON(w, http.StatusCreated, report)
		case http.MethodGet:
			mu.Lock()
			var list []map[string]interface{}
			for _, r := range costReports {
				list = append(list, r)
			}
			mu.Unlock()
			if list == nil {
				list = []map[string]interface{}{}
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"cost_reports": list,
				"links":        map[string]interface{}{},
			})
		}
		return
	}

	// /cost_reports/{token}
	parts := strings.Split(strings.TrimPrefix(path, "/cost_reports/"), "/")
	token := parts[0]
	switch r.Method {
	case http.MethodGet:
		mu.Lock()
		report, ok := costReports[token]
		mu.Unlock()
		if !ok {
			http.Error(w, `{"errors":["not found"]}`, http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, report)
	case http.MethodPut:
		mu.Lock()
		report, ok := costReports[token]
		if ok {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			for k, v := range body {
				report[k] = v
			}
			report["updated_at"] = time.Now().Format(time.RFC3339)
		}
		mu.Unlock()
		if !ok {
			http.Error(w, `{"errors":["not found"]}`, http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, report)
	case http.MethodDelete:
		mu.Lock()
		_, ok := costReports[token]
		if ok {
			delete(costReports, token)
		}
		mu.Unlock()
		if !ok {
			http.Error(w, `{"errors":["not found"]}`, http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func main() {
	flag.Parse()
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Mock Vantage API starting on %s (reverse=%v, shuffle=%v)", addr, *reverseOrder, *shuffleOrder)
	http.HandleFunc("/", router)
	log.Fatal(http.ListenAndServe(addr, nil))
}
