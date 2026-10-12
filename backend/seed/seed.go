// Package seed creates the admin and manager accounts on startup. Sign-up
// only ever creates employees, so these two roles exist only through here.
package seed

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/ankita-advitot/rule_based_approval_engine/app/auth"
	"github.com/ankita-advitot/rule_based_approval_engine/config"
	"github.com/ankita-advitot/rule_based_approval_engine/constants"
	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/utils"
	"github.com/jackc/pgx/v5"
)

type account struct {
	name, email, role, grade, password string
	managerEmail                       string
	withBalances                       bool
}

// Accounts creates the admin and the manager when their password is set in
// the config and the account does not exist yet, then points new sign-ups
// at the manager. Existing accounts are never changed.
func Accounts(ctx context.Context, db interfaces.DB, balances interfaces.BalanceRepository, cfg *config.Config) error {
	accounts := []account{
		{name: "Admin", email: constants.AdminEmail, role: constants.RoleAdmin, grade: "Grade 3",
			password: cfg.SeedAdminPassword},
		{name: "Manager", email: constants.ManagerEmail, role: constants.RoleManager, grade: "Grade 2",
			password: cfg.SeedManagerPassword, managerEmail: constants.AdminEmail, withBalances: true},
	}
	for _, a := range accounts {
		if a.password == "" {
			continue
		}
		if err := create(ctx, db, balances, a); err != nil {
			return fmt.Errorf("seeding %s: %w", a.role, err)
		}
	}

	id, err := userID(ctx, db, constants.ManagerEmail)
	if err != nil {
		return err
	}
	auth.ManagerID = id
	if id == 0 {
		log.Println("seed: no manager account; new employees will have no approver until SEED_MANAGER_PASSWORD is set")
	}
	return nil
}

func create(ctx context.Context, db interfaces.DB, balances interfaces.BalanceRepository, a account) error {
	if id, err := userID(ctx, db, a.email); err != nil || id != 0 {
		return err
	}

	hash, err := utils.HashPassword(a.password)
	if err != nil {
		return err
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var managerID *int64
	if a.managerEmail != "" {
		var id int64
		err := tx.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", a.managerEmail).Scan(&id)
		if err == nil {
			managerID = &id
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}

	var newID, gradeID int64
	if err := tx.QueryRow(ctx, `
		INSERT INTO users (name, email, password_hash, grade_id, role, manager_id)
		SELECT $1, $2, $3, g.id, $4::user_role, $5 FROM grades g WHERE g.name = $6
		RETURNING id, grade_id`,
		a.name, a.email, hash, a.role, managerID, a.grade,
	).Scan(&newID, &gradeID); err != nil {
		return err
	}
	if a.withBalances {
		if err := balances.InitializeBalances(ctx, tx, newID, gradeID); err != nil {
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	log.Printf("seed: created %s account, id %d", a.role, newID)
	return nil
}

func userID(ctx context.Context, db interfaces.DB, email string) (int64, error) {
	var id int64
	err := db.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", email).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return id, err
}
