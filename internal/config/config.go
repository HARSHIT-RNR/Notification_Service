package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DBDriver     string
	DBSource     string
	KafkaBrokers []string
	KafkaTopics  []string
	KafkaGroupID string
	SmtpHost     string
	SmtpPort     int
	SmtpUser     string
	SmtpPass     string
	SmtpFrom     string
	GrpcPort     string
	TemplatePath string
}

func LoadConfig() *Config {
	// --- START: Robust .env loading ---
	// This code finds the project root directory by looking for the go.mod file
	// and loads the .env file from there. This makes the app runnable from any subdirectory.
	_, b, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(b), "..", "..") // Go up two directories from /internal/config

	envPath := filepath.Join(projectRoot, ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("Warning: Could not load .env file from %s. Using environment variables. Error: %v", envPath, err)
	} else {
		log.Printf("Successfully loaded .env file from %s", envPath)
	}
	// --- END: Robust .env loading ---

	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	// Construct the database source string from individual env vars
	dbSource := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "2120"),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_NAME", "notification_mails"),
		getEnv("DB_SSL_MODE", "disable"),
	)

	return &Config{
		DBDriver:     getEnv("DB_DRIVER", "postgres"),
		DBSource:     dbSource,
		KafkaBrokers: strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
		KafkaTopics: []string{
			"notification.send-welcome",
			"notification.send-signup-verification",
			"notification.provisioning-started",
			"notification.send-password-setup",
		},
		KafkaGroupID: getEnv("KAFKA_GROUP_ID", "notification-group"),
		SmtpHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SmtpPort:     smtpPort,
		SmtpUser:     getEnv("SMTP_USER", ""),
		SmtpPass:     getEnv("SMTP_PASS", ""),
		SmtpFrom:     getEnv("SMTP_FROM_EMAIL", "your-email@example.com"),
		GrpcPort:     getEnv("GRPC_PORT", "50051"),
		TemplatePath: getEnv("TEMPLATE_PATH", "internal/adapters/templates/emails"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
