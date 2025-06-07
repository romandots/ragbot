package education

import (
	"context"

	"ragbot/internal/repository"
)

// Source defines a knowledge source that can load chunks into the database.
type Source interface {
	Start(ctx context.Context, repo *repository.Repository)
}
