package middleware

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/PresensiGo/backend/internal/model"
	"github.com/PresensiGo/backend/internal/delivery/http"
)

var validate *validator.Validate

func InitValidation() {
	validate = validator.New()
}

func ValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		var body interface{}

		contentType := r.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/json") {
			if err := http.Json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.respondError(w, http.StatusBadRequest, "invalid request body")
				return
			}
		} else {
			next.ServeHTTP(w, r)
			return
		}

		if body != nil {
			if err := validate.Struct(body); err != nil {
				http.respondError(w, http.StatusBadRequest, err.Error())
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}