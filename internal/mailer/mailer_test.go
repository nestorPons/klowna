package mailer

import (
	"encoding/base64"
	"net/mail"
	"strings"
	"testing"

	"klowna/internal/model"
)

func TestLeadMessage(t *testing.T) {
	m := New("smtp.example.com", "587", "web@example.com", "secret", "", "destino@example.com")
	lead := model.Lead{
		Name:      "Ana\r\nBcc: atacante@example.com",
		Email:     "ana@example.com",
		Phone:     "+34 600 000 000",
		Clinic:    "Clínica Sol",
		Sector:    "Belleza",
		Message:   "Hola,\r\nquiero una demo.",
		Source:    "web",
		IP:        "203.0.113.9",
		UserAgent: "Mozilla/5.0",
	}

	raw, err := m.leadMessage(lead)
	if err != nil {
		t.Fatalf("leadMessage: %v", err)
	}

	msg, err := mail.ReadMessage(strings.NewReader(string(raw)))
	if err != nil {
		t.Fatalf("mensaje no válido: %v", err)
	}
	to, err := mail.ParseAddress(msg.Header.Get("To"))
	if err != nil {
		t.Fatalf("To no válido: %v", err)
	}
	if to.Address != "destino@example.com" {
		t.Fatalf("To = %q", to.Address)
	}
	if strings.Contains(msg.Header.Get("Subject"), "\n") {
		t.Fatalf("cabecera Subject con salto de línea: %q", msg.Header.Get("Subject"))
	}
	if reply := msg.Header.Get("Reply-To"); !strings.Contains(reply, "ana@example.com") {
		t.Fatalf("Reply-To = %q", reply)
	}

	body, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(strings.SplitN(string(raw), "\r\n\r\n", 2)[1], "\r\n", ""))
	if err != nil {
		t.Fatalf("cuerpo base64: %v", err)
	}
	for _, want := range []string{"Ana Bcc: atacante@example.com", "ana@example.com", "Clínica Sol", "quiero una demo."} {
		if !strings.Contains(string(body), want) {
			t.Errorf("el cuerpo no contiene %q:\n%s", want, body)
		}
	}
}

func TestLeadMessageEscapesHeaderInjection(t *testing.T) {
	m := New("smtp.example.com", "587", "web@example.com", "secret", "", "destino@example.com")
	msg, err := m.leadMessage(model.Lead{Name: "Eva\nBcc: otro@example.com", Email: "eva@example.com"})
	if err != nil {
		t.Fatalf("leadMessage: %v", err)
	}
	if strings.Contains(strings.SplitN(string(msg), "\r\n\r\n", 2)[0], "\nBcc:") {
		t.Fatalf("inyección de cabecera no neutralizada")
	}
}

func TestEnabled(t *testing.T) {
	if New("", "587", "", "", "", "").Enabled() {
		t.Fatal("un mailer sin host y destino no debe estar activo")
	}
	if !New("smtp.example.com", "", "web@example.com", "", "", "destino@example.com").Enabled() {
		t.Fatal("el mailer debería estar activo")
	}
}
