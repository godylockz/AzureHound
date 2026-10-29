package client

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/SpecterOps/AzureHound/v2/client/query"
	"github.com/stretchr/testify/require"
)

// fakeRestClient is a minimal test double for rest.RestClient that allows
// controlling the Get response per test case.
type fakeRestClient struct {
	getFunc  func(ctx context.Context, path string, params query.Params, headers map[string]string) (*http.Response, error)
	sendFunc func(req *http.Request) (*http.Response, error)
}

func (s *fakeRestClient) Get(ctx context.Context, path string, params query.Params, headers map[string]string) (*http.Response, error) {
	return s.getFunc(ctx, path, params, headers)
}

func (s *fakeRestClient) Delete(context.Context, string, interface{}, query.Params, map[string]string) (*http.Response, error) {
	return nil, nil
}
func (s *fakeRestClient) Patch(context.Context, string, interface{}, query.Params, map[string]string) (*http.Response, error) {
	return nil, nil
}
func (s *fakeRestClient) Post(context.Context, string, interface{}, query.Params, map[string]string) (*http.Response, error) {
	return nil, nil
}
func (s *fakeRestClient) Put(context.Context, string, interface{}, query.Params, map[string]string) (*http.Response, error) {
	return nil, nil
}
func (s *fakeRestClient) Send(req *http.Request) (*http.Response, error) {
	if s.sendFunc != nil {
		return s.sendFunc(req)
	}
	return nil, nil
}
func (s *fakeRestClient) AddAuthenticationToRequest(req *http.Request) (*http.Request, error) {
	return req, nil
}
func (s *fakeRestClient) CloseIdleConnections() {}

func TestGetAzureObjectList_SuccessfulResponse(t *testing.T) {
	body := `{"value": [{"id": "1"}, {"id": "2"}]}`
	client := &fakeRestClient{
		getFunc: func(ctx context.Context, path string, params query.Params, headers map[string]string) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		},
	}

	out := make(chan AzureResult[map[string]string])
	go getAzureObjectList(client, context.Background(), "/test/path", nil, out)

	var results []map[string]string
	for result := range out {
		require.NoError(t, result.Error)
		results = append(results, result.Ok)
	}

	require.Len(t, results, 2)
	require.Equal(t, "1", results[0]["id"])
	require.Equal(t, "2", results[1]["id"])
}

func TestGetAzureObjectList_HungResponseTimesOut(t *testing.T) {
	// Shorten the timeout so the test completes quickly
	original := pageRequestTimeout
	pageRequestTimeout = 500 * time.Millisecond
	defer func() { pageRequestTimeout = original }()

	client := &fakeRestClient{
		getFunc: func(ctx context.Context, path string, params query.Params, headers map[string]string) (*http.Response, error) {
			// Verify the context has a deadline (set by pageRequestTimeout)
			_, hasDeadline := ctx.Deadline()
			require.True(t, hasDeadline, "expected context passed to Get to have a deadline from pageRequestTimeout")

			// Simulate a hung connection: block until the context expires
			<-ctx.Done()
			return nil, ctx.Err()
		},
	}

	out := make(chan AzureResult[map[string]string])
	go getAzureObjectList(client, context.Background(), "/test/path", nil, out)

	// The channel should produce an error and close well within a few seconds
	select {
	case result, ok := <-out:
		require.True(t, ok, "expected a value on out, channel closed")
		require.Error(t, result.Error, "expected an error result from timed-out request")
		require.ErrorIs(t, result.Error, context.DeadlineExceeded)
	case <-time.After(5 * time.Second):
		t.Fatal("getAzureObjectList did not return within expected timeout; pipeline is hung")
	}

	// drain and ensure channel closes
	for range out {
	}
}

func TestGetAzureObjectList_ParentContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	client := &fakeRestClient{
		getFunc: func(ctx context.Context, path string, params query.Params, headers map[string]string) (*http.Response, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		},
	}

	out := make(chan AzureResult[map[string]string])
	go getAzureObjectList(client, ctx, "/test/path", nil, out)

	// Cancel the parent context after a short delay
	time.AfterFunc(100*time.Millisecond, cancel)

	select {
	case result, ok := <-out:
		if ok {
			require.Error(t, result.Error)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("getAzureObjectList did not respect parent context cancellation")
	}

	for range out {
	}
}

func jsonResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestGetAzureObjectList_SkipsUndecodableItems(t *testing.T) {
	client := &fakeRestClient{
		getFunc: func(ctx context.Context, path string, params query.Params, headers map[string]string) (*http.Response, error) {
			// The second item does not fit map[string]string and must not take the rest of the listing with it
			return jsonResponse(`{"value": [{"id": "1"}, {"id": 2}, {"id": "3"}], "nextLink": "https://example.test/next"}`), nil
		},
		sendFunc: func(req *http.Request) (*http.Response, error) {
			return jsonResponse(`{"value": [{"id": "4"}]}`), nil
		},
	}

	out := make(chan AzureResult[map[string]string])
	go getAzureObjectList(client, context.Background(), "/test/path", nil, out)

	var (
		ids  []string
		errs []error
	)
	for result := range out {
		if result.Error != nil {
			errs = append(errs, result.Error)
		} else {
			require.Empty(t, errs, "decode error must be reported after every item")
			ids = append(ids, result.Ok["id"])
		}
	}

	require.Equal(t, []string{"1", "3", "4"}, ids)
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Error(), "skipped 1 item(s) from /test/path")
}

func TestGetAzureObjectList_RetriesTimedOutPage(t *testing.T) {
	original := pageRequestTimeout
	pageRequestTimeout = 100 * time.Millisecond
	defer func() { pageRequestTimeout = original }()

	calls := 0
	client := &fakeRestClient{
		getFunc: func(ctx context.Context, path string, params query.Params, headers map[string]string) (*http.Response, error) {
			calls++
			if calls == 1 {
				<-ctx.Done()
				return nil, ctx.Err()
			}
			return jsonResponse(`{"value": [{"id": "1"}]}`), nil
		},
	}

	out := make(chan AzureResult[map[string]string])
	go getAzureObjectList(client, context.Background(), "/test/path", nil, out)

	var results []map[string]string
	for result := range out {
		require.NoError(t, result.Error)
		results = append(results, result.Ok)
	}

	require.Equal(t, 2, calls)
	require.Len(t, results, 1)
}
