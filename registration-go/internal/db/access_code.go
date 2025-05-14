package db

import (
	"context"
	"registration/internal/access_code"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

const accessCodeAlphabet = "BCDFGHJKLMNPQRSTVWXYZ23456789"
const accessCodeLen = 8

func (db *DBSession) GetAccessCode(ctx context.Context, eventId string, code string, lock bool) (access_code.AccessCode, error) {
	query := "SELECT code, event_id, date_created, date_expires, name, used, options FROM access_code WHERE code = $1 AND event_id = $2"
	if lock {
		query = query + " FOR UPDATE"
	}

	var accessCode access_code.AccessCode
	res := db.QueryRow(ctx, query, code, eventId)
	err := res.Scan(
		&accessCode.Code,
		&accessCode.EventID,
		&accessCode.DateCreated,
		&accessCode.DateExpires,
		&accessCode.Name,
		&accessCode.Used,
		&accessCode.Options,
	)
	return accessCode, err
}

func (db *DBSession) SetAccessCodeUsed(ctx context.Context, code string) error {
	query := "UPDATE access_code SET used = ? WHERE code = ?"
	_, err := db.Exec(ctx, query, true, code)
	return err
}

func (db *DBSession) CreateAccessCode(ctx context.Context, accessCode access_code.AccessCode) (access_code.AccessCode, error) {
	copy := accessCode
	if copy.Code != "" {
		newCode, err := gonanoid.Generate(accessCodeAlphabet, accessCodeLen)
		if err != nil {
			return access_code.AccessCode{}, err
		}
		copy.Code = newCode
	}
	copy.DateCreated = time.Now()
	copy.Used = false

	query := "INSERT INTO access_code (code, event_id, date_created, date_expires, name, used, options) VALUES ($1, $2, $3, $4, $5, $6, $7)"
	_, err := db.Exec(ctx, query, copy.Code, copy.EventID, copy.DateCreated, copy.DateExpires, copy.Name, copy.Used, copy.Options)
	return copy, err
}
