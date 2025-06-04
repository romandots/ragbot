package education

import (
	"context"
	"database/sql"
)

// Source defines a knowledge source that can load chunks into the database.
type Source interface {
	Start(ctx context.Context, db *sql.DB)
}
