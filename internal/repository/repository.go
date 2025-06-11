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

const historyCallRequested = "** хочет, чтобы ему перезвонили **"

func New(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

// --- chunk operations ---

// AddChunk inserts a new chunk. It returns its ID the row was inserted.
func (r *Repository) AddChunk(ctx context.Context, content, source string) (int, error) {
	var id int
	err := r.db.QueryRowContext(ctx,
		"INSERT INTO chunks(content, source) VALUES($1,$2) ON CONFLICT (content) DO NOTHING RETURNING id",
		content, source,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repository) DeleteChunk(ctx context.Context, id int) (string, error) {
	var content string
	err := r.db.QueryRowContext(ctx, "DELETE FROM chunks WHERE id=$1 RETURNING content", id).Scan(&content)
	if err != nil {
		return "", err
	}
	return content, nil
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

func (r *Repository) GetChunkByExtID(ctx context.Context, source, extID string) (id int, createdAt time.Time, content string, found bool, err error) {
	err = r.db.QueryRowContext(ctx,
		"SELECT id, created_at, content FROM chunks WHERE source=$1 AND ext_id=$2",
		source, extID,
	).Scan(&id, &createdAt, &content)
	if err == sql.ErrNoRows {
		return 0, time.Time{}, "", false, nil
	}
	if err != nil {
		return 0, time.Time{}, "", false, err
	}
	return id, createdAt, content, true, nil
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

func (r *Repository) UpdateChunkCreatedAt(ctx context.Context, id int, createdAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE chunks SET created_at=$1 WHERE id=$2",
		createdAt, id,
	)
	return err
}

// ListChunksWithoutExtID returns all chunks that don't have an external ID.
func (r *Repository) ListChunksWithoutExtID(ctx context.Context) ([]models.Chunk, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, content FROM chunks WHERE ext_id IS NULL ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chunks []models.Chunk
	for rows.Next() {
		var c models.Chunk
		if err := rows.Scan(&c.ID, &c.Content); err != nil {
			return chunks, err
		}
		chunks = append(chunks, c)
	}
	return chunks, nil
}

// --- conversation operations ---

type ChatInfo struct {
	ID           string
	ChatID       int64
	Username     sql.NullString
	Title        sql.NullString
	Summary      sql.NullString
	Interest     sql.NullString
	Name         sql.NullString
	Phone        sql.NullString
	AmoContactID sql.NullInt64
}

type HistoryItem struct {
	Role    string
	Content string
}

// AddVisit stores a visit to the landing page.
func (r *Repository) AddVisit(ctx context.Context, ip, ua, referer, utmSource, utmMedium, utmCampaign, utmContent, utmTerm string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO visits(ip, user_agent, referer, utm_source, utm_medium, utm_campaign, utm_content, utm_term) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`,
		ip, ua, referer, utmSource, utmMedium, utmCampaign, utmContent, utmTerm,
	)
	return err
}

func (r *Repository) EnsureSession(ctx context.Context, chatID int64, username string) (string, error) {
	var uuid string
	err := r.db.QueryRowContext(ctx, `SELECT uuid FROM conversations WHERE chat_id=$1`, chatID).Scan(&uuid)
	if err == sql.ErrNoRows {
		err = r.db.QueryRowContext(ctx, `INSERT INTO conversations(chat_id, username) VALUES($1,$2) RETURNING uuid`, chatID, username).Scan(&uuid)
	} else if err == nil {
		if username != "" {
			r.db.ExecContext(ctx, `UPDATE conversations SET username=$1 WHERE chat_id=$2`, username, chatID)
		}
	}
	return uuid, err
}

func (r *Repository) GetChatInfoByChatID(ctx context.Context, chatID int64) (ChatInfo, error) {
	var info ChatInfo
	err := r.db.QueryRowContext(ctx,
		`SELECT uuid, username, summary, title, interest, name, phone, amo_contact_id FROM conversations WHERE chat_id=$1`, chatID).
		Scan(&info.ID, &info.Username, &info.Summary, &info.Title, &info.Interest, &info.Name, &info.Phone, &info.AmoContactID)
	if err != nil {
		return info, err
	}
	info.ChatID = chatID
	return info, nil
}

func (r *Repository) GetChatInfoByUUID(ctx context.Context, uuid string) (ChatInfo, error) {
	var info ChatInfo
	err := r.db.QueryRowContext(ctx,
		`SELECT chat_id, username, summary, title, interest, name, phone, amo_contact_id FROM conversations WHERE uuid=$1`, uuid).
		Scan(&info.ChatID, &info.Username, &info.Summary, &info.Title, &info.Interest, &info.Name, &info.Phone, &info.AmoContactID)
	if err != nil {
		return info, err
	}
	info.ID = uuid
	return info, nil
}

func (r *Repository) UpdateSummary(ctx context.Context, chatID int64, summary, title, interest string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE conversations SET summary=$1, title=$2, interest=$3, updated_at=NOW() WHERE chat_id=$4`, summary, title, interest, chatID)
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

func (r *Repository) UpdateAmoContactID(ctx context.Context, chatID int64, contactID sql.NullInt64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE conversations SET amo_contact_id=$1, updated_at=NOW() WHERE chat_id=$2`, contactID, chatID)
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

// CountUniqueChats returns number of unique chat IDs in conversation history.
func (r *Repository) CountUniqueChats(ctx context.Context) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT chat_id) FROM conversation_history`).Scan(&n)
	return n, err
}

// CountDeals returns number of unique chats where a call to manager was requested.
func (r *Repository) CountDeals(ctx context.Context) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT chat_id) FROM conversation_history WHERE content=$1`, historyCallRequested).Scan(&n)
	return n, err
}

// CountCommandUsage returns number of times a command was issued.
func (r *Repository) CountCommandUsage(ctx context.Context, cmd string) (int, error) {
	var n int
	like := cmd + `%`
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM conversation_history WHERE content LIKE $1`, like).Scan(&n)
	return n, err
}

// MessageCountsBeforeDeal returns for each chat id that generated a lead the number of user messages before the request.
func (r *Repository) MessageCountsBeforeDeal(ctx context.Context) ([]struct {
	ChatID   int64
	Username sql.NullString
	Count    int
}, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT c.chat_id, c.username FROM conversations c JOIN conversation_history h ON c.chat_id=h.chat_id WHERE h.content=$1`, historyCallRequested)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []struct {
		ChatID   int64
		Username sql.NullString
		Count    int
	}
	for rows.Next() {
		var chatID int64
		var username sql.NullString
		if err := rows.Scan(&chatID, &username); err != nil {
			return result, err
		}
		var count int
		err := r.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM conversation_history WHERE chat_id=$1 AND role='user' AND id <= (SELECT id FROM conversation_history WHERE chat_id=$1 AND content=$2 ORDER BY id ASC LIMIT 1)`,
			chatID, historyCallRequested).Scan(&count)
		if err != nil {
			return result, err
		}
		result = append(result, struct {
			ChatID   int64
			Username sql.NullString
			Count    int
		}{chatID, username, count})
	}
	return result, nil
}
