package render_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"klowna/internal/model"
	"klowna/internal/render"
	"klowna/web"
)

func samplePageData() render.PageData {
	sector := model.Sector{
		ID: 1, Slug: "salud", Name: "Salud", Icon: "heart",
		Tagline: "Cuida de tus pacientes", Description: "Descripción del sector.",
	}
	specialty := model.Specialty{
		ID: 1, SectorID: 1, SectorSlug: "salud", Slug: "odontologia",
		Name: "Odontología", Description: "Descripción de la especialidad.",
	}
	return render.PageData{
		Title:       "Página de prueba",
		Description: "Descripción de prueba",
		Year:        2026,
		Site: model.Settings{
			SiteName: "Klowna", Tagline: "Software de gestión",
			Phone: "+34 919 22 77 24", Email: "hola@klowna.com",
			WhatsApp: "34919227724", LegalNote: "Demo",
		},
		Sectors:      []model.Sector{sector},
		Solutions:    []model.Solution{{Slug: "saas", Name: "Klowna SaaS", Icon: "cloud"}},
		Challenges:   []model.Challenge{{Title: "Reto", Description: "Descripción", Icon: "heart"}},
		Stats:        []model.Stat{{Value: "+1", Label: "Cita"}},
		Testimonials: []model.Testimonial{{Sector: "Salud", Company: "Clínica Demo", Quote: "Muy bien"}},
		Plans:        []model.Plan{{Name: "Pro", Price: "79 €", Period: "/mes", Features: []string{"Agenda", "CRM"}}},
		FAQs:         []model.FAQ{{Question: "¿Pregunta?", Answer: "Respuesta."}},
		Sector:       &sector,
		Specialty:    &specialty,
		Specialties:  []model.Specialty{specialty},
		Leads:        []model.Lead{{ID: 1, Name: "Ana", Email: "ana@test.com", Sector: "Salud", Message: "Hola", CreatedAt: time.Now()}},
	}
}

func TestPagesRender(t *testing.T) {
	renderer, err := render.New(web.Templates)
	if err != nil {
		t.Fatalf("render.New: %v", err)
	}

	pages := []string{"home", "pricing", "contact", "sector", "specialty", "leads", "notfound"}
	for _, page := range pages {
		t.Run(page, func(t *testing.T) {
			var buf bytes.Buffer
			if err := renderer.Render(&buf, page, samplePageData()); err != nil {
				t.Fatalf("render %s: %v", page, err)
			}
			out := buf.String()
			if !strings.Contains(out, "<!DOCTYPE html>") || !strings.Contains(out, "</html>") {
				t.Fatalf("render %s: incomplete document", page)
			}
			if strings.Contains(out, "no page template") {
				t.Fatalf("render %s: unexpected content", page)
			}
		})
	}
}

func TestUnknownPage(t *testing.T) {
	renderer, err := render.New(web.Templates)
	if err != nil {
		t.Fatalf("render.New: %v", err)
	}
	if err := renderer.Render(&bytes.Buffer{}, "missing", samplePageData()); err == nil {
		t.Fatal("expected an error for an unknown page")
	}
}
