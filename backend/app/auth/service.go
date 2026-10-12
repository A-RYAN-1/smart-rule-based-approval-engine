package auth

import (
	"context"
	"log"
	"strings"

	"github.com/ankita-advitot/rule_based_approval_engine/constants"
	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/models"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/utils"
)

// ManagerID is the id of the seeded manager account that new employees
// report to (0 when no manager exists yet). seed.Accounts sets it on startup.
var ManagerID int64

// handles authentication and user registration business logic
type AuthService struct {
	userRepo    interfaces.UserRepository
	balanceRepo interfaces.BalanceRepository
	db          interfaces.DB
}

// NewAuthService creates a new instance of AuthService
func NewAuthService(ctx context.Context, userRepo interfaces.UserRepository, balanceRepo interfaces.BalanceRepository, db interfaces.DB) interfaces.AuthService {
	return &AuthService{
		userRepo:    userRepo,
		balanceRepo: balanceRepo,
		db:          db,
	}
}

// RegisterUser registers a new user
func (s *AuthService) RegisterUser(ctx context.Context, name, email, password string) error {

	if strings.TrimSpace(email) == "" {
		log.Println("Validation failed: email empty")
		return apperrors.ErrEmailRequired
	}
	if strings.TrimSpace(password) == "" {
		log.Println("Validation failed: password empty")
		return apperrors.ErrPasswordRequired
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return apperrors.ErrTransactionBegin
	}
	defer tx.Rollback(ctx)

	// Check email uniqueness
	exists, err := s.userRepo.CheckEmailExists(ctx, tx, email)
	if err != nil {
		return apperrors.ErrDatabase
	}
	if exists {
		log.Println("register: email already registered")
		return apperrors.ErrEmailAlreadyRegistered
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		log.Println("Password hashing failed:", err)
		return apperrors.ErrPasswordHashFailed
	}

	// Self-registration always creates an employee. The admin and manager
	// accounts are created only by the startup seed, never by sign-up.
	if isReservedEmail(email) {
		return apperrors.ErrEmailAlreadyRegistered
	}
	role := constants.RoleEmployee
	gradeID := int64(1)
	var managerID *int64
	if ManagerID != 0 {
		id := ManagerID
		managerID = &id
	}

	log.Printf("Role decided: role=%s grade=%d managerID=%v\n", role, gradeID, managerID)

	// Insert user
	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: hashedPassword,
		GradeID:      gradeID,
		Role:         role,
		ManagerID:    managerID,
	}

	userID, err := s.userRepo.Create(ctx, tx, user)
	if err != nil {
		log.Printf("CREATE USER FAILED: %+v\n", err)
		return err
	}

	log.Println("User inserted successfully, userID:", userID)

	// Initialize balances ONLY for employee & manager
	if role != constants.RoleAdmin {
		log.Println("Initializing balances for user:", userID)

		err = s.balanceRepo.InitializeBalances(ctx, tx, userID, gradeID)
		if err != nil {
			return apperrors.ErrBalanceUpdateFailed
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return apperrors.ErrTransactionCommit
	}

	log.Println("register: user created, id", userID)
	return nil
}

// LoginUser authenticates a user and returns a JWT token
func (s *AuthService) LoginUser(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", apperrors.ErrInvalidCredentials
	}

	if err := utils.CheckPassword(password, user.PasswordHash); err != nil {
		return "", "", apperrors.ErrInvalidCredentials
	}

	token, err := utils.GenerateToken(user.ID, user.Role)
	if err != nil {
		return "", "", apperrors.ErrOperationFailed
	}

	return token, user.Role, nil
}

// GetUserByEmail fetches a user by email
func (s *AuthService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}

// GetUserByID fetches a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func isReservedEmail(email string) bool {
	e := strings.ToLower(strings.TrimSpace(email))
	return e == constants.AdminEmail || e == constants.ManagerEmail
}
