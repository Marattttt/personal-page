package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/Marattttt/personal-page/authorizer/pkg/config"
	"github.com/Marattttt/personal-page/authorizer/pkg/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrNotAuthorized = errors.New("Not authorized")

type TokenInvalidError struct {
	Cause error
}

func (t TokenInvalidError) Error() string {
	if t.Cause != nil {
		return fmt.Sprintf("invalid token: %s", t.Cause.Error())
	}

	return "invalid token"
}

func (t TokenInvalidError) Unwrap() error {
	return t.Cause
}

const (
	MethodSha256 = "sha256"
	MethodSHA512 = "sha512"
)

type UserRepo interface {
	Get(ctx context.Context, id int) (*models.User, error)
}

type RefreshValidator interface {
	ValidateRefresh(jwt.Token) error
}

type Auth struct {
	repo      UserRepo
	conf      config.AuthConfig
	logger    slog.Logger
	refresher RefreshValidator
}

func New(repo UserRepo, conf config.AuthConfig, refresher RefreshValidator, logger slog.Logger) Auth {
	return Auth{
		repo:      repo,
		conf:      conf,
		logger:    logger,
		refresher: refresher,
	}
}

func (a Auth) VerifyAccess(tokStr string) (*jwt.Token, error) {
	return verifyToken(tokStr, a.conf.AccessSecret)
}

func (a Auth) VerifyRefresh(tokStr string) (*jwt.Token, error) {
	tok, err := verifyToken(tokStr, a.conf.RefreshSecret)

	if err != nil {
		return nil, err
	}

	if err := a.refresher.ValidateRefresh(*tok); err != nil {
		return nil, err
	}

	return tok, nil
}

func verifyToken(tokStr string, secret string) (*jwt.Token, error) {
	tok, err := jwt.Parse(tokStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Method.Alg())
		}

		return []byte(secret), nil

	})

	if err != nil {
		return nil, err
	}

	if !tok.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return tok, nil
}

func GenerateAccess(user models.User, conf *config.AuthConfig) (*string, error) {
	claims := jwt.MapClaims{
		"iss": conf.Issuer,
		"sub": strconv.Itoa(user.Id),
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(conf.AccessValidTime).Unix(),
	}

	tok := jwt.NewWithClaims(&jwt.SigningMethodHMAC{}, claims)

	signed, err := tok.SignedString(conf.AccessSecret)
	if err != nil {
		return nil, fmt.Errorf("singing: %w", err)
	}

	return &signed, nil
}

func GenerateRefresh(user models.User, conf *config.AuthConfig) (*string, error) {
	claims := jwt.MapClaims{
		"iss": conf.Issuer,
		"sub": strconv.Itoa(user.Id),
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(conf.RefreshValidTime).Unix(),

		"jti": uuid.New().String(),
	}

	tok := jwt.NewWithClaims(&jwt.SigningMethodHMAC{}, claims)

	signed, err := tok.SignedString(conf.RefreshSecret)
	if err != nil {
		return nil, fmt.Errorf("singing: %w", err)
	}

	return &signed, nil
}

func GeneratePair(user models.User, conf *config.AuthConfig) (*string, *string, error) {
	access, err := GenerateAccess(user, conf)
	if err != nil {
		return nil, nil, fmt.Errorf("access: %w", err)
	}

	refresh, err := GenerateRefresh(user, conf)
	if err != nil {
		return nil, nil, fmt.Errorf("refresh: %w", err)
	}

	return access, refresh, nil
}

func HashPassword(pass []byte) (*[]byte, error) {
	hashed, err := bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("generating: %w", err)
	}

	return &hashed, nil
}
