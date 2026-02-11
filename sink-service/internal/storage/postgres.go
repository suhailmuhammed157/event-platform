package storage

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

type Event struct {
	ID      string
	Payload []byte
	Msg     kafka.Message
}

type PostgresSink struct {
	db *sql.DB
}

func NewPostgresSink(dsn string) (*PostgresSink, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	return &PostgresSink{db: db}, nil
}

func (p *PostgresSink) InsertBatch(ctx context.Context, events []Event) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO events (id, payload)
		VALUES ($1, $2)
		ON CONFLICT (id) DO NOTHING
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range events {
		if _, err := stmt.ExecContext(ctx, e.ID, e.Payload); err != nil {
			return err
		}
	}

	return tx.Commit()
}
