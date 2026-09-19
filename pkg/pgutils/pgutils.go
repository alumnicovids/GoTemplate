package pgutil

import (
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func ToUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func ToUUIDPtr(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return ToUUID(*id)
}

func FromUUID(id pgtype.UUID) uuid.UUID {
	return uuid.UUID(id.Bytes)
}

func FromUUIDPtr(id pgtype.UUID) *string {
	if !id.Valid {
		return nil
	}
	s := FromUUID(id).String()
	return &s
}

func ToBoolPtr(b *bool) pgtype.Bool {
	if b == nil {
		return pgtype.Bool{}
	}
	return pgtype.Bool{Bool: *b, Valid: true}
}

func ToTextPtr(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func FromTextPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func ToDatePtr(s *string) (pgtype.Date, error) {
	if s == nil {
		return pgtype.Date{}, nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return pgtype.Date{}, err
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}

func FromDatePtr(d pgtype.Date) *string {
	if !d.Valid {
		return nil
	}
	s := d.Time.Format("2006-01-02")
	return &s
}

func ToTimeOfDay(s string) (pgtype.Time, error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return pgtype.Time{}, err
	}
	micros := int64(t.Hour())*3600e6 + int64(t.Minute())*60e6
	return pgtype.Time{Microseconds: micros, Valid: true}, nil
}

func FromTimeOfDay(t pgtype.Time) string {
	total := t.Microseconds / 1e6
	return fmt.Sprintf("%02d:%02d", total/3600, (total%3600)/60)
}

func ReadErr(err error, notFound error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return notFound
	}
	return err
}

func WriteErr(err error, uniqueConstraints map[string]error, invalidRef, notFound error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			if mapped, ok := uniqueConstraints[pgErr.ConstraintName]; ok {
				return mapped
			}
		case "23503":
			return invalidRef
		}
	}
	return ReadErr(err, notFound)
}
