package db

import (
	"fmt"
	"goforge/internal/domain/params"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	DSN string `json:"dsn"`
	AdminDSN string
	DatabaseName string
	Skip bool
	Force bool
}

func ParseConfig(p params.Params) (Config, error) {
	dsn := strings.TrimSpace(p.Param("dsn"))

	if dsn == "" {
		env := strings.TrimSpace(p.Param("env"))

		if env == "" {
			env = "development"
		}

		fileDSN, err := loadDSNFromConfig(env)

		if err != nil {
			return Config{}, err
		}

		dsn = fileDSN
	}

	u, err := url.Parse(dsn)

	if err != nil {
		return Config{}, fmt.Errorf("invalid DSN: %w", err)
	}

	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return Config{}, fmt.Errorf("unsupported DSN scheme %q (only postgres/postgresql supported)", u.Scheme)
	}

	dbName := strings.TrimPrefix(path.Clean(u.Path), "/")
	if dbName == "" || dbName == "." {
		return Config{}, fmt.Errorf("dsn must include target database name in path")
	}

	adminURL := *u
	adminURL.Path = "/postgres"

	return Config{
		DSN: dsn,
		AdminDSN: adminURL.String(),
		Skip: p.BoolParam("skip"),
		Force: p.BoolParam("force"),
		DatabaseName: dbName,
	}, nil
}

func loadDSNFromConfig(env string) (string, error) {
	configPath := filepath.Join("config", "database.toml")

	data, err := os.ReadFile(configPath)

	if err != nil {
		return "", err
	}

	var configByEnv map[string]Config

	if err := toml.Unmarshal(data, &configByEnv); err != nil {
		return "", err
	}

	dbConfig, ok := configByEnv[env]

	if !ok {
		return "", fmt.Errorf("no database config found for env %s", env)
	}

	dsn := dbConfig.DSN

	if dsn == "" {
		return "", fmt.Errorf("no database DSN found for env %s", env)
	}

	return dsn, nil
}
