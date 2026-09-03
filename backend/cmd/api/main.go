package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"

	"github.com/PresensiGo/backend/internal/config"
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

	userRepo := repository.NewUserRepository(db)
	attRepo := repository.NewAttendanceRepository(db)

	authUc := usecase.NewAuthUsecase(userRepo, cfg)
	attUc := usecase.NewAttendanceUsecase(attRepo, cfg)

	httpHandler := deliveryhttp.NewHandler(authUc, attUc)

	middleware.InitJWT(cfg.JWT.Secret, cfg.JWT.ExpireHour)

	r := mux.NewRouter()

	r.Use(middleware.AuthMiddleware)

	httpHandler.RegisterRoutes(r)

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}).Methods("GET")

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