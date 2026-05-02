// SPDX-License-Identifier: Apache-2.0

package httpc_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/nathanbrophy/glacier/httpc"
)

// ExampleGet demonstrates how to send a typed GET request. The response body
// is automatically JSON-decoded into the target type.
func ExampleGet() {
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	// Use a fake transport so this example runs without a network.
	fakeRT := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body := `{"id":42,"name":"Ada"}`
		return &http.Response{
			StatusCode: 200,
			Status:     "200 OK",
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})
	c := httpc.New(httpc.WithTransport(fakeRT))

	ctx := context.Background()
	user, resp, err := httpc.GetWith[User](c, ctx, "https://api.example.com/users/42")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer resp.Drain()

	fmt.Printf("%s (status %d)\n", user.Name, resp.StatusCode)
	// Output: Ada (status 200)
}

// ExampleJSONBody demonstrates a POST with a closure-generated JSON body.
// The closure is invoked once per retry attempt, making it retry-safe.
func ExampleJSONBody() {
	type NewUser struct {
		Name string `json:"name"`
	}
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	fakeRT := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 201,
			Status:     "201 Created",
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"id":1,"name":"Ada"}`)),
		}, nil
	})
	c := httpc.New(httpc.WithTransport(fakeRT))

	ctx := context.Background()
	created, _, err := httpc.PostWith[User](c, ctx, "https://api.example.com/users",
		httpc.JSONBody(func() NewUser {
			return NewUser{Name: "Ada"}
		}),
	)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("created user %d\n", created.ID)
	// Output: created user 1
}

// ExampleWithDryRun demonstrates ctx-propagated dry-run: no network call is
// made; a *RequestPlan is emitted to the sink instead.
func ExampleWithDryRun() {
	c := httpc.New()

	var plans []*httpc.RequestPlan
	ctx := httpc.WithDryRun(context.Background(),
		httpc.WithPlanSink(func(p *httpc.RequestPlan) {
			plans = append(plans, p)
		}),
	)

	// No network call is made; plan is captured.
	_, _, _ = httpc.GetWith[any](c, ctx, "https://api.example.com/users/42")

	fmt.Printf("would %s %s\n", plans[0].Request.Method, plans[0].Request.URL)
	// Output: would GET https://api.example.com/users/42
}
