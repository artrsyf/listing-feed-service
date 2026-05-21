package executors

import (
	"context"
	"database/sql"
	"time"
)

type PostgresExecutor struct {
	db *sql.DB
}

func NewPostgresExecutor(db *sql.DB) *PostgresExecutor {
	return &PostgresExecutor{db: db}
}

func (p *PostgresExecutor) RunQuery(ctx context.Context, query string) (time.Duration, error) {

	start := time.Now()

	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	// fully consume result (important for real benchmark)
	for rows.Next() {
	}

	return time.Since(start), nil
}
