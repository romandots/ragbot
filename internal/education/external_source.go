package education

import (
	"context"
	"database/sql"
	"log"
)

// ExternalDBSource is a stub for future external DB integration.
type ExternalDBSource struct{}

func (e *ExternalDBSource) Start(ctx context.Context, db *sql.DB) {
	log.Println("ExternalDBSource not implemented yet")
}
