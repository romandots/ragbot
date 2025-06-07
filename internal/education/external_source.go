package education

import (
	"context"
	"log"

	"ragbot/internal/repository"
)

// ExternalDBSource is a stub for future external DB integration.
type ExternalDBSource struct{}

func (e *ExternalDBSource) Start(ctx context.Context, repo *repository.Repository) {
	log.Println("ExternalDBSource not implemented yet")
}
