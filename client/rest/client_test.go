// Copyright (C) 2024 Specter Ops, Inc.
//
// This file is part of AzureHound.
//
// AzureHound is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// AzureHound is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package rest

import (
	"net/http"
	"net/http/httptest"
	"time"

	"testing"

	"github.com/SpecterOps/AzureHound/v2/client/config"
)

func TestClosedConnection(t *testing.T) {
	var testServer *httptest.Server
	attempt := 0
	var mockHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		attempt++
		testServer.CloseClientConnections()
	}

	testServer = httptest.NewServer(mockHandler)
	defer testServer.Close()

	defaultConfig := config.Config{
		Username:  "azurehound",
		Password:  "we_collect",
		Authority: testServer.URL,
	}

	if client, err := NewRestClient(testServer.URL, defaultConfig); err != nil {
		t.Fatalf("error initializing rest client %v", err)
	} else {
		requestCompleted := false

		// make request in separate goroutine so its not blocking after we validated the retry
		go func() {
			req := httptest.NewRequest("GET", testServer.URL, nil)
			client.AddAuthenticationToRequest(req) // Authenticate() because it uses the internal client.send method.
			// CloseClientConnections should block the request from completing, however if it completes then the test fails.
			requestCompleted = true
		}()

		// block until attempt is > 2 or request succeeds
		for attempt <= 2 {
			if attempt > 1 || requestCompleted {
				break
			}
		}

		if requestCompleted {
			t.Fatalf("expected an attempted retry but the request completed")
		}
	}
}

func TestParseRetryAfter(t *testing.T) {
	if got, ok := ParseRetryAfter("7"); !ok || got != 7*time.Second {
		t.Errorf("seconds: got %v %v", got, ok)
	}
	if got, ok := ParseRetryAfter("Wed, 21 Oct 2015 07:28:00 GMT"); !ok || got != 0 {
		t.Errorf("past date: got %v %v", got, ok)
	}
	if got, ok := ParseRetryAfter(time.Now().Add(time.Hour).UTC().Format(http.TimeFormat)); !ok || got <= 0 {
		t.Errorf("future date: got %v %v", got, ok)
	}
	for _, value := range []string{"", "soon", "-1"} {
		if _, ok := ParseRetryAfter(value); ok {
			t.Errorf("expected %q to be rejected", value)
		}
	}
}

func TestSendRetriesThrottledRequestWithDateRetryAfter(t *testing.T) {
	attempt := 0
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt == 1 {
			w.Header().Set("Retry-After", "Wed, 21 Oct 2015 07:28:00 GMT")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	client, err := NewRestClient(testServer.URL, config.Config{Authority: testServer.URL})
	if err != nil {
		t.Fatalf("error initializing rest client %v", err)
	}

	req, _ := http.NewRequest("GET", testServer.URL, nil)
	if res, err := client.(*restClient).send(req); err != nil {
		t.Fatalf("expected throttled request to be retried, got %v", err)
	} else if res.StatusCode != http.StatusOK || attempt != 2 {
		t.Fatalf("got status %d after %d attempts, want 200 after 2", res.StatusCode, attempt)
	}
}
