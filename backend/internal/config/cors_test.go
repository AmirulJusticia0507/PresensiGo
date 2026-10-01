package config

import (
	"testing"
)

func TestLoadCORSConfig_Development(t *testing.T) {
	config := LoadCORSConfig("development")

	// Check that development config is loaded
	if config == nil {
		t.Fatal("Expected config to be non-nil")
	}

	// Verify credentials are enabled
	if !config.Credentials {
		t.Error("Expected Credentials to be true")
	}

	// Verify localhost is allowed
	hasLocalhost := false
	for _, origin := range config.AllowedOrigins {
		if origin == "http://localhost:3000" {
			hasLocalhost = true
			break
		}
	}
	if !hasLocalhost {
		t.Error("Expected http://localhost:3000 to be in allowed origins")
	}

	// Verify 127.0.0.1 is allowed
	hasLoopback := false
	for _, origin := range config.AllowedOrigins {
		if origin == "http://127.0.0.1:3000" {
			hasLoopback = true
			break
		}
	}
	if !hasLoopback {
		t.Error("Expected http://127.0.0.1:3000 to be in allowed origins")
	}

	// Verify Android emulator IP is allowed
	hasAndroidEmulator := false
	for _, origin := range config.AllowedOrigins {
		if origin == "http://10.0.2.2:3000" {
			hasAndroidEmulator = true
			break
		}
	}
	if !hasAndroidEmulator {
		t.Error("Expected http://10.0.2.2:3000 (Android emulator) to be in allowed origins")
	}

	// Verify no wildcard origin when credentials are true
	for _, origin := range config.AllowedOrigins {
		if origin == "*" {
			t.Error("Expected no wildcard origin when credentials are true")
		}
	}
}

func TestLoadCORSConfig_Staging(t *testing.T) {
	config := LoadCORSConfig("staging")

	// Check that staging config is loaded
	if config == nil {
		t.Fatal("Expected config to be non-nil")
	}

	// Verify credentials are enabled
	if !config.Credentials {
		t.Error("Expected Credentials to be true")
	}

	// Verify staging domain is allowed
	hasStagingDomain := false
	for _, origin := range config.AllowedOrigins {
		if origin == "https://staging-app.example.com" {
			hasStagingDomain = true
			break
		}
	}
	if !hasStagingDomain {
		t.Error("Expected https://staging-app.example.com to be in allowed origins")
	}

	// Verify only one domain is in staging
	if len(config.AllowedOrigins) != 1 {
		t.Errorf("Expected exactly 1 allowed origin for staging, got %d", len(config.AllowedOrigins))
	}

	// Verify no wildcard origin
	for _, origin := range config.AllowedOrigins {
		if origin == "*" {
			t.Error("Expected no wildcard origin in staging")
		}
	}
}

func TestLoadCORSConfig_Production(t *testing.T) {
	config := LoadCORSConfig("production")

	// Check that production config is loaded
	if config == nil {
		t.Fatal("Expected config to be non-nil")
	}

	// Verify credentials are enabled
	if !config.Credentials {
		t.Error("Expected Credentials to be true")
	}

	// Verify production domain is allowed
	hasProductionDomain := false
	for _, origin := range config.AllowedOrigins {
		if origin == "https://app.example.com" {
			hasProductionDomain = true
			break
		}
	}
	if !hasProductionDomain {
		t.Error("Expected https://app.example.com to be in allowed origins")
	}

	// Verify only one domain is in production
	if len(config.AllowedOrigins) != 1 {
		t.Errorf("Expected exactly 1 allowed origin for production, got %d", len(config.AllowedOrigins))
	}

	// Verify no wildcard origin
	for _, origin := range config.AllowedOrigins {
		if origin == "*" {
			t.Error("Expected no wildcard origin in production")
		}
	}
}

func TestLoadCORSConfig_DefaultsToDevelopment(t *testing.T) {
	// Test with empty string
	config := LoadCORSConfig("")
	if config == nil {
		t.Fatal("Expected config to be non-nil")
	}

	// Should have multiple origins like development
	if len(config.AllowedOrigins) < 5 {
		t.Errorf("Expected multiple origins for default (development) config, got %d", len(config.AllowedOrigins))
	}

	// Test with unknown environment
	config = LoadCORSConfig("unknown")
	if config == nil {
		t.Fatal("Expected config to be non-nil")
	}

	// Should have multiple origins like development
	if len(config.AllowedOrigins) < 5 {
		t.Errorf("Expected multiple origins for unknown env to default to development, got %d", len(config.AllowedOrigins))
	}
}

func TestLoadCORSConfig_MethodsAndHeaders(t *testing.T) {
	// Test that all environments have proper methods and headers
	environments := []string{"development", "staging", "production"}

	for _, env := range environments {
		config := LoadCORSConfig(env)

		// Verify methods include common HTTP verbs
		methods := make(map[string]bool)
		for _, method := range config.AllowedMethods {
			methods[method] = true
		}

		requiredMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		for _, method := range requiredMethods {
			if !methods[method] {
				t.Errorf("Environment %s missing method %s", env, method)
			}
		}

		// Verify headers include Content-Type and Authorization
		headers := make(map[string]bool)
		for _, header := range config.AllowedHeaders {
			headers[header] = true
		}

		if !headers["Content-Type"] && !headers["*"] {
			t.Errorf("Environment %s missing Content-Type in headers", env)
		}
		if !headers["Authorization"] && !headers["*"] {
			t.Errorf("Environment %s missing Authorization in headers", env)
		}
	}
}

func TestLoadCORSConfig_NoWildcardWithCredentials(t *testing.T) {
	// Test that when credentials are true, no wildcard origin is used
	environments := []string{"development", "staging", "production"}

	for _, env := range environments {
		config := LoadCORSConfig(env)

		if config.Credentials {
			for _, origin := range config.AllowedOrigins {
				if origin == "*" {
					t.Errorf("Environment %s has wildcard origin with credentials enabled (security issue)", env)
				}
			}
		}
	}
}
