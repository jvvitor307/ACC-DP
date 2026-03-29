package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"acc-dp/producer/internal/handler/api"
	"acc-dp/producer/internal/repository/postgres"
	"acc-dp/producer/internal/service/auth"
	usersvc "acc-dp/producer/internal/service/user"
)

func main() {
	loadDotEnvIfPresent()

	addr := flag.String("addr", envOrDefault("BACKEND_ADDR", ":8088"), "http listen address")
	databaseURL := flag.String("database-url", envOrDefault("DATABASE_URL", ""), "postgres connection string")
	sessionTTL := flag.Duration("session-ttl", parseDurationEnv("AUTH_SESSION_TTL", 24*time.Hour), "session ttl")
	cookieName := flag.String("cookie-name", envOrDefault("AUTH_COOKIE_NAME", "accdp_session"), "auth cookie name")
	cookieSecure := flag.Bool("cookie-secure", parseBoolEnv("AUTH_COOKIE_SECURE", false), "set Secure on auth cookie")
	cookieDomain := flag.String("cookie-domain", envOrDefault("AUTH_COOKIE_DOMAIN", ""), "auth cookie domain")
	flag.Parse()

	if strings.TrimSpace(*databaseURL) == "" {
		log.Fatal("database-url is required (flag or DATABASE_URL)")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	repo, err := postgres.New(*databaseURL)
	if err != nil {
		log.Fatalf("failed to create postgres repository: %v", err)
	}
	defer func() {
		if closeErr := repo.Close(); closeErr != nil {
			log.Printf("failed to close postgres repository: %v", closeErr)
		}
	}()

	if err := repo.Ping(ctx); err != nil {
		log.Fatalf("failed to ping postgres: %v", err)
	}

	if err := repo.RunMigrations(ctx); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	authService := auth.New(repo, *sessionTTL)
	userService := usersvc.New(repo)

	handler := api.NewServer(authService, userService, api.Config{
		CookieName:   *cookieName,
		CookieSecure: *cookieSecure,
		CookieDomain: *cookieDomain,
		SessionTTL:   *sessionTTL,
	})

	server := &http.Server{
		Addr:              *addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("http shutdown error: %v", err)
		}
	}()

	log.Printf("backend listening on %s", *addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("backend failed: %v", err)
	}
}

func loadDotEnvIfPresent() {
	paths := []string{".env", "../.env", "../../.env"}

	for _, path := range paths {
		loaded, err := loadDotEnvFile(path)
		if err != nil {
			log.Printf("failed to load env file %s: %v", path, err)
			continue
		}

		if loaded {
			log.Printf("loaded env file: %s", path)
			return
		}
	}
}

func loadDotEnvFile(path string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	for i := range lines {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return false, fmt.Errorf("invalid env format at line %d", i+1)
		}

		key = strings.TrimSpace(key)
		if key == "" {
			return false, fmt.Errorf("empty env key at line %d", i+1)
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		value = strings.TrimSpace(value)
		value = trimWrappingQuotes(value)

		if err := os.Setenv(key, value); err != nil {
			return false, fmt.Errorf("set env %s: %w", key, err)
		}
	}

	return true, nil
}

func trimWrappingQuotes(value string) string {
	if len(value) < 2 {
		return value
	}

	if value[0] == '"' && value[len(value)-1] == '"' {
		return value[1 : len(value)-1]
	}

	if value[0] == '\'' && value[len(value)-1] == '\'' {
		return value[1 : len(value)-1]
	}

	return value
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func parseDurationEnv(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		log.Printf("invalid duration in %s (%s): %v, using fallback %s", key, value, err, fallback)
		return fallback
	}

	return parsed
}

func parseBoolEnv(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		log.Printf("invalid bool in %s (%s): %v, using fallback %t", key, value, err, fallback)
		return fallback
	}

	return parsed
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetPrefix("backend ")
	log.SetOutput(os.Stdout)
	log.SetPrefix(fmt.Sprintf("backend[%d] ", os.Getpid()))
}
