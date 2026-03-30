package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	domainmachine "acc-dp/producer/internal/domain/machine"
	domainuser "acc-dp/producer/internal/domain/user"
	"acc-dp/producer/internal/service/auth"
	machinesvc "acc-dp/producer/internal/service/machine"
	usersvc "acc-dp/producer/internal/service/user"
)

type contextKey string

const authenticatedContextKey contextKey = "authenticated_user"
const authenticatedMachineContextKey contextKey = "authenticated_machine"

type Config struct {
	CookieName    string
	CookieSecure  bool
	CookieDomain  string
	SessionTTL    time.Duration
	AllowedOrigin string
}

type Server struct {
	mux            *http.ServeMux
	authService    *auth.Service
	userService    *usersvc.Service
	machineService *machinesvc.Service
	cookieName     string
	cookieSecure   bool
	cookieDomain   string
	sessionTTL     time.Duration
}

type AuthContext struct {
	User    *domainuser.User
	Session *domainuser.Session
	Token   string
}

type MachineAuthContext struct {
	Machine *domainmachine.Machine
	Token   string
}

func NewServer(authService *auth.Service, userService *usersvc.Service, machineService *machinesvc.Service, cfg Config) *Server {
	server := &Server{
		mux:            http.NewServeMux(),
		authService:    authService,
		userService:    userService,
		machineService: machineService,
		cookieName:     fallbackString(cfg.CookieName, "accdp_session"),
		cookieSecure:   cfg.CookieSecure,
		cookieDomain:   cfg.CookieDomain,
		sessionTTL:     fallbackDuration(cfg.SessionTTL, 24*time.Hour),
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
	s.mux.HandleFunc("POST /machines/register", s.handleRegisterMachine)
	s.mux.Handle("POST /auth/logout", s.requireAuth(http.HandlerFunc(s.handleLogout)))
	s.mux.Handle("GET /auth/me", s.requireAuth(http.HandlerFunc(s.handleMe)))
	s.mux.Handle("POST /machines/heartbeat", s.requireMachineAuth(http.HandlerFunc(s.handleMachineHeartbeat)))
	s.mux.Handle("GET /machines/me", s.requireMachineAuth(http.HandlerFunc(s.handleMachineMe)))
	s.mux.Handle("GET /machines/me/active-user", s.requireMachineAuth(http.HandlerFunc(s.handleGetMachineActiveUser)))
	s.mux.Handle("POST /machines/{machine_id}/active-user", s.requireAuth(http.HandlerFunc(s.handleSetMachineActiveUser)))
	s.mux.Handle("GET /machines", s.requireAuth(http.HandlerFunc(s.handleListMachines)))
	s.mux.Handle("GET /users", s.requireAuth(http.HandlerFunc(s.handleListUsers)))
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

func (s *Server) handleRegisterMachine(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		Fingerprint string `json:"fingerprint"`
	}

	if err := decodeJSON(r, &body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	registered, err := s.machineService.Register(r.Context(), body.Name, body.Fingerprint)
	if err != nil {
		switch {
		case errors.Is(err, machinesvc.ErrInvalidMachineInput):
			respondError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, machinesvc.ErrMachineAlreadyRegistered):
			respondError(w, http.StatusConflict, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to register machine")
		}
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{
		"machine":       sanitizeMachine(registered.Machine),
		"machine_token": registered.MachineToken,
	})
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

func (s *Server) handleMachineHeartbeat(w http.ResponseWriter, r *http.Request) {
	authContext, ok := machineAuthFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "machine authentication required")
		return
	}

	lastSeenAt, err := s.machineService.Touch(r.Context(), authContext.Machine.ID)
	if err != nil {
		switch {
		case errors.Is(err, machinesvc.ErrMachineNotFound):
			respondError(w, http.StatusNotFound, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to update machine heartbeat")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"status":       "ok",
		"machine_id":   authContext.Machine.ID,
		"last_seen_at": lastSeenAt,
	})
}

func (s *Server) handleMachineMe(w http.ResponseWriter, r *http.Request) {
	authContext, ok := machineAuthFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "machine authentication required")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"machine": sanitizeMachine(authContext.Machine)})
}

func (s *Server) handleSetMachineActiveUser(w http.ResponseWriter, r *http.Request) {
	machineID := strings.TrimSpace(r.PathValue("machine_id"))
	if machineID == "" {
		respondError(w, http.StatusBadRequest, "machine_id is required")
		return
	}

	var body struct {
		UserID string `json:"user_id"`
	}

	if err := decodeJSON(r, &body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if _, err := s.machineService.GetByID(r.Context(), machineID); err != nil {
		switch {
		case errors.Is(err, machinesvc.ErrMachineNotFound):
			respondError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, machinesvc.ErrInvalidMachineInput):
			respondError(w, http.StatusBadRequest, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to load machine")
		}
		return
	}

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

	respondJSON(w, http.StatusOK, map[string]string{
		"status":     "active_user_updated",
		"machine_id": machineID,
	})
}

func (s *Server) handleGetMachineActiveUser(w http.ResponseWriter, r *http.Request) {
	authContext, ok := machineAuthFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "machine authentication required")
		return
	}

	active, err := s.userService.GetActiveUserForMachine(r.Context(), authContext.Machine.ID)
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

	respondJSON(w, http.StatusOK, map[string]any{
		"machine_id": authContext.Machine.ID,
		"user":       sanitizeUser(active),
	})
}

func (s *Server) handleListMachines(w http.ResponseWriter, r *http.Request) {
	machines, err := s.machineService.List(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list machines")
		return
	}

	publicMachines := make([]any, 0, len(machines))
	for i := range machines {
		publicMachines = append(publicMachines, sanitizeMachine(&machines[i]))
	}

	respondJSON(w, http.StatusOK, map[string]any{"machines": publicMachines})
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

func (s *Server) requireMachineAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		machineID := strings.TrimSpace(r.Header.Get("X-Machine-ID"))
		machineToken := strings.TrimSpace(r.Header.Get("X-Machine-Token"))
		if machineID == "" || machineToken == "" {
			respondError(w, http.StatusUnauthorized, "missing machine credentials")
			return
		}

		machine, err := s.machineService.Authenticate(r.Context(), machineID, machineToken)
		if err != nil {
			switch {
			case errors.Is(err, machinesvc.ErrMachineInactive):
				respondError(w, http.StatusForbidden, err.Error())
			default:
				respondError(w, http.StatusUnauthorized, "invalid machine credentials")
			}
			return
		}

		next.ServeHTTP(w, r.WithContext(withMachineAuthContext(r.Context(), &MachineAuthContext{
			Machine: machine,
			Token:   machineToken,
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

func sanitizeMachine(machine *domainmachine.Machine) map[string]any {
	if machine == nil {
		return map[string]any{}
	}

	return map[string]any{
		"id":           machine.ID,
		"name":         machine.Name,
		"status":       machine.Status,
		"created_at":   machine.CreatedAt,
		"updated_at":   machine.UpdatedAt,
		"last_seen_at": machine.LastSeenAt,
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

func withMachineAuthContext(ctx context.Context, authContext *MachineAuthContext) context.Context {
	return context.WithValue(ctx, authenticatedMachineContextKey, authContext)
}

func machineAuthFromContext(ctx context.Context) (*MachineAuthContext, bool) {
	value := ctx.Value(authenticatedMachineContextKey)
	authContext, ok := value.(*MachineAuthContext)
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
