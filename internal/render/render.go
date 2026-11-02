package render

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"strings"
	"time"

	"klowna/internal/model"
)

type Renderer struct {
	pages map[string]*template.Template
	auth  map[string]*template.Template
	app   map[string]*template.Template
	frags map[string]*template.Template
}

func New(fsys fs.FS) (*Renderer, error) {
	templatesFS, err := fs.Sub(fsys, "templates")
	if err != nil {
		return nil, fmt.Errorf("templates fs: %w", err)
	}

	r := &Renderer{
		pages: make(map[string]*template.Template),
		auth:  make(map[string]*template.Template),
		app:   make(map[string]*template.Template),
		frags: make(map[string]*template.Template),
	}

	if err := r.parseSet(templatesFS, "pages/*.html", "base.html", []string{"partials/*.html"}, r.pages); err != nil {
		return nil, err
	}
	if err := r.parseSet(templatesFS, "auth/*.html", "auth.html", nil, r.auth); err != nil {
		return nil, err
	}
	if err := r.parseSet(templatesFS, "app/*.html", "app.html", []string{"partials/*.html", "partials/app/*.html"}, r.app); err != nil {
		return nil, err
	}
	if err := r.parseFragments(templatesFS); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Renderer) parseSet(templatesFS fs.FS, glob, layout string, shared []string, dest map[string]*template.Template) error {
	files, err := fs.Glob(templatesFS, glob)
	if err != nil {
		return fmt.Errorf("list templates %s: %w", glob, err)
	}
	for _, page := range files {
		name := strings.TrimSuffix(strings.TrimPrefix(page, strings.Split(glob, "*")[0]), ".html")
		patterns := append([]string{layout}, shared...)
		patterns = append(patterns, page)
		t, err := template.New(layout).Funcs(funcMap()).ParseFS(templatesFS, patterns...)
		if err != nil {
			return fmt.Errorf("parse template %s: %w", page, err)
		}
		dest[name] = t
	}
	return nil
}

func (r *Renderer) parseFragments(templatesFS fs.FS) error {
	files, err := fs.Glob(templatesFS, "fragments/*.html")
	if err != nil {
		return fmt.Errorf("list fragments: %w", err)
	}
	for _, f := range files {
		name := strings.TrimSuffix(strings.TrimPrefix(f, "fragments/"), ".html")
		t, err := template.New(name).Funcs(funcMap()).ParseFS(templatesFS,
			"partials/*.html", "partials/app/*.html", f)
		if err != nil {
			return fmt.Errorf("parse fragment %s: %w", f, err)
		}
		r.frags[name] = t
	}
	return nil
}

func (r *Renderer) Render(w io.Writer, page string, data PageData) error {
	return r.execute(w, r.pages, "base", page, data)
}

func (r *Renderer) RenderAuth(w io.Writer, page string, data PageData) error {
	return r.execute(w, r.auth, "auth", page, data)
}

func (r *Renderer) RenderApp(w io.Writer, page string, data PageData) error {
	return r.execute(w, r.app, "app", page, data)
}

func (r *Renderer) RenderFragment(w io.Writer, name string, data PageData) error {
	t, ok := r.frags[name]
	if !ok {
		return fmt.Errorf("unknown fragment %q", name)
	}
	return t.ExecuteTemplate(w, name, data)
}

func (r *Renderer) execute(w io.Writer, set map[string]*template.Template, layout, page string, data PageData) error {
	t, ok := set[page]
	if !ok {
		return fmt.Errorf("unknown page template %q", page)
	}
	return t.ExecuteTemplate(w, layout, data)
}

func funcMap() template.FuncMap {
	return template.FuncMap{
		"icon": func(name string) template.HTML {
			if digit, ok := NumberIconDigit(name); ok {
				return template.HTML(`<svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><text x="12" y="12" text-anchor="middle" dominant-baseline="central" font-family="ui-sans-serif, system-ui, sans-serif" font-size="17" font-weight="700">` + digit + `</text></svg>`)
			}
			path, ok := iconPaths[name]
			if !ok {
				return ""
			}
			return template.HTML(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">` + path + `</svg>`)
		},
		"solidIcon": func(name string) template.HTML {
			if digit, ok := NumberIconDigit(name); ok {
				return template.HTML(`<svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><text x="12" y="12" text-anchor="middle" dominant-baseline="central" font-family="ui-sans-serif, system-ui, sans-serif" font-size="17" font-weight="700">` + digit + `</text></svg>`)
			}
			content, ok := solidIconPaths[name]
			if !ok {
				content = solidIconPaths["sparkles"]
			}
			return template.HTML(`<svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">` + content + `</svg>`)
		},
		"safeHTML":  func(s string) template.HTML { return template.HTML(s) },
		"year":      func() int { return time.Now().Year() },
		"fecha":     formatDate,
		"hora":      formatTime,
		"money":     formatMoney,
		"initials":  initials,
		"shortName": shortName,
		"tel":       telHref,
		"join":      func(s []string, sep string) string { return strings.Join(s, sep) },
		"svcNames": func(s []model.AppointmentService) string {
			names := make([]string, 0, len(s))
			for _, it := range s {
				if it.Name != "" {
					names = append(names, it.Name)
				}
			}
			return strings.Join(names, ", ")
		},
		"svcDuration": func(s []model.AppointmentService) int {
			total := 0
			for _, it := range s {
				total += it.DurationMinutes
			}
			return total
		},
		"statusLabel":         model.AppointmentStatusLabel,
		"paymentLabel":        model.PaymentMethodLabel,
		"historyLabel":        model.PatientHistoryTypeLabel,
		"appointmentStatuses": func() []model.LabelOption { return model.AppointmentStatuses },
		"paymentMethods":      func() []model.LabelOption { return model.PaymentMethods },
		"historyTypes":        func() []model.LabelOption { return model.PatientHistoryTypes },
		"net":                 func(a, b float64) float64 { return float64(int64((a-b)*100+0.5)) / 100 },
		"mul":                 func(a, b float64) float64 { return float64(int64(a*b*100+0.5)) / 100 },
		"serviceIcons":        func() []string { return ServiceIconNames },
		"weekdayName": func(d int) string {
			names := [...]string{"domingo", "lunes", "martes", "miércoles", "jueves", "viernes", "sábado"}
			if d < 0 || d > 6 {
				return ""
			}
			return names[d]
		},
		"svcColor": func(c string) string {
			if c == "" {
				return "#75162d"
			}
			return c
		},
		"svcIcon": func(i string) string {
			if i == "" {
				return "sparkles"
			}
			return i
		},
		"firstSvcColor": func(s []model.AppointmentService) string {
			if len(s) == 0 || s[0].Color == "" {
				return "#75162d"
			}
			return s[0].Color
		},
		"fechaCorta": func(s string) string {
			t, err := time.Parse("2006-01-02", s)
			if err != nil {
				return s
			}
			return t.Format("02/01/2006")
		},
		"staffColor": func(c string) string {
			if c == "" {
				return "#75162d"
			}
			return c
		},
		"staffIcon": func(i string) string {
			if i == "" {
				return "user"
			}
			return i
		},
		"slotTop":       slotTop,
		"slotHeight":    slotHeight,
		"hourSlots":     hourSlots,
		"agendaSlots":   agendaSlots,
		"closedMinutes": closedMinutes,
		"iso":           func(t time.Time) string { return t.UTC().Format(time.RFC3339) },
		"assigneeID": func(id *int64) int64 {
			if id == nil {
				return 0
			}
			return *id
		},
		"json": func(v any) string {
			b, err := json.Marshal(v)
			if err != nil {
				return ""
			}
			return string(b)
		},
		"themeCSS": func(t model.Theme) template.CSS {
			var b strings.Builder
			b.WriteString(":root{")
			for _, v := range t.VarsWithPrefix() {
				b.WriteString(v.Name)
				b.WriteString(":")
				b.WriteString(v.Value)
				b.WriteString(";")
			}
			b.WriteString("}")
			return template.CSS(b.String())
		},
	}
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	months := [...]string{
		"enero", "febrero", "marzo", "abril", "mayo", "junio",
		"julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre",
	}
	return fmt.Sprintf("%d de %s de %d", t.Day(), months[int(t.Month())-1], t.Year())
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("15:04")
}

// Franja horaria de la vista de día (08:00–20:00).
const (
	DayStartHour = 8
	DayEndHour   = 20
)

const dayMinutes = (DayEndHour - DayStartHour) * 60

func slotTop(t time.Time) float64 {
	minutes := (t.Hour()-DayStartHour)*60 + t.Minute()
	if minutes < 0 {
		minutes = 0
	}
	if minutes > dayMinutes {
		minutes = dayMinutes
	}
	return float64(minutes) / dayMinutes * 100
}

func slotHeight(start, end time.Time) float64 {
	minutes := int(end.Sub(start).Minutes())
	if minutes < 15 {
		minutes = 15
	}
	if minutes > dayMinutes {
		minutes = dayMinutes
	}
	return float64(minutes) / dayMinutes * 100
}

// hourSlots devuelve las franjas de 15 minutos de una columna (08:00–20:00).
func hourSlots() []HourSlot {
	n := (DayEndHour - DayStartHour) * 4
	h := 100 / float64(n)
	out := make([]HourSlot, 0, n)
	for i := 0; i < n; i++ {
		mins := i * 15
		out = append(out, HourSlot{
			Time:   fmt.Sprintf("%02d:%02d", DayStartHour+mins/60, mins%60),
			Top:    float64(i) * h,
			Height: h,
		})
	}
	return out
}

func slotOffsetMinutes(t time.Time) int {
	minutes := (t.Hour()-DayStartHour)*60 + t.Minute()
	if minutes < 0 {
		return 0
	}
	if minutes > dayMinutes {
		return dayMinutes
	}
	return minutes
}

func rangesOverlap(start, end int, closed []model.ClosedRange) bool {
	for _, c := range closed {
		cs, ce := slotOffsetMinutes(c.Start), slotOffsetMinutes(c.End)
		if ce <= cs {
			continue
		}
		if start < ce && end > cs {
			return true
		}
	}
	return false
}

// agendaSlots devuelve las franjas de 15 minutos que no caen dentro de ningún
// tramo cerrado (horario del centro o bloqueo de un empleado).
func agendaSlots(lists ...[]model.ClosedRange) []HourSlot {
	all := hourSlots()
	closed := make([]model.ClosedRange, 0, 4)
	for _, list := range lists {
		closed = append(closed, list...)
	}
	step := dayMinutes / len(all)
	out := make([]HourSlot, 0, len(all))
	for i, slot := range all {
		if rangesOverlap(i*step, (i+1)*step, closed) {
			continue
		}
		out = append(out, slot)
	}
	return out
}

// closedMinutes serializa los tramos cerrados como "inicio-fin" en minutos
// desde la hora de apertura (08:00), para que el JS bloquee arrastres.
func closedMinutes(lists ...[]model.ClosedRange) string {
	parts := make([]string, 0, 4)
	for _, list := range lists {
		for _, c := range list {
			start, end := slotOffsetMinutes(c.Start), slotOffsetMinutes(c.End)
			if end <= start {
				continue
			}
			parts = append(parts, fmt.Sprintf("%d-%d", start, end))
		}
	}
	return strings.Join(parts, ",")
}

func formatMoney(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	return strings.Replace(s, ".", ",", 1) + " €"
}

// shortName devuelve las iniciales del nombre y del primer apellido.
func shortName(s string) string {
	return initials(s)
}

func initials(s string) string {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return "?"
	}
	out := make([]rune, 0, 2)
	for i, f := range fields {
		if i >= 2 {
			break
		}
		out = append(out, []rune(f)[0])
	}
	return strings.ToUpper(string(out))
}

func telHref(phone string) string {
	replacer := strings.NewReplacer(" ", "", "(", "", ")", "", "-", "", ".", "")
	return "tel:" + replacer.Replace(phone)
}
