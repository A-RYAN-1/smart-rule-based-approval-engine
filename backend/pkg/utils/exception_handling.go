package utils

import (
	"errors"

	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"

	"github.com/jackc/pgx/v5/pgconn"
)

func MapPgError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503":
			return apperrors.ErrForeignKeyViolation
		case "23505":
			return apperrors.ErrDuplicateEntry
		case "23514":
			return apperrors.ErrCheckConstraintFailed
		}
	}
	return apperrors.ErrDatabase
}
