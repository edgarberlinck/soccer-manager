package api

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"manager/game/internal/config"
	repository "manager/game/internal/infrastructure/database/generated"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	queries *repository.Queries
	cfg     config.Config
	http    *http.Client
}

type signUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}

type resendSendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

type link struct {
	Href   string `json:"href"`
	Method string `json:"method"`
}

type authLinks struct {
	Links map[string]link `json:"_links"`
}

type meResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

func NewAuthHandler(queries *repository.Queries, cfg config.Config) *AuthHandler {
	return &AuthHandler{
		queries: queries,
		cfg:     cfg,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (h *AuthHandler) HATEOAS(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, authLinks{Links: h.buildAuthLinks()})
}

func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	if h.queries == nil {
		http.Error(w, "database queries not initialized", http.StatusInternalServerError)
		return
	}

	var req signUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	email, err := normalizeAndValidateEmail(req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.Password) < 8 {
		http.Error(w, "password must have at least 8 characters", http.StatusBadRequest)
		return
	}

	_, err = h.queries.GetUserByEmail(r.Context(), email)
	if err == nil {
		http.Error(w, "email already registered", http.StatusConflict)
		return
	}
	if !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "failed to validate user", http.StatusInternalServerError)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "failed to secure password", http.StatusInternalServerError)
		return
	}

	verificationToken, err := generateToken(32)
	if err != nil {
		http.Error(w, "failed to generate verification token", http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(time.Duration(h.cfg.AuthVerifyTokenTTLMinutes) * time.Minute)
	_, err = h.queries.CreateUser(r.Context(), repository.CreateUserParams{
		ID:                         uuid.New(),
		Username:                   email,
		PasswordHash:               string(hashedPassword),
		Active:                     false,
		VerificationToken:          sql.NullString{String: verificationToken, Valid: true},
		VerificationTokenExpiresAt: sql.NullTime{Time: expiresAt, Valid: true},
	})
	if err != nil {
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	if err := h.sendVerificationEmail(email, verificationToken); err != nil {
		http.Error(w, "failed to send verification email", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{
		"message": "signup successful, please verify your email before signing in",
		"_links":  h.buildAuthLinks(),
	})
}

func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	if h.queries == nil {
		http.Error(w, "database queries not initialized", http.StatusInternalServerError)
		return
	}

	var req signInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	email, err := normalizeAndValidateEmail(req.Email)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	user, err := h.queries.GetUserByEmail(r.Context(), email)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if !user.Active {
		http.Error(w, "email not verified yet", http.StatusForbidden)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := h.issueToken(user.ID, user.Username)
	if err != nil {
		http.Error(w, "failed to create token", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, authResponse{Token: token})
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	_, err := h.queries.VerifyUserByToken(r.Context(), sql.NullString{String: token, Valid: true})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "invalid or expired token", http.StatusBadRequest)
			return
		}

		http.Error(w, "failed to verify user", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "email verified successfully"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, err := h.parseBearerClaims(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userIDRaw, ok := claims["sub"].(string)
	if !ok || strings.TrimSpace(userIDRaw) == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.queries.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	respondJSON(w, http.StatusOK, meResponse{ID: user.ID, Email: user.Username})
}

func (h *AuthHandler) issueToken(userID uuid.UUID, email string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(h.cfg.AuthJWTExpirationMinutes) * time.Minute)

	claims := jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
		"iat":   now.Unix(),
		"exp":   expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.AuthJWTSecret))
}

func (h *AuthHandler) parseBearerClaims(r *http.Request) (jwt.MapClaims, error) {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, errors.New("missing bearer token")
	}

	rawToken := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if rawToken == "" {
		return nil, errors.New("missing token")
	}

	token, err := jwt.Parse(rawToken, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(h.cfg.AuthJWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	return claims, nil
}

func (h *AuthHandler) sendVerificationEmail(email, token string) error {
	if h.cfg.ResendAPIKey == "" || h.cfg.ResendFromEmail == "" || h.cfg.AppBaseURL == "" {
		return errors.New("resend/app configuration missing")
	}

	verifyURL := fmt.Sprintf("%s/auth/verify?token=%s", strings.TrimSuffix(h.cfg.AppBaseURL, "/"), token)
	htmlBody := fmt.Sprintf("<p>Welcome to Soccer Manager.</p><p>Confirm your account by clicking <a href=\"%s\">this verification link</a>.</p>", verifyURL)

	payload := resendSendEmailRequest{
		From:    h.cfg.ResendFromEmail,
		To:      []string{email},
		Subject: "Verify your Soccer Manager account",
		HTML:    htmlBody,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+h.cfg.ResendAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("resend returned status %d", resp.StatusCode)
	}

	return nil
}

func (h *AuthHandler) buildAuthLinks() map[string]link {
	baseURL := strings.TrimSuffix(h.cfg.AppBaseURL, "/")
	if baseURL == "" {
		baseURL = ""
	}

	return map[string]link{
		"self": {
			Href:   baseURL + "/auth",
			Method: http.MethodGet,
		},
		"signup": {
			Href:   baseURL + "/auth/signup",
			Method: http.MethodPost,
		},
		"signin": {
			Href:   baseURL + "/auth/signin",
			Method: http.MethodPost,
		},
		"me": {
			Href:   baseURL + "/auth/me",
			Method: http.MethodGet,
		},
		"verify": {
			Href:   baseURL + "/auth/verify?token={token}",
			Method: http.MethodGet,
		},
	}
}

func normalizeAndValidateEmail(email string) (string, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if normalizedEmail == "" {
		return "", errors.New("email is required")
	}

	if _, err := mail.ParseAddress(normalizedEmail); err != nil {
		return "", errors.New("invalid email format")
	}

	return normalizedEmail, nil
}

func generateToken(size int) (string, error) {
	randomBytes := make([]byte, size)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
