package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	domainuser "acc-dp/producer/internal/domain/user"
	"acc-dp/producer/internal/service/auth"
	usersvc "acc-dp/producer/internal/service/user"
)

type contextKey string

const authenticatedContextKey contextKey = "authenticated_user"

type Config struct {
	CookieName    string
	CookieSecure  bool
	CookieDomain  string
	SessionTTL    time.Duration
	AllowedOrigin string
}

type Server struct {
	mux          *http.ServeMux
	authService  *auth.Service
	userService  *usersvc.Service
	cookieName   string
	cookieSecure bool
	cookieDomain string
	sessionTTL   time.Duration
}

type AuthContext struct {
	User    *domainuser.User
	Session *domainuser.Session
	Token   string
}

func NewServer(authService *auth.Service, userService *usersvc.Service, cfg Config) *Server {
	server := &Server{
		mux:          http.NewServeMux(),
		authService:  authService,
		userService:  userService,
		cookieName:   fallbackString(cfg.CookieName, "accdp_session"),
		cookieSecure: cfg.CookieSecure,
		cookieDomain: cfg.CookieDomain,
		sessionTTL:   fallbackDuration(cfg.SessionTTL, 24*time.Hour),
	}

	server.registerRoutes()

	return server
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("POST /auth/register", s.handleRegister)
	s.mux.HandleFunc("POST /auth/login", s.handleLogin)
	s.mux.Handle("POST /auth/logout", s.requireAuth(http.HandlerFunc(s.handleLogout)))
	s.mux.Handle("GET /auth/me", s.requireAuth(http.HandlerFunc(s.handleMe)))
	s.mux.Handle("GET /users", s.requireAuth(http.HandlerFunc(s.handleListUsers)))
	s.mux.Handle("POST /users/active", s.requireAuth(http.HandlerFunc(s.handleSetActiveUser)))
	s.mux.Handle("GET /users/active", s.requireAuth(http.HandlerFunc(s.handleGetActiveUser)))
	s.mux.Handle("GET /users/active/all", s.requireAuth(http.HandlerFunc(s.handleListActiveUsers)))
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	respondJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"time":   time.Now().UTC(),
	})
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Name     string `json:"name"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := decodeJSON(r, &body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := s.authService.Register(r.Context(), body.Username, body.Name, body.Password, body.Role)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidInput) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusBadRequest, "failed to register user")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{"user": sanitizeUser(created)})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := decodeJSON(r, &body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := s.authService.Login(r.Context(), body.Username, body.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidInput):
			respondError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, auth.ErrInvalidCredentials):
			respondError(w, http.StatusUnauthorized, err.Error())
		default:
			respondError(w, http.StatusUnauthorized, "login failed")
		}
		return
	}

	s.setSessionCookie(w, result.Token)
	respondJSON(w, http.StatusOK, map[string]any{"user": sanitizeUser(result.User)})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	authContext, ok := authFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	if err := s.authService.Logout(r.Context(), authContext.Session.ID); err != nil {
		respondError(w, http.StatusInternalServerError, "logout failed")
		return
	}

	s.clearSessionCookie(w)
	respondJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	authContext, ok := authFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"user": sanitizeUser(authContext.User)})
}

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.userService.ListUsers(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	publicUsers := make([]any, 0, len(users))
	for i := range users {
		publicUsers = append(publicUsers, sanitizeUser(&users[i]))
	}

	respondJSON(w, http.StatusOK, map[string]any{"users": publicUsers})
}

func (s *Server) handleSetActiveUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		UserID    string `json:"user_id"`
		MachineID string `json:"machine_id"`
	}

	if err := decodeJSON(r, &body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	machineID := strings.TrimSpace(body.MachineID)
	if err := s.userService.SetActiveUserForMachine(r.Context(), machineID, body.UserID); err != nil {
		switch {
		case errors.Is(err, usersvc.ErrInvalidUserInput):
			respondError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, usersvc.ErrUserNotFound):
			respondError(w, http.StatusNotFound, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to set active user")
		}
		return
	}

	if machineID == "" {
		machineID = usersvc.DefaultMachineID
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status":     "active_user_updated",
		"machine_id": machineID,
	})
}

func (s *Server) handleGetActiveUser(w http.ResponseWriter, r *http.Request) {
	machineID := strings.TrimSpace(r.URL.Query().Get("machine_id"))
	active, err := s.userService.GetActiveUserForMachine(r.Context(), machineID)
	if err != nil {
		switch {
		case errors.Is(err, usersvc.ErrActiveUserNotSet):
			respondError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, usersvc.ErrUserNotFound):
			respondError(w, http.StatusNotFound, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to get active user")
		}
		return
	}

	if machineID == "" {
		machineID = usersvc.DefaultMachineID
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"machine_id": machineID,
		"user":       sanitizeUser(active),
	})
}

func (s *Server) handleListActiveUsers(w http.ResponseWriter, r *http.Request) {
	assignments, err := s.userService.ListActiveUsers(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list active users")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"active_users": assignments})
}

func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(readBearerToken(r))
		if token == "" {
			token = readCookieToken(r, s.cookieName)
		}
		if token == "" {
			respondError(w, http.StatusUnauthorized, "missing authentication token")
			return
		}

		user, session, err := s.authService.Authenticate(r.Context(), token)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "invalid or expired session")
			return
		}

		next.ServeHTTP(w, r.WithContext(withAuthContext(r.Context(), &AuthContext{
			User:    user,
			Session: session,
			Token:   token,
		})))
	})
}

func (s *Server) setSessionCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     s.cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(s.sessionTTL),
	}

	if s.cookieDomain != "" {
		cookie.Domain = s.cookieDomain
	}

	http.SetCookie(w, cookie)
}

func (s *Server) clearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     s.cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	}

	if s.cookieDomain != "" {
		cookie.Domain = s.cookieDomain
	}

	http.SetCookie(w, cookie)
}

func sanitizeUser(user *domainuser.User) map[string]any {
	if user == nil {
		return map[string]any{}
	}

	return map[string]any{
		"id":         user.ID,
		"username":   user.Username,
		"name":       user.Name,
		"role":       user.Role,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return err
	}

	if decoder.More() {
		return fmt.Errorf("unexpected additional content")
	}

	return nil
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func readBearerToken(r *http.Request) string {
	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	if authorization == "" {
		return ""
	}

	parts := strings.SplitN(authorization, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}

func readCookieToken(r *http.Request, cookieName string) string {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(cookie.Value)
}

func withAuthContext(ctx context.Context, authContext *AuthContext) context.Context {
	return context.WithValue(ctx, authenticatedContextKey, authContext)
}

func authFromContext(ctx context.Context) (*AuthContext, bool) {
	value := ctx.Value(authenticatedContextKey)
	authContext, ok := value.(*AuthContext)
	if !ok || authContext == nil {
		return nil, false
	}

	return authContext, true
}

func fallbackString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func fallbackDuration(value, fallback time.Duration) time.Duration {
	if value <= 0 {
		return fallback
	}
	return value
}
