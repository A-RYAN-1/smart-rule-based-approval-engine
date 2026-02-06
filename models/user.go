package models

import "time"

type User struct {
	ID           int64     `db:"id"`
	Name         string    `db:"name"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	GradeID      int64     `db:"grade_id"`
	Role         string    `db:"role"`
	ManagerID    *int64    `db:"manager_id"`
	CreatedAt    time.Time `db:"created_at"`
}
