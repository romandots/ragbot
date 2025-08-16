package amo

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"ragbot/internal/config"
	"ragbot/internal/conversation"
	"ragbot/internal/repository"
)

var (
	regOnce     sync.Once
	lastChatID  int64
	lastContact sql.NullInt64
)

type memDriver struct{}

type memConn struct{}

func (memDriver) Open(name string) (driver.Conn, error) { return memConn{}, nil }

func (memConn) Prepare(string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (memConn) Close() error                        { return nil }
func (memConn) Begin() (driver.Tx, error)           { return nil, driver.ErrSkip }

func (memConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if strings.Contains(query, "amo_contact_id") {
		if len(args) > 0 {
			if args[0].Value == nil {
				lastContact = sql.NullInt64{}
			} else if v, ok := args[0].Value.(int64); ok {
				lastContact = sql.NullInt64{Int64: v, Valid: true}
			}
		}
		if len(args) > 1 {
			if v, ok := args[1].Value.(int64); ok {
				lastChatID = v
			}
		}
	}
	return driver.RowsAffected(1), nil
}

func (memConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return nil, driver.ErrSkip
}

func newTestRepo(t *testing.T) *repository.Repository {
	t.Helper()
	regOnce.Do(func() { sql.Register("amomem", memDriver{}) })
	db, err := sql.Open("amomem", "")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	lastChatID = 0
	lastContact = sql.NullInt64{}
	return repository.New(db)
}

type fakeHTTPClient struct{ requests []*http.Request }

func (f *fakeHTTPClient) Do(req *http.Request) (*http.Response, error) {
	f.requests = append(f.requests, req)
	if strings.Contains(req.URL.Path, "/contacts") {
		body := `{"_embedded":{"contacts":[{"id":123}]}}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
	}
	if strings.Contains(req.URL.Path, "/leads/complex") {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{}`))}, nil
	}
	return nil, nil
}

func TestSendLeadExistingContact(t *testing.T) {
	repo := newTestRepo(t)
	client := &AmoClient{HTTPClient: &fakeHTTPClient{}}
	config.Config = &config.AppConfig{AmoDomain: "example.com", AmoAccessToken: "token"}
	info := conversation.ChatInfo{
		ChatID:       1,
		Name:         sql.NullString{String: "A", Valid: true},
		Phone:        sql.NullString{String: "1", Valid: true},
		AmoContactID: sql.NullInt64{Int64: 321, Valid: true},
	}
	fhc := client.HTTPClient.(*fakeHTTPClient)
	if err := client.SendLeadToAMO(repo, &info, "link"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fhc.requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(fhc.requests))
	}
	if !strings.Contains(fhc.requests[0].URL.Path, "/leads/complex") {
		t.Fatalf("unexpected path: %s", fhc.requests[0].URL.Path)
	}
	if lastContact.Valid {
		t.Fatalf("contact id should not be updated")
	}
}

func TestSendLeadCreatesContact(t *testing.T) {
	repo := newTestRepo(t)
	client := &AmoClient{HTTPClient: &fakeHTTPClient{}}
	config.Config = &config.AppConfig{AmoDomain: "example.com", AmoAccessToken: "token"}
	info := conversation.ChatInfo{
		ChatID: 2,
		Name:   sql.NullString{String: "B", Valid: true},
		Phone:  sql.NullString{String: "2", Valid: true},
	}
	fhc := client.HTTPClient.(*fakeHTTPClient)
	if err := client.SendLeadToAMO(repo, &info, "link"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fhc.requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(fhc.requests))
	}
	if !lastContact.Valid || lastContact.Int64 != 123 || lastChatID != 2 {
		t.Fatalf("contact id not saved correctly: %+v %d", lastContact, lastChatID)
	}
}

func TestSendLeadWithTags(t *testing.T) {
	repo := newTestRepo(t)
	client := &AmoClient{HTTPClient: &fakeHTTPClient{}}
	config.Config = &config.AppConfig{AmoDomain: "example.com", AmoAccessToken: "token"}
	
	// Load config with tags
	loadConfig()
	
	info := conversation.ChatInfo{
		ChatID:  3,
		Name:    sql.NullString{String: "Test User", Valid: true},
		Phone:   sql.NullString{String: "1234567890", Valid: true},
		Title:   sql.NullString{String: "Test Lead", Valid: true},
		Summary: sql.NullString{String: "Клиент интересуется консультацией по услугам", Valid: true},
		Interest: sql.NullString{String: "Услуги", Valid: true},
	}
	
	fhc := client.HTTPClient.(*fakeHTTPClient)
	if err := client.SendLeadToAMO(repo, &info, "https://example.com/chat"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Should create contact and lead
	if len(fhc.requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(fhc.requests))
	}
	
	// Check that lead request contains tags
	leadRequest := fhc.requests[1] // Second request should be for lead
	if !strings.Contains(leadRequest.URL.Path, "/leads/complex") {
		t.Fatalf("expected lead request, got: %s", leadRequest.URL.Path)
	}
	
	// Read request body to check for tags
	body, err := io.ReadAll(leadRequest.Body)
	if err != nil {
		t.Fatalf("failed to read request body: %v", err)
	}
	
	// Check that body contains tags field
	if !strings.Contains(string(body), `"tags"`) {
		t.Fatalf("request body should contain tags field: %s", string(body))
	}
	
	// Check for specific tags
	if !strings.Contains(string(body), `"RAG Бот"`) {
		t.Fatalf("request body should contain RAG Бот tag: %s", string(body))
	}
	
	if !strings.Contains(string(body), `"Новый клиент"`) {
		t.Fatalf("request body should contain Новый клиент tag: %s", string(body))
	}
	
	if !strings.Contains(string(body), `"Консультация"`) {
		t.Fatalf("request body should contain Консультация tag: %s", string(body))
	}
	
	if !strings.Contains(string(body), `"Услуга"`) {
		t.Fatalf("request body should contain Услуга tag: %s", string(body))
	}
	
	if !strings.Contains(string(body), `"Интерес: Услуги"`) {
		t.Fatalf("request body should contain interest tag: %s", string(body))
	}
}
