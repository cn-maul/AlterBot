package web

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/cn-maul/Gentry/database"
)

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

func TestMonitorSnapshotResponseIncludesFormattedPrice(t *testing.T) {
	payload, err := json.Marshal(monitorSnapshotResponse{
		MonitorSnapshot: database.MonitorSnapshot{ItemKey: "sku-1", PriceMinor: 12345, PriceValid: true, Currency: "CNY"},
		PriceDisplay:    "¥123.45",
	})
	if err != nil {
		t.Fatalf("marshal snapshot response: %v", err)
	}
	encoded := string(payload)
	if !strings.Contains(encoded, `"price_minor":12345`) || !strings.Contains(encoded, `"price_display":"¥123.45"`) {
		t.Fatalf("snapshot response is missing price fields: %s", encoded)
	}
}

func TestDetectionFingerprintCoversFullFieldSemantics(t *testing.T) {
	fields := []fieldRequest{{Name: "price", Selector: ".price", Type: "text", Attr: "", Transform: "trim"}}
	base := computeDetectionFingerprint(
		"https://example.com", "body", "", fields, "field_transition",
		`{"type":"field_transition","identity":{"source":"source_url"},"conditions":[{"field":"price","value_type":"money","operator":"decreased"}],"on_first_baseline":"silent"}`,
		`{"price":"money"}`,
	)
	changedAttr := append([]fieldRequest(nil), fields...)
	changedAttr[0].Attr = "data-price"
	if base == computeDetectionFingerprint("https://example.com", "body", "", changedAttr, "field_transition", `{"type":"field_transition","identity":{"source":"source_url"},"conditions":[{"field":"price","value_type":"money","operator":"decreased"}],"on_first_baseline":"silent"}`, `{"price":"money"}`) {
		t.Error("changing Attr must change the detection fingerprint")
	}
	changedTransform := append([]fieldRequest(nil), fields...)
	changedTransform[0].Transform = "lower"
	if base == computeDetectionFingerprint("https://example.com", "body", "", changedTransform, "field_transition", `{"type":"field_transition","identity":{"source":"source_url"},"conditions":[{"field":"price","value_type":"money","operator":"decreased"}],"on_first_baseline":"silent"}`, `{"price":"money"}`) {
		t.Error("changing Transform must change the detection fingerprint")
	}
}

func TestDetectionFingerprintCanonicalizesOrder(t *testing.T) {
	fieldsA := []fieldRequest{
		{Name: "title", Selector: "h1", Type: "text"},
		{Name: "price", Selector: ".price", Type: "text"},
	}
	fieldsB := []fieldRequest{fieldsA[1], fieldsA[0]}
	configA := `{"type":"field_transition","identity":{"field":"title"},"conditions":[{"field":"price","value_type":"money","operator":"decreased"}],"on_first_baseline":"silent"}`
	configB := `{"on_first_baseline":"silent","conditions":[{"operator":"decreased","value_type":"money","field":"price"}],"identity":{"field":"title"},"type":"field_transition"}`
	fpA := computeDetectionFingerprint("https://example.com", "body", "", fieldsA, "field_transition", configA, `{"title":"text","price":"money"}`)
	fpB := computeDetectionFingerprint("https://example.com", "body", "", fieldsB, "field_transition", configB, `{"price":"money","title":"text"}`)
	if fpA != fpB {
		t.Error("field order and JSON key order should not change the detection fingerprint")
	}
}
