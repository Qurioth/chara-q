// Package config は環境変数から API の設定を読み込む。
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config は API の実行時設定。
type Config struct {
	// Port は待ち受けるポート。Cloud Run が渡す PORT に従う。
	Port int
	// DatabaseURL は Postgres の接続文字列。
	DatabaseURL string
	// CORSAllowedOrigins は CORS で許可するオリジンの一覧。
	CORSAllowedOrigins []string
	// Development が true のときだけ GraphQL Playground とイントロスペクションを有効にする。
	Development bool
}

// Load は環境変数から設定を読み込む。
func Load() (Config, error) {
	return load(os.Getenv)
}

func load(getenv func(string) string) (Config, error) {
	var errs []error

	port := 8080
	if v := getenv("PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil || p < 1 || p > 65535 {
			errs = append(errs, fmt.Errorf("PORT が不正です: %q", v))
		} else {
			port = p
		}
	}

	databaseURL := getenv("DATABASE_URL")
	if databaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL が設定されていません"))
	}

	var origins []string
	for o := range strings.SplitSeq(getenv("CORS_ALLOWED_ORIGINS"), ",") {
		if o = strings.TrimSpace(o); o != "" {
			origins = append(origins, o)
		}
	}

	var development bool
	switch env := getenv("APP_ENV"); env {
	case "", "production":
	case "development":
		development = true
	default:
		errs = append(errs, fmt.Errorf("APP_ENV が不正です: %q（development か production）", env))
	}

	if err := errors.Join(errs...); err != nil {
		return Config{}, err
	}
	return Config{
		Port:               port,
		DatabaseURL:        databaseURL,
		CORSAllowedOrigins: origins,
		Development:        development,
	}, nil
}
