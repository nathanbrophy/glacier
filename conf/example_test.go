// SPDX-License-Identifier: Apache-2.0

package conf_test

import (
	"context"
	"fmt"

	"github.com/nathanbrophy/glacier/conf"
)

// ExampleDecode demonstrates one-shot decoding without the global registry.
func ExampleDecode() {
	type ServerConfig struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}

	cfg, err := conf.Decode[ServerConfig](
		context.Background(),
		conf.WithSet("host", "localhost"),
		conf.WithSet("port", 8080),
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("host=%s port=%d\n", cfg.Host, cfg.Port)
	// Output:
	// host=localhost port=8080
}

// ExampleRegister demonstrates section registration with an accessor.
func ExampleRegister() {
	type DBConfig struct {
		DSN     string `json:"dsn"`
		MaxConn int    `json:"max_conn"`
	}

	// Typically called at package init time.
	getDB := conf.Register("db", DBConfig{DSN: "postgres://localhost/dev", MaxConn: 10})

	// Before conf.Load, the accessor returns the registered defaults.
	db := getDB()
	fmt.Printf("dsn=%s max_conn=%d\n", db.DSN, db.MaxConn)
	// Output:
	// dsn=postgres://localhost/dev max_conn=10
}

// ExampleLoader_Load demonstrates loading config with a Loader.
func ExampleLoader_Load() {
	l := conf.NewLoader()

	// Load with no sources :  all registered sections keep their defaults.
	if err := l.Load(context.Background()); err != nil {
		panic(err)
	}
}

// ExampleWithDefaults demonstrates layering additional defaults on top of
// struct defaults.
func ExampleWithDefaults() {
	type AppConfig struct {
		LogLevel string `json:"log_level"`
		Workers  int    `json:"workers"`
	}

	cfg, err := conf.Decode[AppConfig](
		context.Background(),
		conf.WithDefaults(func() map[string]any {
			return map[string]any{
				"log_level": "info",
				"workers":   4,
			}
		}),
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("level=%s workers=%d\n", cfg.LogLevel, cfg.Workers)
	// Output:
	// level=info workers=4
}
