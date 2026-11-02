// Package captcha valida tokens de Cloudflare Turnstile en el servidor.
package captcha

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultEndpoint = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

var (
	ErrNotConfigured = errors.New("turnstile: falta TURNSTILE_SECRET")
	ErrMissingToken  = errors.New("turnstile: token vacío")
)

type Verifier struct {
	Secret   string
	Endpoint string
	Client   *http.Client
	// Skip desactiva la validación (solo desarrollo con TURNSTILE_SKIP=true).
	Skip bool
}

func New(secret string, skip bool) *Verifier {
	return &Verifier{
		Secret:   strings.TrimSpace(secret),
		Endpoint: defaultEndpoint,
		Client:   &http.Client{Timeout: 8 * time.Second},
		Skip:     skip,
	}
}

func (v *Verifier) Enabled() bool {
	return v != nil && (v.Skip || v.Secret != "")
}

// Verify comprueba el token contra siteverify. remoteIP es opcional.
func (v *Verifier) Verify(ctx context.Context, token, remoteIP string) error {
	if v.Skip {
		return nil
	}
	if v.Secret == "" {
		return ErrNotConfigured
	}
	if strings.TrimSpace(token) == "" {
		return ErrMissingToken
	}

	form := url.Values{
		"secret":   {v.Secret},
		"response": {token},
	}
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}

	endpoint := v.Endpoint
	if endpoint == "" {
		endpoint = defaultEndpoint
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("turnstile request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := v.Client
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("turnstile siteverify: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if err != nil {
		return fmt.Errorf("turnstile read: %w", err)
	}
	var result struct {
		Success    bool     `json:"success"`
		ErrorCodes []string `json:"error-codes"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("turnstile decode: %w", err)
	}
	if !result.Success {
		if len(result.ErrorCodes) > 0 {
			return fmt.Errorf("turnstile rechazó el token: %s", strings.Join(result.ErrorCodes, ", "))
		}
		return fmt.Errorf("turnstile rechazó el token")
	}
	return nil
}
