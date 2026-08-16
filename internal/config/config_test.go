package config

import "testing"

func TestPublicURLUsesLoopbackForWildcardAddress(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"0.0.0.0:8080": "http://127.0.0.1:8080",
		"[::]:8080":    "http://127.0.0.1:8080",
		"127.0.0.1:80": "http://127.0.0.1:80",
	}
	for address, want := range tests {
		if got := (Config{Address: address}).PublicURL(); got != want {
			t.Errorf("PublicURL() for %q = %q, want %q", address, got, want)
		}
	}
}

func TestPublicURLPrefersConfiguredBaseURL(t *testing.T) {
	t.Parallel()

	got := (Config{Address: "0.0.0.0:8080", BaseURL: "https://analytics.example.com/"}).PublicURL()
	if got != "https://analytics.example.com" {
		t.Fatalf("PublicURL() = %q", got)
	}
}
