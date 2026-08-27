package handlers

import "testing"

func TestValidOIDCReturnURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		host string
		want string
	}{
		{name: "dev frontend port", raw: "http://localhost:5173/", host: "localhost:8080", want: "http://localhost:5173/"},
		{name: "same production host", raw: "https://dash.example.com/", host: "dash.example.com", want: "https://dash.example.com/"},
		{name: "external host rejected", raw: "https://evil.example/", host: "dash.example.com", want: ""},
		{name: "relative URL rejected", raw: "//evil.example/", host: "dash.example.com", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validOIDCReturnURL(tt.raw, tt.host); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
