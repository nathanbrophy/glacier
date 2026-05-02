// SPDX-License-Identifier: Apache-2.0

package httpmock_test

import (
	"fmt"
	"io"
	"net/http"

	"github.com/nathanbrophy/glacier/httpmock"
)

// ExampleTransport_OnRequest demonstrates the headline usage pattern:
// plug httpmock.Transport into an http.Client, declare a stub, and exercise
// the code under test with zero real network calls.
func ExampleTransport_OnRequest() {
	// Create a transport (no *testing.T available in Example functions).
	rt := httpmock.New()

	// Register a stub: GET /users/42 → 200 JSON.
	rt.OnRequest().
		Method("GET").
		Path("/users/42").
		Respond(httpmock.Status(200))

	// Plug the transport into an http.Client.
	client := &http.Client{Transport: rt}

	resp, err := client.Get("https://api.example.com/users/42")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	fmt.Println("status:", resp.StatusCode)
	// Output: status: 200
}
