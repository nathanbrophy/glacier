// SPDX-License-Identifier: Apache-2.0

package log_test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/nathanbrophy/glacier/log"
)

// ExampleNewHandler shows how to configure the Glacier text handler.
func ExampleNewHandler() {
	// Write to stderr with debug level and no color (suitable for CI).
	h := log.NewHandler(os.Stderr,
		log.WithLevel(log.LevelDebug),
		log.WithColor(log.ColorNever),
	)
	l := slog.New(h)
	l.Info("server started", "port", 8080)
}

// ExampleWith shows how to attach attributes to a context so every log call
// made with that context includes them automatically.
func ExampleWith() {
	var buf bytes.Buffer
	h := log.NewHandler(&buf,
		log.WithColor(log.ColorNever),
	)
	l := slog.New(h)

	ctx := context.Background()
	ctx = log.With(ctx, slog.String("request_id", "req-001"))
	ctx = log.With(ctx, slog.String("user_id", "u-42"))

	l.InfoContext(ctx, "handled")
	fmt.Println(containsSubstring(buf.String(), "request_id=req-001"))
	// Output:
	// true
}

// ExampleRedact shows how to mark a value as sensitive so it is never logged
// in plain text, regardless of which handler receives it.
func ExampleRedact() {
	var buf bytes.Buffer
	h := log.NewHandler(&buf, log.WithColor(log.ColorNever))
	l := slog.New(h)

	apiKey := "sk-super-secret-key"
	l.Info("auth", "api_key", log.Redact(apiKey))

	fmt.Println(containsSubstring(buf.String(), "[REDACTED]"))
	fmt.Println(containsSubstring(buf.String(), apiKey))
	// Output:
	// true
	// false
}
