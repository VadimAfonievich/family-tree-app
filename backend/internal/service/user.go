package service

import (
	"context"

	"family-tree-backend/internal/auth"
	"family-tree-backend/internal/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type UserService struct {
	queries *db.Queries
	jwt     *auth.JWTService
	tgToken string
	log     zerolog.Logger
}

func NewUserService(pool *pgxpool.Pool, jwt *auth.JWTService, tgToken string, log zerolog.Logger) *UserService {
	return &UserService{
		queries: db.New(pool),
		jwt:     jwt,
		tgToken: tgToken,
		log:     log,
	}
}

func (s *UserService) AuthByTelegram(ctx context.Context, initData string) (string, error) {
	data, err := auth.ValidateInitData(initData, s.tgToken)
	if err != nil {
		s.log.Warn().Err(err).Msg("invalid telegram initData")
		return "", err
	}

	telegramID, username, err := auth.ParseTelegramUser(data["user"])
	if err != nil {
		return "", err
	}

	// Create or get user
	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		TelegramID: telegramID,
		Username:   username,
	})
	if err != nil {
		return "", err
	}

	// Generate JWT
	token, err := s.jwt.GenerateToken(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (db.User, error) {
	return s.queries.GetUserByID(ctx, id)
}
