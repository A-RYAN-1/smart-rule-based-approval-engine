package repositories

import (
	"context"

	"rule-based-approval-engine/internal/models"
	"rule-based-approval-engine/internal/pkg/apperrors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	db *pgxpool.Pool
}

// instance
func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	err := r.db.QueryRow(
		ctx,
		`SELECT id, name, email, password_hash, grade_id, role, manager_id, created_at
		 FROM users WHERE email=$1`,
		email,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.GradeID,
		&user.Role,
		&user.ManagerID,
		&user.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, apperrors.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User

	err := r.db.QueryRow(
		ctx,
		`SELECT id, name, email, password_hash, grade_id, role, manager_id, created_at
		 FROM users WHERE id=$1`,
		id,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.GradeID,
		&user.Role,
		&user.ManagerID,
		&user.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, apperrors.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Create(ctx context.Context, tx pgx.Tx, user *models.User) (int64, error) {
	var userID int64

	err := tx.QueryRow(
		ctx,
		`INSERT INTO users (name, email, password_hash, grade_id, role, manager_id)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id`,
		user.Name,
		user.Email,
		user.PasswordHash,
		user.GradeID,
		user.Role,
		user.ManagerID,
	).Scan(&userID)

	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (r *userRepository) CheckEmailExists(ctx context.Context, tx pgx.Tx, email string) (bool, error) {
	var count int

	err := tx.QueryRow(
		ctx,
		`SELECT COUNT(*) FROM users WHERE email=$1`,
		email,
	).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *userRepository) GetRole(ctx context.Context, tx pgx.Tx, userID int64) (string, error) {
	var role string

	err := tx.QueryRow(
		ctx,
		`SELECT role FROM users WHERE id=$1`,
		userID,
	).Scan(&role)

	if err != nil {
		return "", err
	}

	return role, nil
}

func (r *userRepository) GetGrade(ctx context.Context, tx pgx.Tx, userID int64) (int64, error) {
	var gradeID int64

	err := tx.QueryRow(
		ctx,
		`SELECT grade_id FROM users WHERE id=$1`,
		userID,
	).Scan(&gradeID)

	if err == pgx.ErrNoRows {
		return 0, apperrors.ErrUserNotFound
	}
	if err != nil {
		return 0, err
	}

	return gradeID, nil
}
