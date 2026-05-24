package executors

import (
	"context"
	"database/sql"
	"time"
)

type PostgresExecutor struct {
	db       *sql.DB
	joinMode string
}

func NewPostgresExecutor(db *sql.DB) *PostgresExecutor {
	return &PostgresExecutor{db: db}
}

func (p *PostgresExecutor) SetJoinMode(mode string) {
	p.joinMode = mode
}

func (p *PostgresExecutor) RunQuery(ctx context.Context, query string) (time.Duration, error) {

	start := time.Now()

	if p.joinMode != "" {
		tx, err := p.db.BeginTx(ctx, nil)
		if err != nil {
			return 0, err
		}
		if _, err := tx.ExecContext(ctx, joinModeSQL(p.joinMode)); err != nil {
			tx.Rollback()
			return 0, err
		}

		rows, err := tx.QueryContext(ctx, query)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		for rows.Next() {
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			tx.Rollback()
			return 0, err
		}
		rows.Close()
		if err := tx.Commit(); err != nil {
			return 0, err
		}

		return time.Since(start), nil
	}

	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	// fully consume result (important for real benchmark)
	for rows.Next() {
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	return time.Since(start), nil
}

func joinModeSQL(mode string) string {
	switch mode {
	case "hash":
		return "SET LOCAL enable_hashjoin = on; SET LOCAL enable_mergejoin = off; SET LOCAL enable_nestloop = off"
	case "merge":
		return "SET LOCAL enable_mergejoin = on; SET LOCAL enable_hashjoin = off; SET LOCAL enable_nestloop = off"
	case "nested_loop":
		return "SET LOCAL enable_nestloop = on; SET LOCAL enable_hashjoin = off; SET LOCAL enable_mergejoin = off"
	default:
		return "SET LOCAL enable_hashjoin = on; SET LOCAL enable_mergejoin = on; SET LOCAL enable_nestloop = on"
	}
}
