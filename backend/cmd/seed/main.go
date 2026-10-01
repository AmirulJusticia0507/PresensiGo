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

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Hash passwords
	adminHash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	employeeHash, _ := bcrypt.GenerateFromPassword([]byte("employee123"), bcrypt.DefaultCost)

	// Clear existing data (only if needed - be careful in production)
	db.Exec("DELETE FROM attendances")
	db.Exec("DELETE FROM users")

	// Admin user
	adminID := uuid.New()
	_, err = db.Exec(`
		INSERT INTO users (id, name, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (email) DO NOTHING
	`, adminID, "Admin", "admin@presensigo.local", string(adminHash), "admin")
	if err != nil {
		log.Println("Error inserting admin user:", err)
	}

	// Employee 1
	emp1ID := uuid.New()
	_, err = db.Exec(`
		INSERT INTO users (id, name, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (email) DO NOTHING
	`, emp1ID, "Employee One", "employee1@presensigo.local", string(employeeHash), "employee")
	if err != nil {
		log.Println("Error inserting employee 1:", err)
	}

	// Employee 2
	emp2ID := uuid.New()
	_, err = db.Exec(`
		INSERT INTO users (id, name, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (email) DO NOTHING
	`, emp2ID, "Employee Two", "employee2@presensigo.local", string(employeeHash), "employee")
	if err != nil {
		log.Println("Error inserting employee 2:", err)
	}

	fmt.Println("Seed completed successfully!")
	fmt.Println("")
	fmt.Println("Admin user:")
	fmt.Println("  Email: admin@presensigo.local")
	fmt.Println("  Password: admin123")
	fmt.Println("")
	fmt.Println("Employee users:")
	fmt.Println("  Email: employee1@presensigo.local")
	fmt.Println("  Email: employee2@presensigo.local")
	fmt.Println("  Password: employee123")
}
