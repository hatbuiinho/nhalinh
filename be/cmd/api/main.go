package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"nhalinh/be/internal/config"
	"nhalinh/be/internal/db"
	"nhalinh/be/internal/device"
	"nhalinh/be/internal/httpapi"
	"nhalinh/be/internal/memorial"
	"nhalinh/be/internal/migration"
	"nhalinh/be/internal/ota"
	"nhalinh/be/internal/storage"
	"nhalinh/be/internal/user"
)

func main() {
	config.LoadDotEnv()

	addr := env("ADDR", ":8080")
	ctx := context.Background()

	stores := newStores(ctx)
	if stores.pool != nil {
		defer stores.pool.Close()
	}

	deviceService := device.NewService(stores.devices, time.Now)
	userService := user.NewService(stores.users, time.Now)
	memorialService := memorial.NewService(stores.memorials, time.Now)
	if err := userService.EnsureInitialAdmin(ctx, user.CreateInput{
		Username:    os.Getenv("INITIAL_ADMIN_USERNAME"),
		DisplayName: env("INITIAL_ADMIN_DISPLAY_NAME", "Ban quản trị"),
		Password:    os.Getenv("INITIAL_ADMIN_PASSWORD"),
		Role:        user.RoleAdmin,
	}); err != nil {
		log.Fatalf("create initial admin: %v", err)
	}
	otaStorageDir := env("OTA_STORAGE_DIR", "storage/ota")
	otaService := ota.NewService(otaStorageDir)
	objectStorage := newObjectStorage()
	router := httpapi.NewRouter(deviceService, userService, memorialService, otaService, otaStorageDir, objectStorage)

	log.Printf("api listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}

func newObjectStorage() *storage.MinIO {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		log.Printf("MINIO_ENDPOINT is not set; avatar uploads are disabled")
		return nil
	}
	useSSL, _ := strconv.ParseBool(env("MINIO_USE_SSL", "false"))
	client, err := storage.NewMinIO(storage.Config{
		Endpoint: endpoint, AccessKey: os.Getenv("MINIO_ACCESS_KEY"), SecretKey: os.Getenv("MINIO_SECRET_KEY"),
		Bucket: os.Getenv("MINIO_BUCKET"), Region: env("MINIO_REGION", "us-east-1"), UseSSL: useSSL,
		PublicBase: os.Getenv("MINIO_PUBLIC_BASE_URL"),
	})
	if err != nil {
		log.Fatalf("configure object storage: %v", err)
	}
	return client
}

func env(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

type stores struct {
	pool      *pgxpool.Pool
	devices   device.Store
	users     user.Store
	memorials memorial.Store
}

func newStores(ctx context.Context) stores {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Printf("DATABASE_URL is not set; using in-memory stores")
		return stores{
			devices:   device.NewMemoryStore(),
			users:     user.NewMemoryStore(),
			memorials: memorial.NewMemoryStore(),
		}
	}

	pool, err := db.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	if err := migration.Up(ctx, pool); err != nil {
		pool.Close()
		log.Fatalf("migrate postgres: %v", err)
	}

	log.Printf("using postgres stores (schema up to date)")
	return stores{
		pool:      pool,
		devices:   device.NewPostgresStore(pool),
		users:     user.NewPostgresStore(pool),
		memorials: memorial.NewPostgresStore(pool),
	}
}
