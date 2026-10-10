package services

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/apperrors"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/common"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/models"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/repositories"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/resources"
	"go.uber.org/zap"
	"golang.org/x/crypto/argon2"
)

type UserService struct {
	users  *repositories.UserRepository
	logger *zap.SugaredLogger
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

const (
	InvalidCredentials  = "invalid_credentials"
	InvalidRefreshToken = "invalid_refresh_token"
)

func NewUserService(db *sql.DB, logger *zap.SugaredLogger) *UserService {
	return &UserService{
		users:  repositories.NewUserRepository(db),
		logger: logger,
	}
}

func (s *UserService) Register(ctx context.Context, login, password string) (TokenPair, error) {
	if login == "" || password == "" {
		return TokenPair{}, apperrors.BadRequest(InvalidCredentials)
	}

	hash, err := s.hashPassword(password)
	if err != nil {
		s.logger.Errorw("hash password", "error", err)
		return TokenPair{}, apperrors.Internal(apperrors.InternalServerError)
	}

	created, err := s.users.Create(ctx, models.User{Login: login, Password: hash})
	if err != nil {
		return TokenPair{}, mapRepositoryError(s.logger, err, "user")
	}

	return s.tokenPair(created.ID)
}

func (s *UserService) Login(ctx context.Context, login, password string) (TokenPair, error) {
	if login == "" || password == "" {
		return TokenPair{}, apperrors.BadRequest(InvalidCredentials)
	}

	user, err := s.users.GetByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return TokenPair{}, apperrors.Unauthorized(InvalidCredentials)
		}
		return TokenPair{}, mapRepositoryError(s.logger, err, "user")
	}

	if !s.verifyPassword(user.Password, password) {
		return TokenPair{}, apperrors.Unauthorized(InvalidCredentials)
	}

	return s.tokenPair(user.ID)
}

func (s *UserService) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	userID, err := common.UserIDFromToken(refreshToken, common.TokenTypeRefresh)
	if err != nil {
		return TokenPair{}, apperrors.Unauthorized(InvalidRefreshToken)
	}

	if _, err := s.users.GetByID(ctx, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return TokenPair{}, apperrors.Unauthorized(InvalidRefreshToken)
		}
		return TokenPair{}, mapRepositoryError(s.logger, err, "user")
	}

	return s.tokenPair(userID)
}

func (s *UserService) GetBalance(ctx context.Context, userID uuid.UUID) (models.UserBalance, error) {
	balance, err := s.users.GetBalance(ctx, userID)
	if err != nil {
		return models.UserBalance{}, mapRepositoryError(s.logger, err, "user")
	}

	return balance, nil
}

func (s *UserService) tokenPair(userID uuid.UUID) (TokenPair, error) {
	access, err := common.Sign(userID, common.TokenTypeAccess, resources.AccessTTL)
	if err != nil {
		s.logger.Errorw("sign access token", "error", err)
		return TokenPair{}, apperrors.Internal(apperrors.InternalServerError)
	}

	refresh, err := common.Sign(userID, common.TokenTypeRefresh, resources.RefreshTTL)
	if err != nil {
		s.logger.Errorw("sign refresh token", "error", err)
		return TokenPair{}, apperrors.Internal(apperrors.InternalServerError)
	}

	return TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

func (s *UserService) hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return base64.RawStdEncoding.EncodeToString(salt) + "$" + base64.RawStdEncoding.EncodeToString(hash), nil
}

func (s *UserService) verifyPassword(encoded, password string) bool {
	saltPart, hashPart, ok := strings.Cut(encoded, "$")
	if !ok {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(saltPart)
	if err != nil {
		return false
	}
	expected, err := base64.RawStdEncoding.DecodeString(hashPart)
	if err != nil {
		return false
	}

	actual := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return subtle.ConstantTimeCompare(actual, expected) == 1
}
