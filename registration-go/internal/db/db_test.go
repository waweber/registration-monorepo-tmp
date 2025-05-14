package db_test

import (
	"context"
	"os"
	"registration/internal/db"
	"registration/internal/registration"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestDB(t *testing.T) {
	ctx := context.Background()
	url := os.Getenv("TEST_DB_URL")
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Error(err)
	}
	defer pool.Close()

	now := time.Now()

	reg := registration.Registration{
		RegistrationFields: registration.RegistrationFields{
			Id: "1",
			EventId: "1",
			Status: "created",
			Version: 1,
			DateCreated: now,
			Options: []string{"a", "b"},
		},
	}

	txn, err := db.NewDBSession(ctx, pool)
	if err != nil {
		t.Error(err)
	}
	defer txn.Rollback(ctx)

	err = txn.AddRegistration(ctx, reg)
	if err != nil {
		t.Error(err)
	}

	res, err := txn.GetRegistration(ctx, "1", "1", true)
	if err != nil {
		t.Error(err)
	}

	name := "Test"
	res.Nickname = &name

	res, err = txn.UpdateRegistration(ctx, res)
	if err != nil {
		t.Error(err)
	}

	if res.Version != 2 {
		t.Errorf("expected version 2, got %v", res.Version)
	}

	res2, err := txn.GetRegistration(ctx, "1", "1", false)
	if err != nil {
		t.Error(err)
	}

	if res2.Version != 2 {
		t.Errorf("expected version 2, got %v", res2.Version)
	}

	if *res2.Nickname != "Test" {
		t.Errorf("expected Test, got %v", res2.Nickname)
	}
}
