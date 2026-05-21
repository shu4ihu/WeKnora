package notion

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Tencent/WeKnora/internal/datasource"
	"github.com/Tencent/WeKnora/internal/types"
)

func makeNotionConfig(cfg *Config, baseURL string, resourceIDs []string) *types.DataSourceConfig {
	return &types.DataSourceConfig{
		Type: types.ConnectorTypeNotion,
		Credentials: map[string]interface{}{
			"api_key": cfg.APIKey,
		},
		ResourceIDs: resourceIDs,
		Settings: map[string]interface{}{
			"base_url": baseURL,
		},
	}
}

func TestConnectorType(t *testing.T) {
	c := NewConnector()
	if c.Type() != types.ConnectorTypeNotion {
		t.Errorf("Type() = %q, want %q", c.Type(), types.ConnectorTypeNotion)
	}
}

func TestConnectorValidate(t *testing.T) {
	ts, cfg := fakeNotion()
	defer ts.Close()

	c := NewConnector()
	err := c.Validate(context.Background(), makeNotionConfig(cfg, ts.URL, nil))
	if err != nil {
		t.Fatalf("Validate() error: %v", err)
	}
}

func TestConnectorValidate_BadToken(t *testing.T) {
	ts, _ := fakeNotion()
	defer ts.Close()

	c := NewConnector()
	err := c.Validate(context.Background(), makeNotionConfig(
		&Config{APIKey: "bad-token"}, ts.URL, nil,
	))
	if err == nil {
		t.Fatal("expected error for bad token")
	}
}

func TestConnectorListResources(t *testing.T) {
	ts, cfg := fakeNotion()
	defer ts.Close()

	c := NewConnector()
	resources, err := c.ListResources(context.Background(), makeNotionConfig(cfg, ts.URL, nil))
	if err != nil {
		t.Fatalf("ListResources() error: %v", err)
	}

	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}
	if resources[0].ExternalID != "page-1" || resources[0].Type != "page" || resources[0].Name != "Test Page" {
		t.Errorf("resource[0] = %+v", resources[0])
	}
	if resources[1].ExternalID != "db-1" || resources[1].Type != "database" || resources[1].Name != "Test Database" {
		t.Errorf("resource[1] = %+v, want Name=%q", resources[1], "Test Database")
	}
}

func TestConnectorFetchAll(t *testing.T) {
	ts, cfg := fakeNotion()
	defer ts.Close()

	c := NewConnector()
	items, err := c.FetchAll(context.Background(), makeNotionConfig(cfg, ts.URL, []string{"page-1"}), []string{"page-1"})
	if err != nil {
		t.Fatalf("FetchAll() error: %v", err)
	}

	if len(items) == 0 {
		t.Fatal("expected at least 1 item")
	}

	// First item should be the page with markdown content
	found := false
	for _, item := range items {
		if item.ExternalID == "page-1" {
			found = true
			if item.ContentType != "text/markdown" {
				t.Errorf("ContentType = %q", item.ContentType)
			}
			if item.Metadata["channel"] != "notion" {
				t.Errorf("channel = %q", item.Metadata["channel"])
			}
			if len(item.Content) == 0 {
				t.Error("expected non-empty content")
			}
		}
	}
	if !found {
		t.Error("page-1 not found in items")
	}
}

func TestConnectorFetchAll_Database(t *testing.T) {
	ts, cfg := fakeNotion()
	defer ts.Close()

	c := NewConnector()
	items, err := c.FetchAll(context.Background(), makeNotionConfig(cfg, ts.URL, []string{"db-1"}), []string{"db-1"})
	if err != nil {
		t.Fatalf("FetchAll() error: %v", err)
	}

	if len(items) == 0 {
		t.Fatal("expected at least 1 item")
	}

	// The entire database should be synced as a single table knowledge item
	found := false
	for _, item := range items {
		if item.ExternalID == "db-1" {
			found = true
			if item.Metadata["object_type"] != "database" {
				t.Errorf("object_type = %q, want %q", item.Metadata["object_type"], "database")
			}
			if item.ContentType != "text/markdown" {
				t.Errorf("ContentType = %q", item.ContentType)
			}
			if len(item.Content) == 0 {
				t.Error("expected non-empty content")
			}
			// Verify it contains table headers
			contentStr := string(item.Content)
			if !strings.Contains(contentStr, "| Title |") {
				t.Errorf("expected table header in content, got: %s", contentStr)
			}
		}
	}
	if !found {
		t.Error("db-1 not found in items")
	}
}

func TestConnectorFetchAll_DoesNotTreatAuthFailureAsDatabase(t *testing.T) {
	mux := http.NewServeMux()
	writeUnauthorized := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  http.StatusUnauthorized,
			"code":    "unauthorized",
			"message": "API token is invalid.",
		})
	}
	mux.HandleFunc("/v1/search", writeUnauthorized)
	mux.HandleFunc("/v1/pages/page-1", writeUnauthorized)
	server := httptest.NewServer(mux)
	defer server.Close()

	c := NewConnector()
	_, err := c.FetchAll(context.Background(), makeNotionConfig(&Config{APIKey: "bad-token"}, server.URL, []string{"page-1"}), []string{"page-1"})
	if err == nil {
		t.Fatal("expected fetch error")
	}
	if !errors.Is(err, datasource.ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials error, got %v", err)
	}
}

func TestConnectorFetchAll_DoesNotFallbackFromDataSourceAuthFailureToDatabase(t *testing.T) {
	databaseFallbackCalls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/search", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"object":      "list",
			"results":     []interface{}{},
			"has_more":    false,
			"next_cursor": nil,
		})
	})
	mux.HandleFunc("/v1/pages/db-bad", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	mux.HandleFunc("/v1/data_sources/db-bad", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  http.StatusUnauthorized,
			"code":    "unauthorized",
			"message": "API token is invalid.",
		})
	})
	mux.HandleFunc("/v1/databases/db-bad", func(w http.ResponseWriter, r *http.Request) {
		databaseFallbackCalls++
		http.NotFound(w, r)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	c := NewConnector()
	_, err := c.FetchAll(context.Background(), makeNotionConfig(&Config{APIKey: "bad-token"}, server.URL, []string{"db-bad"}), []string{"db-bad"})
	if err == nil {
		t.Fatal("expected fetch error")
	}
	if !errors.Is(err, datasource.ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials error, got %v", err)
	}
	if databaseFallbackCalls != 0 {
		t.Fatalf("expected no database fallback for data_source auth failure, got %d calls", databaseFallbackCalls)
	}
}

// TestConnectorFetchAll_SingleRecord verifies that selecting a single database
// row by ID routes through fetchPage's record-detection branch and produces an
// item via buildRecordItem (instead of being silently dropped as an empty page).
func TestConnectorFetchAll_SingleRecord(t *testing.T) {
	ts, cfg := fakeNotion()
	defer ts.Close()

	c := NewConnector()
	items, err := c.FetchAll(context.Background(), makeNotionConfig(cfg, ts.URL, []string{"record-1"}), []string{"record-1"})
	if err != nil {
		t.Fatalf("FetchAll() error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ExternalID != "record-1" {
		t.Errorf("ExternalID = %q, want record-1", items[0].ExternalID)
	}
	if items[0].Metadata["object_type"] != "page" {
		t.Errorf("object_type = %q, want page", items[0].Metadata["object_type"])
	}
	if len(items[0].Content) == 0 {
		t.Error("expected non-empty content from buildRecordItem")
	}
}

func TestConnectorFetchIncremental_NoChanges(t *testing.T) {
	ts, cfg := fakeNotion()
	defer ts.Close()

	c := NewConnector()
	config := makeNotionConfig(cfg, ts.URL, []string{"page-1"})

	// First: full fetch to establish baseline
	_, err := c.FetchAll(context.Background(), config, []string{"page-1"})
	if err != nil {
		t.Fatalf("FetchAll() error: %v", err)
	}

	// Build a cursor that matches current state
	cursor := &types.SyncCursor{
		ConnectorCursor: map[string]interface{}{
			"page_edit_times": map[string]interface{}{
				"page-1": "2026-01-15T10:00:00Z",
			},
		},
	}

	items, newCursor, err := c.FetchIncremental(context.Background(), config, cursor)
	if err != nil {
		t.Fatalf("FetchIncremental() error: %v", err)
	}

	// No changes expected (timestamps match)
	if len(items) != 0 {
		t.Errorf("expected 0 items for no changes, got %d", len(items))
	}
	if newCursor == nil {
		t.Fatal("expected non-nil cursor")
	}
}

func TestConnectorFetchIncremental_ReturnsDiscoveryError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/search", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  http.StatusUnauthorized,
			"code":    "unauthorized",
			"message": "API token is invalid.",
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	c := NewConnector()
	config := makeNotionConfig(&Config{APIKey: "bad-token"}, server.URL, []string{"page-1"})
	cursor := &types.SyncCursor{
		ConnectorCursor: map[string]interface{}{
			"page_edit_times": map[string]interface{}{
				"page-1": "2026-01-15T10:00:00Z",
			},
		},
	}

	items, newCursor, err := c.FetchIncremental(context.Background(), config, cursor)
	if err == nil {
		t.Fatal("expected discovery error")
	}
	if !errors.Is(err, datasource.ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials error, got %v", err)
	}
	if len(items) != 0 || newCursor != nil {
		t.Fatalf("expected no items/cursor on discovery failure, got items=%d cursor=%#v", len(items), newCursor)
	}
}

func TestConnectorFetchIncremental_ReturnsDatabaseQueryError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/search", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"object": "list",
			"results": []interface{}{
				map[string]interface{}{
					"id":               "db-bad",
					"object":           "data_source",
					"url":              "https://notion.so/DB-Bad",
					"last_edited_time": "2026-01-16T10:00:00.000Z",
					"in_trash":         false,
					"parent":           map[string]interface{}{"type": "workspace", "workspace": true},
					"title":            []interface{}{map[string]interface{}{"plain_text": "Bad Database"}},
				},
			},
			"has_more":    false,
			"next_cursor": nil,
		})
	})
	mux.HandleFunc("/v1/data_sources/db-bad", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "db-bad",
			"object": "data_source",
			"title":  []interface{}{map[string]interface{}{"plain_text": "Bad Database"}},
		})
	})
	mux.HandleFunc("/v1/data_sources/db-bad/query", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  http.StatusUnauthorized,
			"code":    "unauthorized",
			"message": "API token is invalid.",
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	c := NewConnector()
	config := makeNotionConfig(&Config{APIKey: "bad-token"}, server.URL, []string{"db-bad"})
	cursor := &types.SyncCursor{
		ConnectorCursor: map[string]interface{}{
			"page_edit_times": map[string]interface{}{
				"db-bad":   "2026-01-15T10:00:00Z",
				"record-1": "2026-01-15T10:00:00Z",
			},
		},
	}

	items, newCursor, err := c.FetchIncremental(context.Background(), config, cursor)
	if err == nil {
		t.Fatal("expected database query error")
	}
	if !errors.Is(err, datasource.ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials error, got %v", err)
	}
	if len(items) != 0 || newCursor != nil {
		t.Fatalf("expected no items/cursor on database query failure, got items=%d cursor=%#v", len(items), newCursor)
	}
}

func TestPropertyToString(t *testing.T) {
	tests := []struct {
		name     string
		value    map[string]interface{}
		expected string
	}{
		{
			name: "select",
			value: map[string]interface{}{
				"type":   "select",
				"select": map[string]interface{}{"name": "Done"},
			},
			expected: "Done",
		},
		{
			name: "rich_text",
			value: map[string]interface{}{
				"type": "rich_text",
				"rich_text": []interface{}{
					map[string]interface{}{"plain_text": "Hello"},
				},
			},
			expected: "Hello",
		},
		{
			name: "number",
			value: map[string]interface{}{
				"type":   "number",
				"number": 42.0,
			},
			expected: "42",
		},
		{
			name: "checkbox true",
			value: map[string]interface{}{
				"type":     "checkbox",
				"checkbox": true,
			},
			expected: "true",
		},
		{
			name: "date",
			value: map[string]interface{}{
				"type": "date",
				"date": map[string]interface{}{"start": "2026-01-15", "end": "2026-01-20"},
			},
			expected: "2026-01-15 ~ 2026-01-20",
		},
		{
			name: "multi_select",
			value: map[string]interface{}{
				"type": "multi_select",
				"multi_select": []interface{}{
					map[string]interface{}{"name": "Tag1"},
					map[string]interface{}{"name": "Tag2"},
				},
			},
			expected: "Tag1, Tag2",
		},
		{
			name:     "nil value",
			value:    map[string]interface{}{"type": "rich_text", "rich_text": nil},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := propertyToString(tt.value)
			if got != tt.expected {
				t.Errorf("propertyToString() = %q, want %q", got, tt.expected)
			}
		})
	}
}
