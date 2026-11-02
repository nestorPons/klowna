package captcha

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVerifySuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("secret") != "clave" || r.FormValue("response") != "token" {
			t.Errorf("form = %v", r.Form)
		}
		if r.FormValue("remoteip") != "203.0.113.9" {
			t.Errorf("remoteip = %q", r.FormValue("remoteip"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer srv.Close()

	v := New("clave", false)
	v.Endpoint = srv.URL
	if err := v.Verify(context.Background(), "token", "203.0.113.9"); err != nil {
		t.Fatalf("Verify: %v", err)
	}
}

func TestVerifyRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":false,"error-codes":["invalid-input-response"]}`))
	}))
	defer srv.Close()

	v := New("clave", false)
	v.Endpoint = srv.URL
	err := v.Verify(context.Background(), "token", "")
	if err == nil {
		t.Fatal("se esperaba error")
	}
	if !contains(err.Error(), "invalid-input-response") {
		t.Fatalf("error inesperado: %v", err)
	}
}

func TestVerifyWithoutConfig(t *testing.T) {
	v := New("", false)
	if err := v.Verify(context.Background(), "token", ""); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("err = %v", err)
	}
	if err := v.Verify(context.Background(), "", ""); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("err = %v", err)
	}
}

func TestVerifyMissingToken(t *testing.T) {
	v := New("clave", false)
	if err := v.Verify(context.Background(), "", ""); !errors.Is(err, ErrMissingToken) {
		t.Fatalf("err = %v", err)
	}
}

func TestSkip(t *testing.T) {
	v := New("", true)
	if err := v.Verify(context.Background(), "", ""); err != nil {
		t.Fatalf("con Skip no debe verificar: %v", err)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
