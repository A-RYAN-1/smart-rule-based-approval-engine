package seed

import (
	"context"
	"log"

	"github.com/ankita-advitot/rule_based_approval_engine/config"
	"github.com/ankita-advitot/rule_based_approval_engine/constants"
	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
)

// Login is one demo account shown on the login page.
type Login struct {
	Role     string `json:"role"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// DemoLogins lists the seed accounts whose password is set. Only served
// when DEMO_MODE is on: the passwords are then public by design.
func DemoLogins(cfg *config.Config) []Login {
	var out []Login
	for _, l := range []Login{
		{constants.RoleAdmin, constants.AdminEmail, cfg.SeedAdminPassword},
		{constants.RoleManager, constants.ManagerEmail, cfg.SeedManagerPassword},
		{constants.RoleEmployee, constants.EmployeeEmail, cfg.SeedEmployeePassword},
	} {
		if l.Password != "" {
			out = append(out, l)
		}
	}
	return out
}

// defaultRules is the rule set the migrations install, per grade.
const defaultRules = `
INSERT INTO rules (request_type, condition, action, grade_id)
SELECT r.request_type::request_type_enum, r.condition::jsonb, 'AUTO_APPROVE', g.id
FROM (VALUES
  ('LEAVE',    '{"max_days": 3}',      'STANDARD'),
  ('EXPENSE',  '{"max_amount": 5000}', 'STANDARD'),
  ('DISCOUNT', '{"max_percent": 10}',  'STANDARD'),
  ('LEAVE',    '{"max_days": 5}',      'Grade 2'),
  ('LEAVE',    '{"max_days": 10}',     'Grade 3'),
  ('EXPENSE',  '{"max_amount": 10000}', 'Grade 2'),
  ('EXPENSE',  '{"max_amount": 20000}', 'Grade 3'),
  ('DISCOUNT', '{"max_percent": 20}',  'Grade 2'),
  ('DISCOUNT', '{"max_percent": 30}',  'Grade 3')
) AS r(request_type, condition, grade_name)
JOIN grades g ON g.name = r.grade_name`

// ResetDemoRules puts the default rules back, so whatever a visitor did
// with the public admin login is undone by the next daily cron run.
func ResetDemoRules(ctx context.Context, db interfaces.DB) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "DELETE FROM rules"); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, defaultRules); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	log.Println("seed: demo rules reset to the defaults")
	return nil
}
