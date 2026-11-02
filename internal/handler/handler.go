package handler

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"klowna/internal/captcha"
	"klowna/internal/config"
	"klowna/internal/mailer"
	"klowna/internal/model"
	"klowna/internal/render"
	"klowna/internal/store"
	"klowna/web"
)

var emailRE = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

type Handler struct {
	store    *store.Store
	render   *render.Renderer
	cfg      config.Config
	mailer   *mailer.Mailer
	captcha  *captcha.Verifier
	staticFS fs.FS
}

func New(st *store.Store, r *render.Renderer, cfg config.Config) (*Handler, error) {
	staticFS, err := fs.Sub(web.Static, "static")
	if err != nil {
		return nil, fmt.Errorf("static fs: %w", err)
	}
	return &Handler{
		store:    st,
		render:   r,
		cfg:      cfg,
		mailer:   mailer.New(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom, cfg.SMTPTo),
		captcha:  captcha.New(cfg.TurnstileSecret, cfg.TurnstileSkip),
		staticFS: staticFS,
	}, nil
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", h.Home)
	mux.HandleFunc("GET /precios", h.Pricing)
	mux.HandleFunc("GET /preguntas-frecuentes", h.FAQ)
	mux.HandleFunc("GET /contacto", h.Contact)
	mux.HandleFunc("POST /contacto", h.CreateLead)
	mux.HandleFunc("POST /api/leads", h.CreateLead)
	mux.HandleFunc("GET /healthz", h.Health)

	mux.HandleFunc("GET /login", h.Login)
	mux.HandleFunc("POST /login", h.LoginPost)
	mux.HandleFunc("POST /logout", h.Logout)

	appMux := http.NewServeMux()
	appMux.HandleFunc("GET /app", h.Dashboard)
	appMux.HandleFunc("GET /app/agenda", h.Agenda)
	appMux.HandleFunc("POST /app/appointments", h.managerOnly(h.CreateAppointment))
	appMux.HandleFunc("POST /app/appointments/{id}", h.managerOnly(h.UpdateAppointment))
	appMux.HandleFunc("POST /app/appointments/{id}/move", h.managerOnly(h.MoveAppointment))
	appMux.HandleFunc("DELETE /app/appointments/{id}", h.managerOnly(h.DeleteAppointment))
	appMux.HandleFunc("POST /app/appointments/{id}/factura", h.managerOnly(h.CreateInvoiceFromAppointment))
	appMux.HandleFunc("GET /app/usuarios/{id}", h.UserProfile)
	appMux.HandleFunc("POST /app/usuarios/{id}", h.UpdateUserProfile)
	appMux.HandleFunc("GET /app/usuarios/{id}/avatar", h.UserAvatar)
	appMux.HandleFunc("GET /app/pacientes", h.managerOnly(h.Patients))
	appMux.HandleFunc("POST /app/pacientes", h.managerOnly(h.CreatePatient))
	appMux.HandleFunc("GET /app/pacientes/{id}", h.managerOnly(h.PatientDetail))
	appMux.HandleFunc("POST /app/pacientes/{id}", h.managerOnly(h.UpdatePatient))
	appMux.HandleFunc("POST /app/pacientes/{id}/historial", h.managerOnly(h.CreatePatientHistory))
	appMux.HandleFunc("POST /app/pacientes/{id}/historial/{entryID}", h.managerOnly(h.UpdatePatientHistory))
	appMux.HandleFunc("POST /app/pacientes/{id}/historial/{entryID}/eliminar", h.managerOnly(h.DeletePatientHistory))
	appMux.HandleFunc("GET /app/caja", h.managerOnly(h.Caja))
	appMux.HandleFunc("POST /app/caja/abrir", h.managerOnly(h.OpenCash))
	appMux.HandleFunc("POST /app/caja/movimiento", h.managerOnly(h.CreateCashMovement))
	appMux.HandleFunc("POST /app/caja/movimiento/{id}/eliminar", h.managerOnly(h.DeleteCashMovement))
	appMux.HandleFunc("POST /app/caja/cerrar", h.managerOnly(h.CloseCash))
	appMux.HandleFunc("GET /app/servicios", h.managerOnly(h.Services))
	appMux.HandleFunc("POST /app/servicios", h.managerOnly(h.CreateService))
	appMux.HandleFunc("POST /app/servicios/{id}", h.managerOnly(h.UpdateService))
	appMux.HandleFunc("DELETE /app/servicios/{id}", h.managerOnly(h.DeleteService))
	appMux.HandleFunc("GET /app/facturacion", h.managerOnly(h.Invoices))
	appMux.HandleFunc("POST /app/facturas", h.managerOnly(h.CreateInvoice))
	appMux.HandleFunc("POST /app/facturas/{id}/pagar", h.managerOnly(h.PayInvoice))
	appMux.HandleFunc("GET /app/facturas/{id}", h.managerOnly(h.InvoiceDetail))
	appMux.HandleFunc("GET /app/facturas/{id}/pdf", h.managerOnly(h.InvoicePDF))
	appMux.HandleFunc("GET /app/facturas/{id}/editar", h.managerOnly(h.InvoiceEdit))
	appMux.HandleFunc("POST /app/facturas/{id}", h.managerOnly(h.UpdateInvoice))
	appMux.HandleFunc("GET /app/leads", h.managerOnly(h.AppLeads))
	appMux.HandleFunc("GET /app/configuracion", h.managerOnly(h.SettingsRedirect))
	appMux.HandleFunc("GET /app/configuracion/{section}", h.managerOnly(h.Settings))
	appMux.HandleFunc("POST /app/configuracion", h.managerOnly(h.SaveSettings))
	appMux.HandleFunc("POST /app/configuracion/facturacion", h.managerOnly(h.SaveBillingSettings))
	appMux.HandleFunc("POST /app/configuracion/horario", h.managerOnly(h.SaveClinicSchedule))
	appMux.HandleFunc("POST /app/configuracion/bloque", h.managerOnly(h.CreateScheduleBlock))
	appMux.HandleFunc("POST /app/configuracion/bloque/{id}", h.managerOnly(h.UpdateScheduleBlock))
	appMux.HandleFunc("POST /app/configuracion/bloque/{id}/eliminar", h.managerOnly(h.DeleteScheduleBlock))
	appMux.HandleFunc("POST /app/configuracion/festivo", h.managerOnly(h.CreateHoliday))
	appMux.HandleFunc("POST /app/configuracion/festivo/{id}/eliminar", h.managerOnly(h.DeleteHoliday))
	appMux.HandleFunc("POST /app/configuracion/equipo/{id}", h.managerOnly(h.UpdateStaff))

	app := h.requireAuth(appMux)
	mux.Handle("GET /app", app)
	mux.Handle("GET /app/agenda", app)
	mux.Handle("POST /app/appointments", app)
	mux.Handle("POST /app/appointments/{id}", app)
	mux.Handle("POST /app/appointments/{id}/move", app)
	mux.Handle("DELETE /app/appointments/{id}", app)
	mux.Handle("POST /app/appointments/{id}/factura", app)
	mux.Handle("GET /app/usuarios/{id}", app)
	mux.Handle("POST /app/usuarios/{id}", app)
	mux.Handle("GET /app/usuarios/{id}/avatar", app)
	mux.Handle("GET /app/pacientes", app)
	mux.Handle("POST /app/pacientes", app)
	mux.Handle("GET /app/pacientes/{id}", app)
	mux.Handle("POST /app/pacientes/{id}", app)
	mux.Handle("POST /app/pacientes/{id}/historial", app)
	mux.Handle("POST /app/pacientes/{id}/historial/{entryID}", app)
	mux.Handle("POST /app/pacientes/{id}/historial/{entryID}/eliminar", app)
	mux.Handle("GET /app/caja", app)
	mux.Handle("POST /app/caja/abrir", app)
	mux.Handle("POST /app/caja/movimiento", app)
	mux.Handle("POST /app/caja/movimiento/{id}/eliminar", app)
	mux.Handle("POST /app/caja/cerrar", app)
	mux.Handle("GET /app/servicios", app)
	mux.Handle("POST /app/servicios", app)
	mux.Handle("POST /app/servicios/{id}", app)
	mux.Handle("DELETE /app/servicios/{id}", app)
	mux.Handle("GET /app/facturacion", app)
	mux.Handle("POST /app/facturas", app)
	mux.Handle("POST /app/facturas/{id}/pagar", app)
	mux.Handle("GET /app/facturas/{id}", app)
	mux.Handle("GET /app/facturas/{id}/pdf", app)
	mux.Handle("GET /app/facturas/{id}/editar", app)
	mux.Handle("POST /app/facturas/{id}", app)
	mux.Handle("GET /app/leads", app)
	mux.Handle("GET /app/configuracion", app)
	mux.Handle("GET /app/configuracion/{section}", app)
	mux.Handle("POST /app/configuracion", app)
	mux.Handle("POST /app/configuracion/facturacion", app)
	mux.Handle("POST /app/configuracion/horario", app)
	mux.Handle("POST /app/configuracion/bloque", app)
	mux.Handle("POST /app/configuracion/bloque/{id}", app)
	mux.Handle("POST /app/configuracion/bloque/{id}/eliminar", app)
	mux.Handle("POST /app/configuracion/festivo", app)
	mux.Handle("POST /app/configuracion/festivo/{id}/eliminar", app)
	mux.Handle("POST /app/configuracion/equipo/{id}", app)

	admin := h.basicAuth(http.HandlerFunc(h.Leads))
	mux.Handle("GET /admin/leads", admin)

	mux.HandleFunc("GET /{sector}", h.Sector)
	mux.HandleFunc("GET /{sector}/{specialty}", h.Specialty)
	mux.HandleFunc("GET /", h.NotFound)

	staticHandler := cacheControl(http.StripPrefix("/static/", http.FileServer(http.FS(h.staticFS))))
	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/static/") {
			staticHandler.ServeHTTP(w, r)
			return
		}
		mux.ServeHTTP(w, r)
	})

	return h.logging(h.security(root))
}

func (h *Handler) baseData(ctx context.Context, path string) (render.PageData, error) {
	site, err := h.store.Settings(ctx)
	if err != nil {
		return render.PageData{}, err
	}
	sectors, err := h.store.Sectors(ctx)
	if err != nil {
		return render.PageData{}, err
	}
	solutions, err := h.store.Solutions(ctx)
	if err != nil {
		return render.PageData{}, err
	}
	return render.PageData{
		Site:             site,
		Sectors:          sectors,
		Solutions:        solutions,
		Path:             path,
		Year:             time.Now().Year(),
		TurnstileSiteKey: h.cfg.TurnstileSiteKey,
	}, nil
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	d, err := h.baseData(r.Context(), "/")
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	if d.Challenges, err = h.store.Challenges(r.Context()); err != nil {
		h.internalError(w, r, err)
		return
	}
	if d.Stats, err = h.store.Stats(r.Context()); err != nil {
		h.internalError(w, r, err)
		return
	}
	if d.Testimonials, err = h.store.Testimonials(r.Context()); err != nil {
		h.internalError(w, r, err)
		return
	}
	d.Title = "Software para clínicas, belleza y salud | " + d.Site.SiteName
	d.Description = "Agenda online, historia clínica, marketing automatizado y facturación en una sola plataforma para clínicas, centros de belleza y salud."
	h.renderPage(w, r, "home", d)
}

func (h *Handler) Pricing(w http.ResponseWriter, r *http.Request) {
	d, err := h.baseData(r.Context(), "/precios")
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	if d.Plans, err = h.store.Plans(r.Context()); err != nil {
		h.internalError(w, r, err)
		return
	}
	if d.FAQs, err = h.store.FAQs(r.Context()); err != nil {
		h.internalError(w, r, err)
		return
	}
	d.Title = "Precios | " + d.Site.SiteName
	d.Description = "La aplicación tal y como está es gratis. Desarrollo a medida desde 45 €, alojamiento 10 €/mes y servicio técnico 5 €/mes."
	h.renderPage(w, r, "pricing", d)
}

func (h *Handler) FAQ(w http.ResponseWriter, r *http.Request) {
	d, err := h.baseData(r.Context(), "/preguntas-frecuentes")
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	if d.FAQs, err = h.store.FAQs(r.Context()); err != nil {
		h.internalError(w, r, err)
		return
	}
	d.Title = "Preguntas frecuentes | " + d.Site.SiteName
	d.Description = "Resolvemos las dudas más comunes sobre el ecosistema Klowna."
	h.renderPage(w, r, "faq", d)
}

func (h *Handler) Contact(w http.ResponseWriter, r *http.Request) {
	d, err := h.baseData(r.Context(), "/contacto")
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	d.FormOK = r.URL.Query().Get("ok") == "1"
	d.FormError = r.URL.Query().Get("error")
	d.Title = "Contacto y demo | " + d.Site.SiteName
	d.Description = "Habla con un especialista y descubre cómo Klowna puede ayudar a tu clínica."
	h.renderPage(w, r, "contact", d)
}

func (h *Handler) Sector(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("sector")
	sector, err := h.store.SectorBySlug(r.Context(), slug)
	if store.IsNotFound(err) {
		h.NotFound(w, r)
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	d, err := h.baseData(r.Context(), "/"+slug)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	d.Sector = &sector
	d.Specialties = sector.Specialties
	d.Title = sector.Name + " | " + d.Site.SiteName
	d.Description = sector.Tagline
	if d.Stats, err = h.store.Stats(r.Context()); err != nil {
		h.internalError(w, r, err)
		return
	}
	if d.Challenges, err = h.store.Challenges(r.Context()); err != nil {
		h.internalError(w, r, err)
		return
	}
	h.renderPage(w, r, "sector", d)
}

func (h *Handler) Specialty(w http.ResponseWriter, r *http.Request) {
	sectorSlug := r.PathValue("sector")
	specialtySlug := r.PathValue("specialty")

	sector, err := h.store.SectorBySlug(r.Context(), sectorSlug)
	if store.IsNotFound(err) {
		h.NotFound(w, r)
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	specialty, err := h.store.SpecialtyBySlug(r.Context(), sectorSlug, specialtySlug)
	if store.IsNotFound(err) {
		h.NotFound(w, r)
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	d, err := h.baseData(r.Context(), "/"+sectorSlug+"/"+specialtySlug)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	d.Sector = &sector
	d.Specialty = &specialty
	d.Specialties = sector.Specialties
	d.Title = specialty.Name + " | " + d.Site.SiteName
	d.Description = specialty.Description
	h.renderPage(w, r, "specialty", d)
}

func (h *Handler) Leads(w http.ResponseWriter, r *http.Request) {
	d, err := h.baseData(r.Context(), "/admin/leads")
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	d.Leads, err = h.store.Leads(r.Context(), 200)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	d.Title = "Leads | " + d.Site.SiteName
	d.Description = "Panel interno de solicitudes."
	h.renderPage(w, r, "leads", d)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := h.store.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "unavailable"})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) NotFound(w http.ResponseWriter, r *http.Request) {
	d, err := h.baseData(r.Context(), r.URL.Path)
	if err != nil {
		http.Error(w, "No encontrado", http.StatusNotFound)
		return
	}
	d.Title = "Página no encontrada | " + d.Site.SiteName
	d.Description = "La página que buscas no existe."
	h.renderPageStatus(w, r, "notfound", d, http.StatusNotFound)
}

func (h *Handler) CreateLead(w http.ResponseWriter, r *http.Request) {
	jsonRequest := strings.Contains(r.Header.Get("Content-Type"), "application/json")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var lead model.Lead
	if jsonRequest {
		if err := json.NewDecoder(r.Body).Decode(&lead); err != nil {
			h.respondLead(w, r, jsonRequest, http.StatusBadRequest, "Datos no válidos.")
			return
		}
		lead.Source = "api"
	} else {
		if err := r.ParseForm(); err != nil {
			h.respondLead(w, r, jsonRequest, http.StatusBadRequest, "Datos no válidos.")
			return
		}
		lead = model.Lead{
			Name:         r.FormValue("name"),
			Email:        r.FormValue("email"),
			Phone:        r.FormValue("phone"),
			Clinic:       r.FormValue("clinic"),
			Sector:       r.FormValue("sector"),
			Message:      r.FormValue("message"),
			Consent:      r.FormValue("consent") == "on" || r.FormValue("consent") == "true" || r.FormValue("consent") == "1",
			Website:      r.FormValue("website"),
			Source:       "web",
			CaptchaToken: r.FormValue("cf-turnstile-response"),
		}
	}

	if strings.TrimSpace(lead.Website) != "" {
		h.respondLead(w, r, jsonRequest, http.StatusCreated, "")
		return
	}

	lead.Name = trim(lead.Name, 120)
	lead.Email = trim(lead.Email, 160)
	lead.Phone = trim(lead.Phone, 40)
	lead.Clinic = trim(lead.Clinic, 160)
	lead.Sector = trim(lead.Sector, 120)
	lead.Message = trim(lead.Message, 2000)

	if msg := validateLead(lead); msg != "" {
		h.respondLead(w, r, jsonRequest, http.StatusUnprocessableEntity, msg)
		return
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	lead.IP = host
	lead.UserAgent = trim(r.UserAgent(), 255)

	if err := h.captcha.Verify(r.Context(), lead.CaptchaToken, host); err != nil {
		slog.Warn("lead captcha rejected", "err", err, "ip", lead.IP)
		h.respondLead(w, r, jsonRequest, http.StatusForbidden, "No hemos podido verificar que no eres un robot. Recarga la página e inténtalo de nuevo.")
		return
	}

	if _, err := h.store.CreateLead(r.Context(), lead); err != nil {
		slog.Error("create lead", "err", err)
		h.respondLead(w, r, jsonRequest, http.StatusInternalServerError, "No hemos podido guardar tu solicitud. Inténtalo de nuevo.")
		return
	}

	if err := h.mailer.SendLead(r.Context(), lead); err != nil {
		slog.Error("send lead email", "err", err, "lead", lead.Email)
	}
	h.respondLead(w, r, jsonRequest, http.StatusCreated, "")
}

func validateLead(l model.Lead) string {
	if len([]rune(l.Name)) < 2 {
		return "Escribe tu nombre."
	}
	if l.Email == "" && l.Phone == "" {
		return "Déjanos un email o un teléfono de contacto."
	}
	if l.Email != "" && !emailRE.MatchString(l.Email) {
		return "El email no tiene un formato válido."
	}
	if !l.Consent {
		return "Acepta la política de privacidad para continuar."
	}
	return ""
}

func (h *Handler) respondLead(w http.ResponseWriter, r *http.Request, jsonRequest bool, status int, errMsg string) {
	if jsonRequest {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		if errMsg != "" {
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": errMsg})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
		return
	}
	if errMsg != "" {
		http.Redirect(w, r, "/contacto?error="+url.QueryEscape(errMsg), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/contacto?ok=1", http.StatusSeeOther)
}

func (h *Handler) renderPage(w http.ResponseWriter, r *http.Request, page string, d render.PageData) {
	h.renderPageStatus(w, r, page, d, http.StatusOK)
}

func (h *Handler) renderPageStatus(w http.ResponseWriter, r *http.Request, page string, d render.PageData, status int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := h.render.Render(w, page, d); err != nil {
		slog.Error("render page", "page", page, "path", r.URL.Path, "err", err)
	}
}

func (h *Handler) internalError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("request failed", "path", r.URL.Path, "err", err)
	http.Error(w, "Se ha producido un error. Inténtalo de nuevo.", http.StatusInternalServerError)
}

func (h *Handler) basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		validUser := subtle.ConstantTimeCompare([]byte(user), []byte(h.cfg.AdminUser)) == 1
		validPass := subtle.ConstantTimeCompare([]byte(pass), []byte(h.cfg.AdminPassword)) == 1
		if !ok || !validUser || !validPass {
			w.Header().Set("WWW-Authenticate", `Basic realm="Klowna admin", charset="UTF-8"`)
			http.Error(w, "No autorizado", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) security(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "geolocation=(), camera=(), microphone=()")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self' https://challenges.cloudflare.com; connect-src 'self' https://challenges.cloudflare.com; frame-src https://challenges.cloudflare.com; font-src 'self'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

func (h *Handler) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(sw, r)
		status := sw.status
		if status == 0 {
			status = http.StatusOK
		}
		slog.Info("http", "method", r.Method, "path", r.URL.Path, "status", status, "duration", time.Since(start).String())
	})
}

func cacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		next.ServeHTTP(w, r)
	})
}

func trim(s string, max int) string {
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if len(runes) > max {
		return string(runes[:max])
	}
	return s
}
