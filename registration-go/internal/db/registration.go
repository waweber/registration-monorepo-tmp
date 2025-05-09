package db

import (
	"context"
	"errors"
	"registration/internal/registration"
	"time"

	"github.com/jackc/pgx/v5"
)


func (db *DBSession) GetRegistration(ctx context.Context, eventId string, id string, lock bool) (*registration.Registration, error) {
	query := "SELECT id, event_id, status, version, date_created, date_updated, options, first_name, last_name, preferred_name, nickname, number, email, account_id, checked_in, date_checked_in, extra_data FROM registration WHERE event_id = $1 AND id = $2"

	if lock {
		query = query + " FOR UPDATE"
	}

	row := db.QueryRow(ctx, query, eventId, id)

	reg := &registration.Registration{}

	err := row.Scan(
		&reg.RegistrationFields.Id,
		&reg.RegistrationFields.EventID,
		&reg.RegistrationFields.Status,
		&reg.RegistrationFields.Version,
		&reg.RegistrationFields.DateCreated,
		&reg.RegistrationFields.DateUpdated,
		&reg.RegistrationFields.Options,
		&reg.RegistrationFields.FirstName,
		&reg.RegistrationFields.LastName,
		&reg.RegistrationFields.PreferredName,
		&reg.RegistrationFields.Nickname,
		&reg.RegistrationFields.Number,
		&reg.RegistrationFields.Email,
		&reg.RegistrationFields.AccountID,
		&reg.RegistrationFields.CheckedIn,
		&reg.RegistrationFields.DateCheckedIn,
		&reg.ExtraData,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return reg, nil
}

func (db *DBSession) AddRegistration(ctx context.Context, reg *registration.Registration) error {
	_, err := db.Exec(
		ctx,
		"INSERT INTO registration (id, event_id, status, version, date_created, date_updated, options, first_name, last_name, preferred_name, nickname, number, email, account_id, checked_in, date_checked_in, extra_data) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)",
		reg.RegistrationFields.Id,
		reg.RegistrationFields.EventID,
		reg.RegistrationFields.Status,
		reg.RegistrationFields.Version,
		reg.RegistrationFields.DateCreated,
		reg.RegistrationFields.DateUpdated,
		reg.RegistrationFields.Options,
		reg.RegistrationFields.FirstName,
		reg.RegistrationFields.LastName,
		reg.RegistrationFields.PreferredName,
		reg.RegistrationFields.Nickname,
		reg.RegistrationFields.Number,
		reg.RegistrationFields.Email,
		reg.RegistrationFields.AccountID,
		reg.RegistrationFields.CheckedIn,
		reg.RegistrationFields.DateCheckedIn,
		reg.ExtraData,
	)

	return err
}

func (db *DBSession) UpdateRegistration(ctx context.Context, reg *registration.Registration) (*registration.Registration, error) {
	var final *registration.Registration = &registration.Registration{}
	*final = *reg
	final.Version += 1

	now := time.Now()

	final.DateUpdated = &now

	_, err := db.Exec(
		ctx,
		"UPDATE registration SET version = $1, date_updated = $2, options = $3, first_name = $4, last_name = $5, preferred_name = $6, nickname = $7, number = $8, email = $9, account_id = $10, checked_in = $11, date_checked_in = $12, extra_data = $13 WHERE id = $14",
		final.RegistrationFields.Version,
		final.RegistrationFields.DateUpdated,
		final.RegistrationFields.Options,
		final.RegistrationFields.FirstName,
		final.RegistrationFields.LastName,
		final.RegistrationFields.PreferredName,
		final.RegistrationFields.Nickname,
		final.RegistrationFields.Number,
		final.RegistrationFields.Email,
		final.RegistrationFields.AccountID,
		final.RegistrationFields.CheckedIn,
		final.RegistrationFields.DateCheckedIn,
		final.ExtraData,
		final.Id,
	)

	return final, err
}

func (db *DBSession) DeleteRegistration(ctx context.Context, id string) error {
	_, err := db.Exec(ctx, "DELETE FROM registration WHERE id = $1", id)
	return err
}
