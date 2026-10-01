package config

// CORSConfig holds CORS configuration for different environments
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	Credentials    bool
}

// LoadCORSConfig returns environment-specific CORS configuration
// env can be "development", "staging", "production"
// Development allows localhost, 127.0.0.1, and Android emulator (10.0.2.2)
// Staging and Production allow explicit frontend domains
func LoadCORSConfig(env string) *CORSConfig {
	switch env {
	case "staging":
		return &CORSConfig{
			AllowedOrigins: []string{
				"https://staging-app.example.com",
			},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type", "Authorization"},
			Credentials:    true,
		}
	case "production":
		return &CORSConfig{
			AllowedOrigins: []string{
				"https://app.example.com",
			},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type", "Authorization"},
			Credentials:    true,
		}
	case "development":
		fallthrough
	default:
		// Development environment (default)
		return &CORSConfig{
			AllowedOrigins: []string{
				"http://localhost:3000",
				"http://localhost:8080",
				"http://localhost:5000",
				"http://127.0.0.1:3000",
				"http://127.0.0.1:8080",
				"http://127.0.0.1:5000",
				"http://10.0.2.2:3000",   // Android emulator
				"http://10.0.2.2:8080",   // Android emulator
				"http://10.0.2.2:5000",   // Android emulator
			},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"*"},
			Credentials:    true,
		}
	}
}
