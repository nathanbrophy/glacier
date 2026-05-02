// SPDX-License-Identifier: Apache-2.0

package httpc_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/httpc"
)

func TestIsDryRun(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	assert.False(t, httpc.IsDryRun(ctx))

	dryCtx := httpc.WithDryRun(ctx)
	assert.True(t, httpc.IsDryRun(dryCtx))
}

func TestDryRunCapturesPlan(t *testing.T) {
	t.Parallel()
	c := httpc.New(httpc.WithTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		t.Fatal("network call should not be made in dry-run mode")
		return nil, nil
	})))

	var plans []*httpc.RequestPlan
	ctx := httpc.WithDryRun(context.Background(),
		httpc.WithPlanSink(func(p *httpc.RequestPlan) {
			plans = append(plans, p)
		}),
	)

	_, err := c.Get(ctx, "http://example.com/test")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(plans))
	assert.NotNil(t, plans[0].Request)
	assert.Equal(t, "GET", plans[0].Request.Method)
}

func TestDryRunReturnsZeroT(t *testing.T) {
	t.Parallel()
	c := httpc.New()
	ctx := httpc.WithDryRun(context.Background(),
		httpc.WithPlanSink(func(p *httpc.RequestPlan) {}),
	)
	user, resp, err := httpc.GetWith[testUser](c, ctx, "http://example.com/")
	assert.NoError(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, testUser{}, user)
}

func TestDryRunErrorsMode(t *testing.T) {
	t.Parallel()
	c := httpc.New()
	ctx := httpc.WithDryRun(context.Background(),
		httpc.WithPlanSink(func(p *httpc.RequestPlan) {}),
		httpc.WithDryRunErrors(),
	)
	_, _, err := httpc.GetWith[testUser](c, ctx, "http://example.com/")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, httpc.ErrDryRun))
}

func TestErrDryRunSentinel(t *testing.T) {
	t.Parallel()
	assert.True(t, errors.Is(httpc.ErrDryRun, httpc.ErrDryRun))
}

func TestRequestPlanRedactsAuthorization(t *testing.T) {
	t.Parallel()
	c := httpc.New(httpc.WithHeaders(http.Header{
		"Authorization": []string{"Bearer secret-token"},
	}))

	var plan *httpc.RequestPlan
	ctx := httpc.WithDryRun(context.Background(),
		httpc.WithPlanSink(func(p *httpc.RequestPlan) { plan = p }),
	)
	_, err := c.Get(ctx, "http://example.com/")
	assert.NoError(t, err)
	assert.NotNil(t, plan)
	vals := plan.Request.Header["Authorization"]
	assert.Equal(t, 1, len(vals))
	assert.Equal(t, "[REDACTED]", vals[0])
}

func TestRequestPlanRedactsCookie(t *testing.T) {
	t.Parallel()
	c := httpc.New(httpc.WithHeaders(http.Header{
		"Cookie": []string{"session=abc"},
	}))

	var plan *httpc.RequestPlan
	ctx := httpc.WithDryRun(context.Background(),
		httpc.WithPlanSink(func(p *httpc.RequestPlan) { plan = p }),
	)
	_, err := c.Get(ctx, "http://example.com/")
	assert.NoError(t, err)
	assert.NotNil(t, plan)
	vals := plan.Request.Header["Cookie"]
	assert.Equal(t, 1, len(vals))
	assert.Equal(t, "[REDACTED]", vals[0])
}

func TestRequestPlanRedactsXApiKey(t *testing.T) {
	t.Parallel()
	c := httpc.New(httpc.WithHeaders(http.Header{
		"X-Api-Key": []string{"my-key"},
	}))

	var plan *httpc.RequestPlan
	ctx := httpc.WithDryRun(context.Background(),
		httpc.WithPlanSink(func(p *httpc.RequestPlan) { plan = p }),
	)
	_, err := c.Get(ctx, "http://example.com/")
	assert.NoError(t, err)
	vals := plan.Request.Header["X-Api-Key"]
	assert.Equal(t, "[REDACTED]", vals[0])
}

func TestRequestPlanRedactsXAuthToken(t *testing.T) {
	t.Parallel()
	c := httpc.New(httpc.WithHeaders(http.Header{
		"X-Auth-Token": []string{"tok"},
	}))

	var plan *httpc.RequestPlan
	ctx := httpc.WithDryRun(context.Background(),
		httpc.WithPlanSink(func(p *httpc.RequestPlan) { plan = p }),
	)
	_, err := c.Get(ctx, "http://example.com/")
	assert.NoError(t, err)
	vals := plan.Request.Header["X-Auth-Token"]
	assert.Equal(t, "[REDACTED]", vals[0])
}

func TestRequestPlanRedactsProxyAuthorization(t *testing.T) {
	t.Parallel()
	c := httpc.New(httpc.WithHeaders(http.Header{
		"Proxy-Authorization": []string{"Basic abc"},
	}))

	var plan *httpc.RequestPlan
	ctx := httpc.WithDryRun(context.Background(),
		httpc.WithPlanSink(func(p *httpc.RequestPlan) { plan = p }),
	)
	_, err := c.Get(ctx, "http://example.com/")
	assert.NoError(t, err)
	vals := plan.Request.Header["Proxy-Authorization"]
	assert.Equal(t, "[REDACTED]", vals[0])
}

func TestRequestPlanRedactsByRegex(t *testing.T) {
	t.Parallel()
	// "X-Auth-Foo" matches (?i)auth pattern.
	c := httpc.New(httpc.WithHeaders(http.Header{
		"X-Auth-Foo": []string{"secret-value"},
	}))

	var plan *httpc.RequestPlan
	ctx := httpc.WithDryRun(context.Background(),
		httpc.WithPlanSink(func(p *httpc.RequestPlan) { plan = p }),
	)
	_, err := c.Get(ctx, "http://example.com/")
	assert.NoError(t, err)
	vals := plan.Request.Header["X-Auth-Foo"]
	assert.Equal(t, "[REDACTED]", vals[0])
}

func TestWithPlanIncludeSecrets(t *testing.T) {
	t.Parallel()
	c := httpc.New(httpc.WithHeaders(http.Header{
		"Authorization": []string{"Bearer real-token"},
	}))

	var plan *httpc.RequestPlan
	ctx := httpc.WithDryRun(context.Background(),
		httpc.WithPlanSink(func(p *httpc.RequestPlan) { plan = p }),
		httpc.WithPlanIncludeSecrets(),
	)
	_, err := c.Get(ctx, "http://example.com/")
	assert.NoError(t, err)
	vals := plan.Request.Header["Authorization"]
	assert.Equal(t, "Bearer real-token", vals[0])
}

func TestDryRunSkipsRetryLoop(t *testing.T) {
	t.Parallel()
	var planCount int
	c := httpc.New()
	ctx := httpc.WithDryRun(context.Background(),
		httpc.WithPlanSink(func(p *httpc.RequestPlan) { planCount++ }),
	)
	_, err := c.Get(ctx, "http://example.com/",
		httpc.WithRetry(httpc.MaxAttempts(5)),
	)
	assert.NoError(t, err)
	// Plan emitted exactly once; retry loop not entered.
	assert.Equal(t, 1, planCount)
}

func TestStaticDryRunHelperOnClient(t *testing.T) {
	t.Parallel()
	var captured *httpc.RequestPlan
	c := httpc.New(httpc.WithTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		t.Fatal("should not reach transport in dry-run mode")
		return nil, nil
	})))
	ctx := httpc.WithDryRun(context.Background(),
		httpc.WithPlanSink(func(p *httpc.RequestPlan) { captured = p }),
	)
	_, err := c.Get(ctx, "http://example.com/api")
	assert.NoError(t, err)
	assert.NotNil(t, captured)
}
