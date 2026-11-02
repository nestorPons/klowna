// Package mailer envía avisos por SMTP de las solicitudes recibidas en la web.
package mailer

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"klowna/internal/model"
)

const defaultTimeout = 10 * time.Second

// Mailer agrupa la configuración de salida. Si Host o To están vacíos, Enabled
// devuelve false y los envíos se omiten sin romper el alta del lead.
type Mailer struct {
	Host    string
	Port    string
	User    string
	Pass    string
	From    string
	To      string
	Timeout time.Duration
}

func New(host, port, user, pass, from, to string) *Mailer {
	if strings.TrimSpace(port) == "" {
		port = "587"
	}
	if strings.TrimSpace(from) == "" {
		from = user
	}
	return &Mailer{
		Host:    strings.TrimSpace(host),
		Port:    strings.TrimSpace(port),
		User:    strings.TrimSpace(user),
		Pass:    pass,
		From:    strings.TrimSpace(from),
		To:      strings.TrimSpace(to),
		Timeout: defaultTimeout,
	}
}

func (m *Mailer) Enabled() bool {
	return m != nil && m.Host != "" && m.To != "" && m.From != ""
}

// SendLead envía el aviso de una nueva solicitud. Devuelve error si el SMTP
// falla; el llamador decide si debe afectar a la respuesta al visitante.
func (m *Mailer) SendLead(ctx context.Context, lead model.Lead) error {
	if !m.Enabled() {
		return nil
	}
	msg, err := m.leadMessage(lead)
	if err != nil {
		return err
	}
	if err := m.send(ctx, msg); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}

func (m *Mailer) leadMessage(lead model.Lead) ([]byte, error) {
	from := &mail.Address{Name: "Klowna web", Address: m.From}
	to := &mail.Address{Address: m.To}
	subject := mime.QEncoding.Encode("utf-8", "Nueva solicitud de demo · "+headerSafe(lead.Name))

	var b bytes.Buffer
	fmt.Fprintf(&b, "From: %s\r\n", from.String())
	fmt.Fprintf(&b, "To: %s\r\n", to.String())
	if lead.Email != "" {
		reply := &mail.Address{Name: headerSafe(lead.Name), Address: headerSafe(lead.Email)}
		fmt.Fprintf(&b, "Reply-To: %s\r\n", reply.String())
	}
	fmt.Fprintf(&b, "Subject: %s\r\n", subject)
	fmt.Fprintf(&b, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	fmt.Fprintf(&b, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&b, "Content-Type: text/plain; charset=UTF-8\r\n")
	fmt.Fprintf(&b, "Content-Transfer-Encoding: base64\r\n")
	fmt.Fprintf(&b, "\r\n")

	body := m.leadBody(lead)
	encoded := base64.StdEncoding.EncodeToString([]byte(body))
	for len(encoded) > 76 {
		fmt.Fprintf(&b, "%s\r\n", encoded[:76])
		encoded = encoded[76:]
	}
	fmt.Fprintf(&b, "%s\r\n", encoded)

	return b.Bytes(), nil
}

func (m *Mailer) leadBody(lead model.Lead) string {
	var b strings.Builder
	b.WriteString("Nueva solicitud desde la web\r\n\r\n")
	writeField(&b, "Nombre", lead.Name)
	writeField(&b, "Email", lead.Email)
	writeField(&b, "Teléfono", lead.Phone)
	writeField(&b, "Clínica", lead.Clinic)
	writeField(&b, "Sector", lead.Sector)
	writeField(&b, "Origen", lead.Source)
	writeField(&b, "IP", lead.IP)
	writeField(&b, "Navegador", lead.UserAgent)
	b.WriteString("\r\nMensaje:\r\n")
	if strings.TrimSpace(lead.Message) == "" {
		b.WriteString("(sin mensaje)\r\n")
	} else {
		b.WriteString(lead.Message)
		b.WriteString("\r\n")
	}
	b.WriteString("\r\nRecibido: ")
	b.WriteString(time.Now().Format("02/01/2006 15:04"))
	b.WriteString("\r\n")
	return b.String()
}

func (m *Mailer) send(ctx context.Context, msg []byte) error {
	timeout := m.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	addr := net.JoinHostPort(m.Host, m.Port)
	dialer := &net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("dial %s: %w", addr, err)
	}
	defer conn.Close()

	deadline := time.Now().Add(timeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	_ = conn.SetDeadline(deadline)

	client, err := m.newClient(conn, ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	if m.Port != "465" {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: m.Host, MinVersion: tls.VersionTLS12}); err != nil {
				return fmt.Errorf("starttls: %w", err)
			}
		}
	}

	if m.User != "" {
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(smtp.PlainAuth("", m.User, m.Pass, m.Host)); err != nil {
				return fmt.Errorf("auth: %w", err)
			}
		}
	}

	if err := client.Mail(m.From); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	for _, rcpt := range strings.Split(m.To, ",") {
		rcpt = strings.TrimSpace(rcpt)
		if rcpt == "" {
			continue
		}
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("rcpt %s: %w", rcpt, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("write message: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close message: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("quit: %w", err)
	}
	return nil
}

func (m *Mailer) newClient(conn net.Conn, ctx context.Context) (*smtp.Client, error) {
	if m.Port != "465" {
		client, err := smtp.NewClient(conn, m.Host)
		if err != nil {
			return nil, fmt.Errorf("smtp client: %w", err)
		}
		return client, nil
	}
	tlsConn := tls.Client(conn, &tls.Config{ServerName: m.Host, MinVersion: tls.VersionTLS12})
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		return nil, fmt.Errorf("implicit tls: %w", err)
	}
	client, err := smtp.NewClient(tlsConn, m.Host)
	if err != nil {
		return nil, fmt.Errorf("smtp client: %w", err)
	}
	return client, nil
}

func writeField(b *strings.Builder, label, value string) {
	b.WriteString(label)
	b.WriteString(": ")
	value = headerSafe(value)
	if value == "" {
		value = "—"
	}
	b.WriteString(value)
	b.WriteString("\r\n")
}

// headerSafe evita inyecciones de cabecera en los datos del formulario.
func headerSafe(s string) string {
	return strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ").Replace(strings.TrimSpace(s))
}
