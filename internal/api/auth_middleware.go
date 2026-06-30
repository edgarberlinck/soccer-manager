package api

import (
	"context"
	"errors"
	"manager/game/internal/config"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const userIDContextKey contextKey = "auth_user_id"

func NewJWTAuthMiddleware(cfg config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := parseUserIDFromBearerToken(r, cfg.AuthJWTSecret)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userIDContextKey, userID)))
		})
	}
}

func userIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDContextKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, false
	}

	return userID, true
}

func parseUserIDFromBearerToken(r *http.Request, jwtSecret string) (uuid.UUID, error) {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return uuid.Nil, errors.New("missing bearer token")
	}

	rawToken := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if rawToken == "" {
		return uuid.Nil, errors.New("missing token")
	}

	token, err := jwt.Parse(rawToken, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.New("invalid claims")
	}

	userIDRaw, ok := claims["sub"].(string)
	if !ok || strings.TrimSpace(userIDRaw) == "" {
		return uuid.Nil, errors.New("invalid subject")
	}

	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return uuid.Nil, errors.New("invalid subject")
	}

	return userID, nil
}