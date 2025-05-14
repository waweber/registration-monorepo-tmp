package db

import (
	"context"
	"registration/internal/registration"
	"time"
)

func (db *DBSession) GetRegistration(ctx context.Context, eventId string, id string, lock bool) (registration.Registration, error) {
	query := "SELECT id, event_id, status, version, date_created, date_updated, options, first_name, last_name, preferred_name, nickname, number, email, account_id, checked_in, date_checked_in, extra_data FROM registration WHERE event_id = $1 AND id = $2"

	if lock {
		query = query + " FOR UPDATE"
	}

	row := db.QueryRow(ctx, query, eventId, id)

	reg := registration.Registration{}

	err := row.Scan(
		&reg.RegistrationFields.Id,
		&reg.RegistrationFields.EventId,
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
		&reg.RegistrationFields.AccountId,
		&reg.RegistrationFields.CheckedIn,
		&reg.RegistrationFields.DateCheckedIn,
		&reg.ExtraData,
	)

	return reg, err
}

func (db *DBSession) AddRegistration(ctx context.Context, reg registration.Registration) error {
	_, err := db.Exec(
		ctx,
		"INSERT INTO registration (id, event_id, status, version, date_created, date_updated, options, first_name, last_name, preferred_name, nickname, number, email, account_id, checked_in, date_checked_in, extra_data) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)",
		reg.RegistrationFields.Id,
		reg.RegistrationFields.EventId,
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
		reg.RegistrationFields.AccountId,
		reg.RegistrationFields.CheckedIn,
		reg.RegistrationFields.DateCheckedIn,
		reg.ExtraData,
	)

	return err
}

func (db *DBSession) UpdateRegistration(ctx context.Context, reg registration.Registration) (registration.Registration, error) {
	final := reg
	final.Version += 1

	now := time.Now()

	final.DateUpdated = &now

	_, err := db.Exec(
		ctx,
		"UPDATE registration SET status = $1, version = $2, date_created = $3, date_updated = $4, options = $5, first_name = $6, last_name = $7, preferred_name = $8, nickname = $9, number = $10, email = $11, account_id = $12, checked_in = $13, date_checked_in = $14, extra_data = $15 WHERE id = $16",
		final.RegistrationFields.Status,
		final.RegistrationFields.Version,
		final.RegistrationFields.DateCreated,
		final.RegistrationFields.DateUpdated,
		final.RegistrationFields.Options,
		final.RegistrationFields.FirstName,
		final.RegistrationFields.LastName,
		final.RegistrationFields.PreferredName,
		final.RegistrationFields.Nickname,
		final.RegistrationFields.Number,
		final.RegistrationFields.Email,
		final.RegistrationFields.AccountId,
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
