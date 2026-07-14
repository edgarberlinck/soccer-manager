package api

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"manager/game/internal/config"
	repository "manager/game/internal/infrastructure/database/generated"
	"manager/game/internal/infrastructure/mailer"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	resend "github.com/resend/resend-go/v2"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	queries *repository.Queries
	cfg     config.Config
}

type signUpRequest struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	ManagerName   string `json:"manager_name"`
	ClubName      string `json:"club_name"`
	ClubShortName string `json:"club_short_name"`
	Abbreviation  string `json:"abbreviation"`
	Continent     string `json:"continent"`
	Country       string `json:"country"`
}

type signInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}

type link struct {
	Href   string `json:"href"`
	Method string `json:"method"`
}

type authLinks struct {
	Links map[string]link `json:"_links"`
}

type meResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	ManagerName string    `json:"manager_name"`
}

func NewAuthHandler(queries *repository.Queries, cfg config.Config) *AuthHandler {
	return &AuthHandler{
		queries: queries,
		cfg:     cfg,
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

	req.ManagerName = strings.TrimSpace(req.ManagerName)
	req.ClubName = strings.TrimSpace(req.ClubName)
	req.ClubShortName = strings.TrimSpace(req.ClubShortName)
	req.Abbreviation = strings.ToUpper(strings.TrimSpace(req.Abbreviation))
	req.Continent = strings.TrimSpace(req.Continent)
	req.Country = strings.TrimSpace(req.Country)

	if req.ManagerName == "" || req.ClubName == "" {
		http.Error(w, "manager_name and club_name are required", http.StatusBadRequest)
		return
	}

	if req.ClubShortName == "" {
		req.ClubShortName = req.ClubName
	}

	if req.Abbreviation == "" {
		req.Abbreviation = defaultAbbreviation(req.ClubName)
	}

	if req.Continent == "" {
		req.Continent = "Europe"
	}

	if req.Country == "" {
		req.Country = "Portugal"
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

	newUser, err := h.queries.CreateUser(r.Context(), repository.CreateUserParams{
		ID:                         uuid.New(),
		Username:                   email,
		PasswordHash:               string(hashedPassword),
		Active:                     true,
		VerificationToken:          sql.NullString{},
		VerificationTokenExpiresAt: sql.NullTime{},
	})
	if err != nil {
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	_, err = h.queries.UpsertUserMeta(r.Context(), repository.UpsertUserMetaParams{
		UserID:      newUser.ID,
		FullName:    sql.NullString{String: req.ManagerName, Valid: true},
		Country:     sql.NullString{String: req.Country, Valid: true},
		SocialLinks: json.RawMessage(`{}`),
		Metadata:    json.RawMessage(`{}`),
	})
	if err != nil {
		http.Error(w, "failed to create user profile", http.StatusInternalServerError)
		return
	}

	club, err := h.queries.CreateClub(r.Context(), repository.CreateClubParams{
		ID:           uuid.New(),
		UserID:       newUser.ID,
		Name:         req.ClubName,
		ShortName:    sql.NullString{String: req.ClubShortName, Valid: true},
		Abbreviation: sql.NullString{String: req.Abbreviation, Valid: true},
		Continent:    sql.NullString{String: req.Continent, Valid: true},
		Country:      sql.NullString{String: req.Country, Valid: true},
	})
	if err != nil {
		http.Error(w, "failed to create club", http.StatusInternalServerError)
		return
	}

	if err := bootstrapStarterSquad(r.Context(), h.queries, club.ID); err != nil {
		http.Error(w, "failed to bootstrap squad", http.StatusInternalServerError)
		return
	}

	token, err := h.issueToken(newUser.ID, newUser.Username)
	if err != nil {
		http.Error(w, "failed to create token", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{
		"token": token,
		"club": map[string]any{
			"id":   club.ID,
			"name": club.Name,
		},
		"_links": h.buildAuthLinks(),
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

	if _, _, err := ensureUserHasClubAndSquad(r.Context(), h.queries, user.ID, user.Username); err != nil {
		http.Error(w, "failed to ensure club", http.StatusInternalServerError)
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
	userID, err := parseUserIDFromBearerToken(r, h.cfg.AuthJWTSecret)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.queries.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	managerName := ""
	if meta, err := h.queries.GetUserMetaByUserID(r.Context(), user.ID); err == nil && meta.FullName.Valid {
		managerName = meta.FullName.String
	}

	respondJSON(w, http.StatusOK, meResponse{ID: user.ID, Email: user.Username, ManagerName: managerName})
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

func (h *AuthHandler) sendVerificationEmail(email, token string) error {
	if h.cfg.ResendAPIKey == "" || h.cfg.ResendFromEmail == "" || h.cfg.AppBaseURL == "" {
		return errors.New("resend/app configuration missing")
	}

	verifyURL := fmt.Sprintf("%s/auth/verify?token=%s", strings.TrimSuffix(h.cfg.AppBaseURL, "/"), token)
	htmlBody, err := mailer.RenderVerificationEmail(mailer.VerificationTemplateData{
		AppName:   "Soccer Manager",
		VerifyURL: verifyURL,
	})
	if err != nil {
		return err
	}

	client := resend.NewClient(h.cfg.ResendAPIKey)
	payload := &resend.SendEmailRequest{
		From:    h.cfg.ResendFromEmail,
		To:      []string{email},
		Subject: "Verify your Soccer Manager account",
		Html:    htmlBody,
	}

	if _, err := client.Emails.SendWithContext(context.Background(), payload); err != nil {
		return fmt.Errorf("failed to send email with resend: %w", err)
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

func defaultAbbreviation(clubName string) string {
	parts := strings.Fields(strings.ToUpper(strings.TrimSpace(clubName)))
	if len(parts) == 0 {
		return "CLB"
	}

	abbr := ""
	for _, p := range parts {
		if p == "" {
			continue
		}
		abbr += string(p[0])
		if len(abbr) == 3 {
			return abbr
		}
	}

	for len(abbr) < 3 {
		abbr += "X"
	}

	return abbr
}
