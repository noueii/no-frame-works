package provider

import (
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-errors/errors"
	"github.com/joho/godotenv"
)

type EnvProvider struct {
	appEnv               string
	appSecret            string
	serverPort           string
	databaseURL          string
	databaseMaxConns     int
	redisHost            string
	redisPort            string
	redisPassword        string
	redisDB              int
	logLevel             string
	sentryDsn            string
	sentryEnv            string
	appBaseURL           string
	appLoginRedirectURL  string
	appLogoutRedirectURL string
	kratosPublicURL      string
	kratosAdminURL       string
}

func NewEnvProvider(rootDir string) (*EnvProvider, error) {
	fallbackLookupEnv := func(key string, fallback string) string {
		value, exists := os.LookupEnv(key)
		if !exists {
			return fallback
		}
		return value
	}

	var firstErr error
	requireLookupEnv := func(key string) string {
		if firstErr != nil {
			return ""
		}
		value, exists := os.LookupEnv(key)
		if !exists {
			firstErr = errors.Errorf("env key: %q is not set", key)
			return ""
		}
		return value
	}

	appServer := fallbackLookupEnv("APP_ENV", "local")

	if appServer == "local" {
		envPath := filepath.Join(rootDir, ".env")
		if err := godotenv.Load(envPath); err != nil {
			if !os.IsNotExist(err) {
				slog.Default().
					Warn("failed to load .env file", slog.String("path", envPath), slog.String("error", err.Error()))
			}
		}
		appServer = fallbackLookupEnv("APP_ENV", "local")
	}

	appSecret := requireLookupEnv("APP_SECRET")
	serverPort := fallbackLookupEnv("SERVER_PORT", "3000")

	databaseURL := requireLookupEnv("DATABASE_URL")
	redisHost := requireLookupEnv("REDIS_HOST")
	redisPort := requireLookupEnv("REDIS_PORT")
	redisPassword := fallbackLookupEnv("REDIS_PASSWORD", "")
	redisDBString := fallbackLookupEnv("REDIS_DB", "0")
	logLevel := requireLookupEnv("LOG_LEVEL")
	sentryDsn := fallbackLookupEnv("SENTRY_DSN", "")
	sentryEnv := fallbackLookupEnv("SENTRY_ENV", "local")
	appBaseURL := requireLookupEnv("APP_BASE_URL")
	appLoginRedirectURL := requireLookupEnv("APP_LOGIN_REDIRECT_URL")
	appLogoutRedirectURL := requireLookupEnv("APP_LOGOUT_REDIRECT_URL")
	kratosPublicURL := requireLookupEnv("KRATOS_PUBLIC_URL")
	kratosAdminURL := fallbackLookupEnv("KRATOS_ADMIN_URL", "")

	if firstErr != nil {
		return nil, firstErr
	}

	databaseMaxConnsString := fallbackLookupEnv("DATABASE_MAX_CONNS", "5")
	parsedDatabaseMaxConns, err := strconv.Atoi(databaseMaxConnsString)
	if err != nil {
		return nil, errors.Errorf(
			"failed to parse DATABASE_MAX_CONNS val %s: %w",
			databaseMaxConnsString,
			err,
		)
	}
	parsedRedisDB, err := strconv.Atoi(redisDBString)
	if err != nil {
		return nil, errors.Errorf("failed to parse REDIS_DB val %s: %w", redisDBString, err)
	}

	envProvider := EnvProvider{
		appEnv:               appServer,
		appSecret:            appSecret,
		serverPort:           serverPort,
		databaseMaxConns:     parsedDatabaseMaxConns,
		databaseURL:          databaseURL,
		redisHost:            redisHost,
		redisPort:            redisPort,
		redisPassword:        redisPassword,
		redisDB:              parsedRedisDB,
		logLevel:             logLevel,
		sentryDsn:            sentryDsn,
		sentryEnv:            sentryEnv,
		appBaseURL:           appBaseURL,
		appLoginRedirectURL:  appLoginRedirectURL,
		appLogoutRedirectURL: appLogoutRedirectURL,
		kratosPublicURL:      kratosPublicURL,
		kratosAdminURL:       kratosAdminURL,
	}

	return &envProvider, nil
}

func (e *EnvProvider) AppEnv() string {
	return e.appEnv
}

func (e *EnvProvider) ServerPort() string {
	return e.serverPort
}

func (e *EnvProvider) AppBaseURL() string {
	return e.appBaseURL
}

func (e *EnvProvider) AppLoginRedirectURL() string {
	return e.appLoginRedirectURL
}

func (e *EnvProvider) AppLogoutRedirectURL() string {
	return e.appLogoutRedirectURL
}

func (e *EnvProvider) KratosPublicURL() string {
	return e.kratosPublicURL
}

func (e *EnvProvider) KratosAdminURL() string {
	return e.kratosAdminURL
}
