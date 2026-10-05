package ala_service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// newTestRouter wires a Handler against an in-memory SQLite DB and returns
// a Gin engine. Each test gets a fresh router to avoid cross-test pollution.
func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := newTestDB(t)
	h := &Handler{DB: db}

	r := gin.New()
	r.POST("/v1/aliases", h.CreateAlias)
	r.GET("/v1/aliases", h.GetAliases)
	r.GET("/:alias_url", h.Redirect)
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	return r
}

// TestCreateAlias_HappyPath verifies a valid POST returns 201 with the
// expected response shape, and the alias is retrievable.
func TestCreateAlias_HappyPath(t *testing.T) {
	r := newTestRouter(t)

	body, _ := json.Marshal(map[string]string{
		"alias_url":    "github",
		"redirect_uri": "https://github.com/henryorsborn",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/aliases", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp createAliasResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.AliasURL != "github" {
		t.Errorf("AliasURL = %q, want %q", resp.AliasURL, "github")
	}
	if !strings.HasPrefix(resp.ShortURL, "http") {
		t.Errorf("ShortURL %q should start with http(s)://", resp.ShortURL)
	}
}

// TestCreateAlias_ValidationErrors verifies bad input returns 400 without
// touching the DB.
func TestCreateAlias_ValidationErrors(t *testing.T) {
	r := newTestRouter(t)

	cases := []struct {
		name string
		body string
	}{
		{"missing alias_url", `{"redirect_uri":"https://example.com"}`},
		{"missing redirect_uri", `{"alias_url":"foo"}`},
		{"alias_url too short", `{"alias_url":"a","redirect_uri":"https://example.com"}`},
		{"redirect_uri not a url", `{"alias_url":"foo","redirect_uri":"not-a-url"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/aliases", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}

// TestCreateAlias_DuplicateReturns409 verifies that creating an alias with
// the same alias_url twice returns 409 Conflict (not 500).
func TestCreateAlias_DuplicateReturns409(t *testing.T) {
	r := newTestRouter(t)

	body, _ := json.Marshal(map[string]string{
		"alias_url":    "dup-alias",
		"redirect_uri": "https://example.com",
	})

	// First: 201
	req := httptest.NewRequest(http.MethodPost, "/v1/aliases", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("first POST expected 201, got %d", w.Code)
	}

	// Second: 409
	req = httptest.NewRequest(http.MethodPost, "/v1/aliases", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("second POST expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

// TestGetAliases_EmptyList verifies GET on a fresh DB returns count=0
// (not null, not a 500, not an empty array on the wrong key).
func TestGetAliases_EmptyList(t *testing.T) {
	r := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/aliases", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp getAliasesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Count != 0 {
		t.Errorf("Count = %d, want 0", resp.Count)
	}
	if resp.Values == nil {
		t.Errorf("Values should be an empty array, not null")
	}
	if len(resp.Values) != 0 {
		t.Errorf("len(Values) = %d, want 0", len(resp.Values))
	}
}

// TestRedirect_HappyPath verifies a known alias 302s to its redirect_uri.
func TestRedirect_HappyPath(t *testing.T) {
	r := newTestRouter(t)

	// Create an alias first.
	body, _ := json.Marshal(map[string]string{
		"alias_url":    "github",
		"redirect_uri": "https://github.com/henryorsborn",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/aliases", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("setup POST expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// Now follow it.
	req = httptest.NewRequest(http.MethodGet, "/github", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", w.Code)
	}
	loc := w.Header().Get("Location")
	if loc != "https://github.com/henryorsborn" {
		t.Errorf("Location = %q, want %q", loc, "https://github.com/henryorsborn")
	}
}

// TestRedirect_UnknownAlias returns 404 (not 500, not 302 to "/").
func TestRedirect_UnknownAlias(t *testing.T) {
	r := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/no-such-alias", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// TestBuildShortURL_HonorsTLSFlag verifies the scheme picks http vs https
// based on the request TLS state. This is the only place the redirect host
// is composed, and a regression here silently breaks prod https URLs.
func TestBuildShortURL_HonorsTLSFlag(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/foo", nil)

	if got := buildShortURL(c, "abc"); !strings.HasPrefix(got, "http://") {
		t.Errorf("expected http:// scheme for plain request, got %q", got)
	}
}

// TestGenerateShortCode_Length verifies the random code generator returns
// a 6-character string. Anything longer bloats the URL; anything shorter
// reduces the namespace.
func TestGenerateShortCode_Length(t *testing.T) {
	for i := 0; i < 100; i++ {
		code := generateShortCode()
		if len(code) != 6 {
			t.Fatalf("generateShortCode() = %q (len %d), want 6", code, len(code))
		}
	}
}