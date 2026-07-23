package web

import "testing"

func TestMergeMaskedSensitiveConfigPreservesStoredSecret(t *testing.T) {
	existing := `{"token":"abcdef123456"}`
	merged, err := mergeMaskedSensitiveConfig("pushplus", map[string]interface{}{
		"token":   "abc****456",
		"channel": "mail",
	}, existing)
	if err != nil {
		t.Fatalf("mergeMaskedSensitiveConfig failed: %v", err)
	}
	if merged["token"] != "abcdef123456" {
		t.Fatalf("stored token was not preserved: %v", merged["token"])
	}
}
