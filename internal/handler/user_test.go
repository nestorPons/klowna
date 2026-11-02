package handler

import "testing"

func TestNormalizeAvatar(t *testing.T) {
	valid := []string{
		"",
		"data:image/png;base64,AAAA",
		"data:image/jpeg;base64,AAAA",
	}
	for _, v := range valid {
		if _, ok := normalizeAvatar(v); !ok {
			t.Fatalf("normalizeAvatar(%q) = false", v)
		}
	}
	invalid := []string{
		"data:text/html;base64,AAAA",
		"data:image/svg+xml;base64,AAAA",
		"javascript:alert(1)",
	}
	for _, v := range invalid {
		if _, ok := normalizeAvatar(v); ok {
			t.Fatalf("normalizeAvatar(%q) = true", v)
		}
	}
}

func TestDecodeAvatar(t *testing.T) {
	contentType, data, ok := decodeAvatar("data:image/png;base64,AAAA")
	if !ok || contentType != "image/png" || len(data) != 3 {
		t.Fatalf("decodeAvatar = %q %d %v", contentType, len(data), ok)
	}
	if _, _, ok := decodeAvatar("data:image/png;base64,%%%"); ok {
		t.Fatal("decodeAvatar debería rechazar base64 inválido")
	}
}
