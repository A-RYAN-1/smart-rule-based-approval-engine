package tests

import (
	"context"
	"testing"

	"github.com/ankita-advitot/rule_based_approval_engine/app/auth"
	"github.com/ankita-advitot/rule_based_approval_engine/app/auth/mocks"
	"github.com/ankita-advitot/rule_based_approval_engine/models"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthService_RegisterUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		userName      string
		email         string
		password      string
		mockSetup     func(u *mocks.UserRepository, b *mocks.BalanceRepository, db *mocks.DB, tx *mocks.Tx)
		expectedError error
	}{
		{
			name:     "Success",
			userName: "John Doe",
			email:    "john@example.com",
			password: "password123",
			mockSetup: func(u *mocks.UserRepository, b *mocks.BalanceRepository, db *mocks.DB, tx *mocks.Tx) {
				db.EXPECT().Begin(ctx).Return(tx, nil)
				u.EXPECT().CheckEmailExists(ctx, tx, "john@example.com").Return(false, nil)
				u.EXPECT().Create(ctx, tx, mock.AnythingOfType("*models.User")).Return(1, nil)
				b.EXPECT().InitializeBalances(ctx, tx, int64(1), int64(1)).Return(nil)
				tx.EXPECT().Commit(ctx).Return(nil)
				tx.EXPECT().Rollback(ctx).Return(nil).Maybe()
			},
			expectedError: nil,
		},
		{
			name:     "Email Already Exists",
			userName: "John Doe",
			email:    "john@example.com",
			password: "password123",
			mockSetup: func(u *mocks.UserRepository, b *mocks.BalanceRepository, db *mocks.DB, tx *mocks.Tx) {
				db.EXPECT().Begin(ctx).Return(tx, nil)
				u.EXPECT().CheckEmailExists(ctx, tx, "john@example.com").Return(true, nil)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			expectedError: apperrors.ErrEmailAlreadyRegistered,
		},
		{
			name:          "Empty Email",
			userName:      "John Doe",
			email:         "",
			password:      "password123",
			mockSetup:     func(u *mocks.UserRepository, b *mocks.BalanceRepository, db *mocks.DB, tx *mocks.Tx) {},
			expectedError: apperrors.ErrEmailRequired,
		},
		{
			name:          "Empty Password",
			userName:      "John Doe",
			email:         "john@example.com",
			password:      "",
			mockSetup:     func(u *mocks.UserRepository, b *mocks.BalanceRepository, db *mocks.DB, tx *mocks.Tx) {},
			expectedError: apperrors.ErrPasswordRequired,
		},
		{
			name:     "DB Transaction Error",
			userName: "John Doe",
			email:    "john@example.com",
			password: "password123",
			mockSetup: func(u *mocks.UserRepository, b *mocks.BalanceRepository, db *mocks.DB, tx *mocks.Tx) {
				db.EXPECT().Begin(ctx).Return(nil, apperrors.ErrTransactionBegin)
			},
			expectedError: apperrors.ErrTransactionBegin,
		},
		{
			name:     "CheckEmail Repository Error",
			userName: "John Doe",
			email:    "john@example.com",
			password: "password123",
			mockSetup: func(u *mocks.UserRepository, b *mocks.BalanceRepository, db *mocks.DB, tx *mocks.Tx) {
				db.EXPECT().Begin(ctx).Return(tx, nil)
				u.EXPECT().CheckEmailExists(ctx, tx, "john@example.com").Return(false, apperrors.ErrDatabase)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			expectedError: apperrors.ErrDatabase,
		},
		{
			name:     "Create User Repository Error",
			userName: "John Doe",
			email:    "john@example.com",
			password: "password123",
			mockSetup: func(u *mocks.UserRepository, b *mocks.BalanceRepository, db *mocks.DB, tx *mocks.Tx) {
				db.EXPECT().Begin(ctx).Return(tx, nil)
				u.EXPECT().CheckEmailExists(ctx, tx, "john@example.com").Return(false, nil)
				u.EXPECT().Create(ctx, tx, mock.Anything).Return(int64(0), apperrors.ErrInsertFailed)
				tx.EXPECT().Rollback(ctx).Return(nil)
			},
			expectedError: apperrors.ErrInsertFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := mocks.NewUserRepository(t)
			mockBalanceRepo := mocks.NewBalanceRepository(t)
			mockDB := mocks.NewDB(t)
			mockTx := mocks.NewTx(t)

			tt.mockSetup(mockUserRepo, mockBalanceRepo, mockDB, mockTx)

			service := auth.NewAuthService(ctx, mockUserRepo, mockBalanceRepo, mockDB)
			err := service.RegisterUser(ctx, tt.userName, tt.email, tt.password)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthService_LoginUser(t *testing.T) {
	ctx := context.Background()

	t.Run("User Not Found", func(t *testing.T) {
		mockUserRepo := mocks.NewUserRepository(t)
		mockUserRepo.EXPECT().GetByEmail(ctx, "non@existent.com").Return(nil, apperrors.ErrUserNotFound)

		service := auth.NewAuthService(ctx, mockUserRepo, nil, nil)
		_, _, err := service.LoginUser(ctx, "non@existent.com", "password")

		assert.ErrorIs(t, err, apperrors.ErrInvalidCredentials)
	})

	t.Run("Success", func(t *testing.T) {
		mockUserRepo := mocks.NewUserRepository(t)
		password := "password123"
		hashed, _ := utils.HashPassword(password)
		mockUserRepo.EXPECT().GetByEmail(ctx, "john@example.com").Return(&models.User{
			ID:           1,
			Email:        "john@example.com",
			PasswordHash: hashed,
			Role:         "ADMIN",
		}, nil)

		service := auth.NewAuthService(ctx, mockUserRepo, nil, nil)
		token, role, err := service.LoginUser(ctx, "john@example.com", password)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Equal(t, "ADMIN", role)
	})
}
