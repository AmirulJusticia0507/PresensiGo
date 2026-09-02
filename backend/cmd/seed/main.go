package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	db, err := sql.Open("postgres", "host=localhost port=5434 user=presensigo password=presensigo123 dbname=presensigo sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	// Clear existing users
	db.Exec("DELETE FROM attendances")
	db.Exec("DELETE FROM users")

	// Admin
	adminID := uuid.New()
	db.Exec(`INSERT INTO users (id, name, email, password_hash, role) VALUES ($1, $2, $3, $4, $5)`,
		adminID, "Admin", "admin@presensigo.com", string(hash), "admin")

	// Employee
	empID := uuid.New()
	db.Exec(`INSERT INTO users (id, name, email, password_hash, role) VALUES ($1, $2, $3, $4, $5)`,
		empID, "Employee", "employee@presensigo.com", string(hash), "employee")

	fmt.Println("Users created successfully!")
	fmt.Println("Email: admin@presensigo.com / employee@presensigo.com")
	fmt.Println("Password: password")
}
