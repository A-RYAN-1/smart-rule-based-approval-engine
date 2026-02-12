package models

import "time"

type User struct {
	ID           int64     `db:"id" json:"id"`
	Name         string    `db:"name" json:"name"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	GradeID      int64     `db:"grade_id" json:"grade_id"`
	Role         string    `db:"role" json:"role"`
	ManagerID    *int64    `db:"manager_id" json:"manager_id"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
