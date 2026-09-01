package updatecheck

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

func jsonClient(body string, onRequest func()) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		if onRequest != nil {
			onRequest()
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})}
}

func TestInfoReportsNewRelease(t *testing.T) {
	checker := New("v1.2.3", "abc123", true, "")
	checker.client = jsonClient(`{"tag_name":"v1.3.0","html_url":"https://example.test/release"}`, nil)

	info := checker.Info(context.Background())
	if !info.UpdateAvailable || info.LatestVersion != "v1.3.0" || info.ReleaseURL != "https://example.test/release" {
		t.Fatalf("unexpected info: %#v", info)
	}
}

func TestDevelopmentBuildDoesNotCheckForUpdates(t *testing.T) {
	called := false
	checker := New("dev", "unknown", true, "")
	checker.client = jsonClient(`{"tag_name":"v9.9.9"}`, func() { called = true })

	info := checker.Info(context.Background())
	if called || info.UpdateAvailable || info.Version != "dev" {
		t.Fatalf("unexpected info: %#v, called: %v", info, called)
	}
}

func TestCurrentDoesNotCheckForUpdates(t *testing.T) {
	called := false
	checker := New("v1.0.0", "abc123", true, "")
	checker.client = jsonClient(`{"tag_name":"v9.9.9"}`, func() { called = true })

	info := checker.Current()
	if called || info.Version != "v1.0.0" || info.LatestVersion != "" || info.UpdateAvailable {
		t.Fatalf("unexpected current info: %#v, called: %v", info, called)
	}
}

func TestLatestVersionOverrideSupportsLocalTesting(t *testing.T) {
	checker := New("v1.0.0", "local", true, "v9.9.9")
	info := checker.Info(context.Background())
	if !info.UpdateAvailable || info.LatestVersion != "v9.9.9" {
		t.Fatalf("unexpected info: %#v", info)
	}
}

func TestReleaseResponseIsCached(t *testing.T) {
	calls := 0
	checker := New("v1.0.0", "local", true, "")
	checker.client = jsonClient(`{"tag_name":"v1.0.1"}`, func() { calls++ })
	checker.cacheDuration = time.Hour

	checker.Info(context.Background())
	checker.Info(context.Background())
	if calls != 1 {
		t.Fatalf("release API calls = %d", calls)
	}
}

func TestNewer(t *testing.T) {
	for _, test := range []struct {
		candidate string
		current   string
		want      bool
	}{
		{"v1.2.0", "v1.1.9", true},
		{"v2.0.0", "v1.99.99", true},
		{"v1.2.3", "v1.2.3", false},
		{"v1.2.2", "v1.2.3", false},
		{"main", "v1.2.3", false},
	} {
		if got := newer(test.candidate, test.current); got != test.want {
			t.Errorf("newer(%q, %q) = %v, want %v", test.candidate, test.current, got, test.want)
		}
	}
}
