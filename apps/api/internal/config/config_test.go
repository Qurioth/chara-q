package config

import (
	"reflect"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    Config
		wantErr bool
	}{
		{
			name: "必須項目だけなら既定値を使う",
			env:  map[string]string{"DATABASE_URL": "postgres://localhost/test"},
			want: Config{Port: 8080, DatabaseURL: "postgres://localhost/test"},
		},
		{
			name: "すべて指定する",
			env: map[string]string{
				"PORT":                 "9000",
				"DATABASE_URL":         "postgres://localhost/test",
				"CORS_ALLOWED_ORIGINS": "http://localhost:3000, https://example.com ,",
				"APP_ENV":              "development",
			},
			want: Config{
				Port:               9000,
				DatabaseURL:        "postgres://localhost/test",
				CORSAllowedOrigins: []string{"http://localhost:3000", "https://example.com"},
				Development:        true,
			},
		},
		{
			name:    "DATABASE_URL がないとエラー",
			env:     map[string]string{},
			wantErr: true,
		},
		{
			name:    "PORT が数値でないとエラー",
			env:     map[string]string{"PORT": "abc", "DATABASE_URL": "postgres://localhost/test"},
			wantErr: true,
		},
		{
			name:    "PORT が範囲外だとエラー",
			env:     map[string]string{"PORT": "70000", "DATABASE_URL": "postgres://localhost/test"},
			wantErr: true,
		},
		{
			name:    "APP_ENV が不明な値だとエラー",
			env:     map[string]string{"APP_ENV": "staging", "DATABASE_URL": "postgres://localhost/test"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := load(func(k string) string { return tt.env[k] })
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
