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

var (
	AdminID   = int64(1)
	ManagerID = int64(2)
)

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
	log.Println("RegisterUser started:", email)

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
		log.Println("Email already registered:", email)
		return apperrors.ErrEmailAlreadyRegistered
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		log.Println("Password hashing failed:", err)
		return apperrors.ErrPasswordHashFailed
	}

	// Decide role, grade, manager_id
	var role string
	var gradeID int64
	var managerID *int64

	switch email {
	case constants.AdminEmail:
		role = constants.RoleAdmin
		gradeID = 3
		managerID = nil

	case constants.ManagerEmail:
		role = constants.RoleManager
		gradeID = 2
		managerID = &AdminID

	default:
		role = constants.RoleEmployee
		gradeID = 1
		managerID = &ManagerID
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
		return apperrors.ErrInsertFailed
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

	log.Println("RegisterUser completed successfully:", email)
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
