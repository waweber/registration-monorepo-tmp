package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type EventStats struct {
	EventID    string
	NextNumber int
}

func (db *DBSession) GetEventStats(ctx context.Context, eventId string, lock bool) (EventStats, error) {
	query := "SELECT event_id, next_number FROM event_stats WHERE event_id = $1"
	if lock {
		query = query + " FOR UPDATE"
	}

	res := db.QueryRow(ctx, query, eventId)

	var stats EventStats
	err := res.Scan(
		&stats.EventID,
		&stats.NextNumber,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		// Create new row
		query := "INSERT INTO event_stats (event_id, next_number) VALUES ($1, $2)"
		_, err := db.Exec(ctx, query, eventId, 1)
		if err != nil {
			return EventStats{}, err
		}
		return EventStats{EventID: eventId, NextNumber: 1}, nil
	}

	return stats, err
}

func (db *DBSession) UpdateEventStats(ctx context.Context, stats EventStats) error {
	query := "UPDATE event_stats SET next_number = $1 WHERE event_id = $2"
	_, err := db.Exec(ctx, query, stats.NextNumber, stats.EventID)
	return err
}
