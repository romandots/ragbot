package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/pgvector/pgvector-go"
	"ragbot/internal/models"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

// --- chunk operations ---

// AddChunk inserts a new chunk. It returns true if the row was inserted.
func (r *Repository) AddChunk(ctx context.Context, content, source string) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO chunks(content, source) VALUES($1,$2) ON CONFLICT (content) DO NOTHING",
		content, source,
	)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (r *Repository) DeleteChunk(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM chunks WHERE id=$1", id)
	return err
}

func (r *Repository) UpdateChunk(ctx context.Context, id int, content string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE chunks SET content=$1, embedding=NULL, processed_at=NULL WHERE id=$2",
		content, id,
	)
	return err
}

// GetUnprocessedChunks returns chunks without embedding limited by n.
func (r *Repository) GetUnprocessedChunks(ctx context.Context, limit int) ([]models.Chunk, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, content FROM chunks WHERE processed_at IS NULL LIMIT $1", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []models.Chunk
	for rows.Next() {
		var c models.Chunk
		if err := rows.Scan(&c.ID, &c.Content); err != nil {
			return res, err
		}
		res = append(res, c)
	}
	return res, nil
}

func (r *Repository) UpdateChunkEmbedding(ctx context.Context, id int, vec []float32) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE chunks SET embedding=$1, processed_at=NOW() WHERE id=$2",
		pgvector.NewVector(vec), id,
	)
	return err
}

func (r *Repository) SearchChunks(ctx context.Context, vec []float32, limit int) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT content FROM chunks WHERE processed_at IS NOT NULL ORDER BY embedding <-> $1 LIMIT $2",
		pgvector.NewVector(vec), limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return out, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (r *Repository) GetChunkByExtID(ctx context.Context, source, extID string) (id int, createdAt time.Time, found bool, err error) {
	err = r.db.QueryRowContext(ctx,
		"SELECT id, created_at FROM chunks WHERE source=$1 AND ext_id=$2",
		source, extID,
	).Scan(&id, &createdAt)
	if err == sql.ErrNoRows {
		return 0, time.Time{}, false, nil
	}
	if err != nil {
		return 0, time.Time{}, false, err
	}
	return id, createdAt, true, nil
}

func (r *Repository) InsertChunkWithExtID(ctx context.Context, content, source, extID string, createdAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO chunks(content, source, ext_id, created_at) VALUES($1,$2,$3,$4)",
		content, source, extID, createdAt,
	)
	return err
}

func (r *Repository) UpdateChunkWithCreatedAt(ctx context.Context, id int, content string, createdAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE chunks SET content=$1, created_at=$2, embedding=NULL, processed_at=NULL WHERE id=$3",
		content, createdAt, id,
	)
	return err
}

// --- conversation operations ---

type ChatInfo struct {
	ID      string
	ChatID  int64
	Summary sql.NullString
	Name    sql.NullString
	Phone   sql.NullString
}

type HistoryItem struct {
	Role    string
	Content string
}

func (r *Repository) EnsureSession(ctx context.Context, chatID int64) (string, error) {
	var uuid string
	err := r.db.QueryRowContext(ctx, `SELECT uuid FROM conversations WHERE chat_id=$1`, chatID).Scan(&uuid)
	if err == sql.ErrNoRows {
		err = r.db.QueryRowContext(ctx, `INSERT INTO conversations(chat_id) VALUES($1) RETURNING uuid`, chatID).Scan(&uuid)
	}
	return uuid, err
}

func (r *Repository) GetChatInfoByChatID(ctx context.Context, chatID int64) (ChatInfo, error) {
	var info ChatInfo
	err := r.db.QueryRowContext(ctx,
		`SELECT uuid, summary, name, phone FROM conversations WHERE chat_id=$1`, chatID).Scan(&info.ID, &info.Summary, &info.Name, &info.Phone)
	if err != nil {
		return info, err
	}
	info.ChatID = chatID
	return info, nil
}

func (r *Repository) GetChatInfoByUUID(ctx context.Context, uuid string) (ChatInfo, error) {
	var info ChatInfo
	err := r.db.QueryRowContext(ctx,
		`SELECT chat_id, summary, name, phone FROM conversations WHERE uuid=$1`, uuid).Scan(&info.ChatID, &info.Summary, &info.Name, &info.Phone)
	if err != nil {
		return info, err
	}
	info.ID = uuid
	return info, nil
}

func (r *Repository) UpdateSummary(ctx context.Context, chatID int64, summary string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE conversations SET summary=$1, updated_at=NOW() WHERE chat_id=$2`, summary, chatID)
	return err
}

func (r *Repository) UpdateName(ctx context.Context, chatID int64, name string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE conversations SET name=$1, updated_at=NOW() WHERE chat_id=$2`, name, chatID)
	return err
}

func (r *Repository) UpdatePhone(ctx context.Context, chatID int64, phone string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE conversations SET phone=$1, updated_at=NOW() WHERE chat_id=$2`, phone, chatID)
	return err
}

func (r *Repository) AppendHistory(ctx context.Context, chatID int64, role, text string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO conversation_history(chat_id, role, content) VALUES ($1, $2, $3)`,
		chatID, role, text,
	)
	return err
}

func (r *Repository) GetHistory(ctx context.Context, chatID int64, limit int) ([]HistoryItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT role, content FROM conversation_history WHERE chat_id=$1 ORDER BY id DESC LIMIT $2`, chatID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []HistoryItem
	for rows.Next() {
		var it HistoryItem
		if err := rows.Scan(&it.Role, &it.Content); err != nil {
			return items, err
		}
		items = append(items, it)
	}
	// reverse
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	return items, nil
}

func (r *Repository) GetFullHistory(ctx context.Context, chatID int64) ([]HistoryItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT role, content FROM conversation_history WHERE chat_id=$1 ORDER BY id ASC`, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []HistoryItem
	for rows.Next() {
		var it HistoryItem
		if err := rows.Scan(&it.Role, &it.Content); err != nil {
			return items, err
		}
		items = append(items, it)
	}
	return items, nil
}
