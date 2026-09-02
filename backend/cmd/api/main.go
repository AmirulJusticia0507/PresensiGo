package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"

	"github.com/PresensiGo/backend/internal/auth"
	"github.com/PresensiGo/backend/internal/config"
	redisc "github.com/PresensiGo/backend/internal/cache"
	deliveryhttp "github.com/PresensiGo/backend/internal/delivery/http"
	"github.com/PresensiGo/backend/internal/delivery/http/middleware"
	"github.com/PresensiGo/backend/internal/repository"
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
	log.Println("Connected to database")

	redisClient, err := redisc.NewClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Printf("Warning: Redis not available: %v", err)
	} else {
		defer redisClient.Close()
		log.Println("Connected to Redis")
	}

	userRepo := repository.NewUserRepository(db)
	attRepo := repository.NewAttendanceRepository(db)

	authUc := usecase.NewAuthUsecase(userRepo, cfg)
	attUc := usecase.NewAttendanceUsecase(attRepo, cfg)

	httpHandler := deliveryhttp.NewHandler(authUc, attUc)

	middleware.InitJWT(cfg.JWT.Secret, cfg.JWT.ExpireHour)

	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthMiddleware)
	if redisClient != nil {
		api.Use(middleware.RateLimitMiddleware(redisClient, 100, 15*time.Minute))
	}

	httpHandler.RegisterRoutes(api)

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}).Methods("GET")

	r.HandleFunc("/api/auth/register", httpHandler.Register).Methods("POST")
	r.HandleFunc("/api/auth/login", httpHandler.Login).Methods("POST")

	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	bh := c.Handler(r)

	fmt.Printf("Server starting on port %s\n", port)
	if err := http.ListenAndServe(":"+port, bh); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}