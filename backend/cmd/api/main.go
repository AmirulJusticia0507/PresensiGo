package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"

	"github.com/PresensiGo/backend/internal/ai"
	"github.com/PresensiGo/backend/internal/config"
	deliveryhttp "github.com/PresensiGo/backend/internal/delivery/http"
	"github.com/PresensiGo/backend/internal/delivery/http/middleware"
	"github.com/PresensiGo/backend/internal/infrastructure"
	"github.com/PresensiGo/backend/internal/repository"
	storage "github.com/PresensiGo/backend/internal/storage"
	"github.com/PresensiGo/backend/internal/usecase"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DB.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✓ Connected to database")

	// Initialize Redis client
	redisClient, err := infrastructure.NewRedisClient(cfg.Redis.Addr)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	userRepo := repository.NewUserRepository(db)
	attRepo := repository.NewAttendanceRepository(db)
	offlineRepo := repository.NewOfflineQueueRepository(db)
	fraudRepo := repository.NewFraudAttemptRepository(db)

	minioCtx, cancelMinio := context.WithTimeout(context.Background(), 10*time.Second)
	minioClient, err := storage.NewClient(
		minioCtx,
		cfg.MinIO.Endpoint,
		cfg.MinIO.AccessKey,
		cfg.MinIO.SecretKey,
		cfg.MinIO.Bucket,
		cfg.MinIO.UseSSL,
	)
	cancelMinio()
	if err != nil {
		log.Fatalf("Failed to initialize MinIO: %v", err)
	}
	log.Printf("Connected to MinIO bucket %q", cfg.MinIO.Bucket)

	faceAI := ai.NewClient(
		cfg.AI.URL,
		time.Duration(cfg.AI.TimeoutSeconds)*time.Second,
	)
	authUc := usecase.NewAuthUsecase(userRepo, cfg, faceAI)
	attUc := usecase.NewAttendanceUsecase(attRepo, userRepo, offlineRepo, cfg, minioClient, faceAI)
	profileUc := usecase.NewProfileUsecase(userRepo)

	httpHandler := deliveryhttp.NewHandler(authUc, attUc, profileUc, db, redisClient, fraudRepo)

	middleware.InitJWT(cfg.JWT.Secret, cfg.JWT.ExpireHour)

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(redisClient)

	r := mux.NewRouter()

	// Apply global middleware stack
	r.Use(middleware.RequestIDMiddleware)

	// Health check endpoints (public, no auth, no rate limit)
	r.HandleFunc("/health", httpHandler.Health).Methods("GET")
	r.HandleFunc("/health/ready", httpHandler.HealthReady).Methods("GET")

	// Auth routes with rate limiting (public endpoints, before auth middleware)
	loginRouter := r.NewRoute().Subrouter()
	loginRouter.Use(rateLimiter.RateLimitMiddleware("login"))
	loginRouter.HandleFunc("/api/auth/login", httpHandler.Login).Methods("POST")

	registerRouter := r.NewRoute().Subrouter()
	registerRouter.Use(rateLimiter.RateLimitMiddleware("register"))
	registerRouter.HandleFunc("/api/auth/register", httpHandler.Register).Methods("POST")

	// Protected routes (auth + rate limit)
	protectedRouter := r.NewRoute().Subrouter()
	protectedRouter.Use(middleware.AuthMiddleware)

	// Attendance check-in/out with specific limits
	checkInRouter := protectedRouter.NewRoute().Subrouter()
	checkInRouter.Use(rateLimiter.RateLimitMiddleware("check_in"))
	checkInRouter.HandleFunc("/api/attendance/check-in", httpHandler.CheckIn).Methods("POST")

	checkOutRouter := protectedRouter.NewRoute().Subrouter()
	checkOutRouter.Use(rateLimiter.RateLimitMiddleware("check_out"))
	checkOutRouter.HandleFunc("/api/attendance/check-out", httpHandler.CheckOut).Methods("POST")

	// Other protected routes with default rate limit
	defaultLimitRouter := protectedRouter.NewRoute().Subrouter()
	defaultLimitRouter.Use(rateLimiter.RateLimitMiddleware("default"))
	defaultLimitRouter.HandleFunc("/api/attendance/today", httpHandler.GetTodayAttendance).Methods("GET")
	defaultLimitRouter.HandleFunc("/api/attendance/history", httpHandler.GetHistory).Methods("GET")
	defaultLimitRouter.HandleFunc("/api/attendance/sync", httpHandler.SyncAttendance).Methods("POST")
	defaultLimitRouter.HandleFunc("/api/attendance/sync/status", httpHandler.GetSyncStatus).Methods("GET")
	defaultLimitRouter.HandleFunc("/api/locations", httpHandler.GetLocations).Methods("GET")
	defaultLimitRouter.HandleFunc("/api/locations", httpHandler.CreateLocation).Methods("POST")
	defaultLimitRouter.HandleFunc("/api/locations/{id}", httpHandler.UpdateLocation).Methods("PUT")
	defaultLimitRouter.HandleFunc("/api/locations/{id}", httpHandler.DeleteLocation).Methods("DELETE")
	defaultLimitRouter.HandleFunc("/api/profile", httpHandler.GetProfile).Methods("GET")
	defaultLimitRouter.HandleFunc("/api/profile", httpHandler.UpdateProfile).Methods("PUT")
	defaultLimitRouter.HandleFunc("/api/profile/face-enrollment", httpHandler.EnrollFace).Methods("POST")
	defaultLimitRouter.HandleFunc("/api/face/challenge", httpHandler.GetFaceChallenge).Methods("POST")
	defaultLimitRouter.HandleFunc("/api/security/location-attempts", httpHandler.ReportFraudAttempt).Methods("POST")
	defaultLimitRouter.HandleFunc("/api/admin/security/location-alerts", httpHandler.GetFraudAlerts).Methods("GET")

	// Password change with stricter rate limit
	passwordLimitRouter := protectedRouter.NewRoute().Subrouter()
	passwordLimitRouter.Use(rateLimiter.RateLimitMiddleware("password_change"))
	passwordLimitRouter.HandleFunc("/api/profile/password", httpHandler.ChangePassword).Methods("PUT")

	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	// Load environment-specific CORS configuration
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}
	corsConfig := config.LoadCORSConfig(environment)

	c := cors.New(cors.Options{
		AllowedOrigins:   corsConfig.AllowedOrigins,
		AllowedMethods:   corsConfig.AllowedMethods,
		AllowedHeaders:   corsConfig.AllowedHeaders,
		AllowCredentials: corsConfig.Credentials,
	})

	bh := c.Handler(r)

	// Start HTTP server in a goroutine
	server := &http.Server{
		Addr:    ":" + port,
		Handler: bh,
	}

	go func() {
		log.Printf("Server starting on port %s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("\nShutdown signal received, gracefully stopping server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
