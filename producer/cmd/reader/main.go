package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"

	"acc-dp/producer/internal/repository/postgres"
	"acc-dp/producer/internal/service/avro"
	"acc-dp/producer/internal/service/normalizer"
	usersvc "acc-dp/producer/internal/service/user"
	"acc-dp/producer/internal/source/acc_shm"
)

func main() {
	loadDotEnvIfPresent()

	interval := flag.Duration("interval", 500*time.Millisecond, "read interval (example: 250ms, 1s)")
	once := flag.Bool("once", false, "read a single sample and exit")
	userID := flag.String("user-id", "", "user id used in event envelope")
	username := flag.String("username", "", "username used in event envelope")
	databaseURL := flag.String("database-url", envOrDefault("DATABASE_URL", ""), "postgres connection string used to read active user")
	machineIDFlag := flag.String("machine-id", envOrDefault("ACCDP_MACHINE_ID", ""), "unique machine id used for active user resolution")
	machineIDPath := flag.String("machine-id-path", envOrDefault("ACCDP_MACHINE_ID_PATH", "./data/machine_id"), "path to local machine id file")
	schemaVersion := flag.Int("schema-version", 1, "event schema version")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	reader, err := acc_shm.NewReader(ctx)
	if err != nil {
		log.Fatalf("failed to create shared memory reader: %v", err)
	}

	normalizerService := normalizer.New()
	serializer, err := avro.NewSerializer()
	if err != nil {
		log.Fatalf("failed to create avro serializer: %v", err)
	}

	machineID, machineIDSource, err := resolveMachineID(*machineIDFlag, *machineIDPath)
	if err != nil {
		log.Fatalf("failed to resolve machine id: %v", err)
	}

	identity, identitySource, closeIdentityStore, err := resolveIdentity(ctx, *userID, *username, *databaseURL, machineID)
	if err != nil {
		log.Fatalf("failed to resolve identity: %v", err)
	}
	defer closeIdentityStore()

	log.Printf(
		"using machine_id=%s machine_id_source=%s identity_source=%s usuario_id=%s username=%s",
		machineID,
		machineIDSource,
		identitySource,
		identity.UsuarioID,
		identity.Username,
	)

	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("failed to close shared memory reader: %v", closeErr)
		}
	}()

	printStaticInfo(ctx, reader, normalizerService, serializer, identity, int32(*schemaVersion))

	if *once {
		if err := printSample(ctx, reader, normalizerService, serializer, identity, int32(*schemaVersion)); err != nil {
			log.Fatalf("failed to read sample: %v", err)
		}
		return
	}

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("shutting down reader: %v", ctx.Err())
			return
		case <-ticker.C:
			if err := printSample(ctx, reader, normalizerService, serializer, identity, int32(*schemaVersion)); err != nil {
				if errors.Is(err, acc_shm.ErrReaderClosed) {
					log.Printf("reader closed: %v", err)
					return
				}
				log.Printf("read error: %v", err)
			}
		}
	}
}

func resolveIdentity(
	ctx context.Context,
	userIDFlag string,
	usernameFlag string,
	databaseURL string,
	machineID string,
) (normalizer.Identity, string, func(), error) {
	cleanup := func() {}

	userID := strings.TrimSpace(userIDFlag)
	username := strings.TrimSpace(usernameFlag)

	if (userID == "") != (username == "") {
		return normalizer.Identity{}, "", cleanup, fmt.Errorf("user-id and username must be provided together")
	}

	if userID != "" && username != "" {
		return normalizer.Identity{UsuarioID: userID, Username: username}, "flags", cleanup, nil
	}

	if strings.TrimSpace(databaseURL) == "" {
		return normalizer.Identity{UsuarioID: "unknown", Username: "unknown"}, "fallback", cleanup, nil
	}

	repo, err := postgres.New(databaseURL)
	if err != nil {
		return normalizer.Identity{}, "", cleanup, fmt.Errorf("create postgres repository: %w", err)
	}

	cleanup = func() {
		if closeErr := repo.Close(); closeErr != nil {
			log.Printf("failed to close identity repository: %v", closeErr)
		}
	}

	if err := repo.Ping(ctx); err != nil {
		cleanup()
		return normalizer.Identity{}, "", func() {}, fmt.Errorf("ping postgres: %w", err)
	}

	activeUser, err := usersvc.New(repo).GetActiveUserForMachine(ctx, machineID)
	if err != nil {
		cleanup()
		if errors.Is(err, usersvc.ErrActiveUserNotSet) {
			return normalizer.Identity{}, "", func() {}, fmt.Errorf("no active user configured for machine_id=%s; set one via backend API or pass --user-id and --username", machineID)
		}
		return normalizer.Identity{}, "", func() {}, fmt.Errorf("load active user for machine: %w", err)
	}

	return normalizer.Identity{
		UsuarioID: activeUser.ID,
		Username:  activeUser.Username,
	}, "postgres", cleanup, nil
}

func resolveMachineID(machineIDOverride string, machineIDPath string) (string, string, error) {
	machineIDOverride = strings.TrimSpace(machineIDOverride)
	if machineIDOverride != "" {
		return machineIDOverride, "override", nil
	}

	machineIDPath = strings.TrimSpace(machineIDPath)
	if machineIDPath == "" {
		machineIDPath = "./data/machine_id"
	}

	storedMachineID, err := readMachineID(machineIDPath)
	if err != nil {
		return "", "", fmt.Errorf("read machine id file: %w", err)
	}
	if storedMachineID != "" {
		return storedMachineID, "file", nil
	}

	generatedMachineID := uuid.NewString()
	if err := writeMachineID(machineIDPath, generatedMachineID); err != nil {
		return "", "", fmt.Errorf("write machine id file: %w", err)
	}

	return generatedMachineID, "generated", nil
}

func readMachineID(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	return strings.TrimSpace(string(content)), nil
}

func writeMachineID(path string, machineID string) error {
	directory := filepath.Dir(path)
	if directory != "." && directory != "" {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			return err
		}
	}

	return os.WriteFile(path, []byte(machineID+"\n"), 0o600)
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

func envOrDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func printStaticInfo(
	ctx context.Context,
	reader *acc_shm.SharedMemoryReader,
	normalizerService *normalizer.Service,
	serializer *avro.Serializer,
	identity normalizer.Identity,
	schemaVersion int32,
) {
	page, err := reader.ReadStatic(ctx)
	if err != nil {
		log.Printf("static page not available: %v", err)
		return
	}

	staticEvent, err := normalizerService.NormalizeStatic(page, identity, schemaVersion)
	if err != nil {
		log.Printf("failed to normalize static page: %v", err)
		return
	}

	encoded, err := serializer.SerializeStatic(staticEvent)
	if err != nil {
		log.Printf("failed to serialize static page: %v", err)
		return
	}

	fmt.Printf("static: sessions=%d cars=%d max_rpm=%d max_fuel=%.2f avro_bytes=%d\n",
		page.NumberOfSessions,
		page.NumCars,
		page.MaxRPM,
		page.MaxFuel,
		len(encoded),
	)
}

func printSample(
	ctx context.Context,
	reader *acc_shm.SharedMemoryReader,
	normalizerService *normalizer.Service,
	serializer *avro.Serializer,
	identity normalizer.Identity,
	schemaVersion int32,
) error {
	physics, err := reader.ReadPhysics(ctx)
	if err != nil {
		return fmt.Errorf("read physics: %w", err)
	}

	graphics, err := reader.ReadGraphics(ctx)
	if err != nil {
		return fmt.Errorf("read graphics: %w", err)
	}

	physicsEvent, err := normalizerService.NormalizePhysics(physics, identity, schemaVersion)
	if err != nil {
		return fmt.Errorf("normalize physics: %w", err)
	}

	graphicsEvent, err := normalizerService.NormalizeGraphics(graphics, identity, schemaVersion)
	if err != nil {
		return fmt.Errorf("normalize graphics: %w", err)
	}

	physicsEncoded, err := serializer.SerializePhysics(physicsEvent)
	if err != nil {
		return fmt.Errorf("serialize physics: %w", err)
	}

	graphicsEncoded, err := serializer.SerializeGraphics(graphicsEvent)
	if err != nil {
		return fmt.Errorf("serialize graphics: %w", err)
	}

	fmt.Printf(
		"physics(packet=%d speed=%.2f rpm=%d gear=%d avro=%dB) graphics(packet=%d status=%d session=%d active_cars=%d avro=%dB)\n",
		physics.PacketID,
		physics.SpeedKmh,
		physics.RPM,
		physics.Gear,
		len(physicsEncoded),
		graphics.PacketID,
		graphics.Status,
		graphics.Session,
		graphics.ActiveCars,
		len(graphicsEncoded),
	)

	return nil
}
