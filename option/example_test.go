// SPDX-License-Identifier: Apache-2.0

package option_test

import (
	"errors"
	"fmt"

	"github.com/nathanbrophy/glacier/option"
)

// exampleConfig is a simple config struct used in examples.
type exampleConfig struct {
	timeout int
	debug   bool
	label   string
}

func withTimeout(ms int) option.Option[exampleConfig] {
	return option.OptionFunc[exampleConfig](func(c *exampleConfig) error {
		if ms <= 0 {
			return fmt.Errorf("option: withtimeout: timeout must be positive, got %d", ms)
		}
		c.timeout = ms
		return nil
	})
}

func withDebug(v bool) option.Option[exampleConfig] {
	return option.OptionFunc[exampleConfig](func(c *exampleConfig) error {
		c.debug = v
		return nil
	})
}

func withLabel(s string) option.Option[exampleConfig] {
	return option.OptionFunc[exampleConfig](func(c *exampleConfig) error {
		c.label = s
		return nil
	})
}

// ---- T#45 ExampleApply ----

// ExampleApply demonstrates folding functional options into a config struct.
func ExampleApply() {
	cfg, err := option.Apply([]option.Option[exampleConfig]{
		withTimeout(100),
		withDebug(true),
		withLabel("prod"),
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("timeout=%d debug=%v label=%s\n", cfg.timeout, cfg.debug, cfg.label)
	// Output: timeout=100 debug=true label=prod
}

// ---- T#46 ExampleStrict ----

// ExampleStrict demonstrates accumulating all option errors in one pass.
func ExampleStrict() {
	bad1 := errors.New("option: withfoo: foo is required")
	bad2 := errors.New("option: withbar: bar must be positive")

	failFoo := option.OptionFunc[exampleConfig](func(_ *exampleConfig) error { return bad1 })
	failBar := option.OptionFunc[exampleConfig](func(_ *exampleConfig) error { return bad2 })

	_, err := option.Apply(
		[]option.Option[exampleConfig]{failFoo, withLabel("ok"), failBar},
		option.Strict(),
	)
	if err != nil {
		// Both errors are reported in a single pass.
		fmt.Println(errors.Is(err, bad1)) // true
		fmt.Println(errors.Is(err, bad2)) // true
	}
	// Output:
	// true
	// true
}

// ---- T#47 ExampleValidate ----

// ExampleValidate demonstrates post-Apply validation with Required.
func ExampleValidate() {
	type serverConfig struct {
		addr   *string
		port   int
		logger *int // stand-in for a real logger
	}

	addr := "localhost"
	cfg := serverConfig{
		addr: &addr,
		port: 8080,
		// logger intentionally left nil
	}

	// For pointer fields, the getter must return explicit nil when the field is nil.
	// A typed nil *int stored in `any` is not == nil (interface boxing).
	err := option.Validate(&cfg,
		option.Required[serverConfig]("logger", func(c *serverConfig) any {
			if c.logger == nil {
				return nil
			}
			return c.logger
		}),
	)
	if err != nil {
		fmt.Println(err)
	}
	// Output: option: required: field "logger" not set
}
