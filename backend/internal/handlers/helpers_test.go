package handlers

import "testing"

func TestValidateItemURL(t *testing.T) {
	allowed := []string{
		"https://example.com", "http://10.0.0.1:8080/x", "ssh://host", "smb://nas/share",
		"/relative/path", "example.com", "mailto:a@example.com", "https://example.com/?q=javascript:x",
	}
	for _, u := range allowed {
		if err := ValidateItemURL(u); err != nil {
			t.Errorf("%q rejected: %v", u, err)
		}
	}
	blocked := []string{
		"javascript:alert(1)", " JavaScript:alert(1)", "java\tscript:alert(1)", "\x01javascript:alert(1)",
		"vbscript:msgbox(1)", "data:text/html,<script>alert(1)</script>", "JAVA\nSCRIPT:alert(1)",
	}
	for _, u := range blocked {
		if err := ValidateItemURL(u); err == nil {
			t.Errorf("%q accepted", u)
		}
	}
}
