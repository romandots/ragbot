package conversation_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"sync"
	"testing"

	"ragbot/internal/conversation"
)

// memory driver for sql.DB used in tests
var (
	memOnce  sync.Once
	memStore map[int64][]conversation.HistoryItem
	memMu    sync.Mutex
)

type memDriver struct{}

type memConn struct{}

type memRows struct {
	items []conversation.HistoryItem
	idx   int
}

func (d memDriver) Open(name string) (driver.Conn, error) { return memConn{}, nil }

func (memConn) Prepare(string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (memConn) Close() error                        { return nil }
func (memConn) Begin() (driver.Tx, error)           { return nil, driver.ErrSkip }

func (memConn) ExecContext(_ context.Context, _ string, args []driver.NamedValue) (driver.Result, error) {
	memMu.Lock()
	defer memMu.Unlock()
	if len(args) >= 3 {
		chatID, _ := args[0].Value.(int64)
		role, _ := args[1].Value.(string)
		content, _ := args[2].Value.(string)
		memStore[chatID] = append(memStore[chatID], conversation.HistoryItem{Role: role, Content: content})
	}
	return driver.RowsAffected(1), nil
}

func (memConn) QueryContext(_ context.Context, _ string, args []driver.NamedValue) (driver.Rows, error) {
	memMu.Lock()
	defer memMu.Unlock()
	var chatID int64
	if len(args) >= 1 {
		chatID, _ = args[0].Value.(int64)
	}
	items := append([]conversation.HistoryItem(nil), memStore[chatID]...)
	if len(items) > 20 {
		items = items[len(items)-20:]
	}
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	return &memRows{items: items}, nil
}

func (r *memRows) Columns() []string { return []string{"role", "content"} }
func (r *memRows) Close() error      { return nil }
func (r *memRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.items) {
		return io.EOF
	}
	item := r.items[r.idx]
	dest[0] = item.Role
	dest[1] = item.Content
	r.idx++
	return nil
}

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	memOnce.Do(func() { sql.Register("mem", memDriver{}) })
	memMu.Lock()
	memStore = make(map[int64][]conversation.HistoryItem)
	memMu.Unlock()
	db, err := sql.Open("mem", "")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return db
}

func TestAppendHistoryLimitsToTwenty(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	chatID := int64(1)
	for i := 0; i < 25; i++ {
		conversation.AppendHistory(db, chatID, "user", fmt.Sprintf("msg%d", i))
	}
	history := conversation.GetHistory(db, chatID)
	if got := len(history); got != 20 {
		t.Fatalf("expected history length 20, got %d", got)
	}
	if history[0].Content != "msg5" || history[len(history)-1].Content != "msg24" {
		t.Fatalf("unexpected messages returned: %+v", history)
	}
}

func TestGetHistoryReturnsCopy(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	chatID := int64(2)
	conversation.AppendHistory(db, chatID, "user", "first")
	conversation.AppendHistory(db, chatID, "assistant", "second")

	h1 := conversation.GetHistory(db, chatID)
	h1[0].Content = "changed"
	h1 = append(h1, conversation.HistoryItem{Role: "user", Content: "extra"})

	h2 := conversation.GetHistory(db, chatID)
	if len(h2) != 2 {
		t.Fatalf("expected history length 2, got %d", len(h2))
	}
	if h2[0].Content != "first" || h2[1].Content != "second" {
		t.Fatalf("unexpected history after modification: %+v", h2)
	}
}
