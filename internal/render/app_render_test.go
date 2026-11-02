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

func TestAppPagesRender(t *testing.T) {
	renderer, err := render.New(web.Templates)
	if err != nil {
		t.Fatalf("render.New: %v", err)
	}

	now := time.Now()
	appointmentID := int64(10)
	invoice := model.Invoice{
		ID:            1,
		UserID:        1,
		PatientID:     2,
		PatientName:   "Laura García",
		AppointmentID: &appointmentID,
		Number:        "F-2026-0001",
		Concept:       "Consulta láser",
		Subtotal:      100,
		VATRate:       21,
		VATAmount:     21,
		Amount:        121,
		Status:        "pendiente",
		IssuedAt:      now,
		CreatedAt:     now,
		Items: []model.InvoiceItem{{
			Description: "Consulta láser", Quantity: 1, UnitPrice: 100, VATRate: 21, LineTotal: 100,
		}},
	}
	appt := model.Appointment{
		ID:          10,
		UserID:      1,
		PatientID:   2,
		PatientName: "Laura García",
		StartsAt:    now,
		EndsAt:      now.Add(30 * time.Minute),
		Status:      "confirmada",
		DepositDue:  20,
		Services:    []model.AppointmentService{{ServiceID: 1, Name: "Consulta láser", DurationMinutes: 30, Price: 100, DepositAmount: 20}},
	}
	data := samplePageData()
	data.User = &model.User{ID: 1, Name: "Clínica Demo", Email: "demo@klowna.com", Role: "manager"}
	data.ConnectedUsers = []model.ConnectedUser{{
		User:      model.User{ID: 1, Name: "Clínica Demo", Email: "demo@klowna.com", Role: "manager", Color: "#75162d", Icon: "star"},
		LastLogin: now, Sessions: 1,
	}}
	data.Theme = model.DefaultTheme()
	data.ThemePresets = model.ThemePresets()
	data.Billing = model.BillingSettings{
		UserID: 1, CompanyName: "Clínica Demo S.L.", TaxID: "B12345678",
		Address: "Calle Mayor 10", Phone: "+34 910 000 000", Email: "demo@klowna.com", VATRate: 21,
	}
	data.Invoice = &invoice
	data.Invoices = []model.Invoice{invoice}
	data.Patients = []model.Patient{{ID: 2, UserID: 1, Name: "Laura García"}}
	data.Patient = &data.Patients[0]
	data.PatientHistory = []model.PatientHistoryEntry{{
		ID: 1, UserID: 1, PatientID: 2, AuthorName: "Clínica Demo",
		EntryType: "seguimiento", Content: "Revisión mensual.", CreatedAt: now, UpdatedAt: now,
	}}
	data.PatientServices = []model.PatientServiceDone{{
		AppointmentID: 10, StartsAt: now, Status: "realizada", ServiceName: "Consulta láser", DurationMinutes: 30, Price: 100,
	}}
	data.Services = []model.Service{{ID: 1, UserID: 1, Name: "Consulta láser", DurationMinutes: 30, Price: 100, DepositRequired: true, DepositAmount: 20}}
	data.Appointments = []model.Appointment{appt}
	data.AppStats = &model.AppStats{Patients: 1, AppointmentsDay: 1, PendingInvoices: 1, PendingAmount: 121}
	cashSession := model.CashSession{
		ID: 1, UserID: 1, Date: data.AgendaDate, OpeningAmount: 100, Status: "abierta",
		OpenedAt: now,
	}
	cashMovements := []model.CashMovement{{
		ID: 1, UserID: 1, SessionID: 1, MovementType: "ingreso", PaymentMethod: "efectivo",
		Amount: 121, Concept: "Cobro factura F-2026-0001", InvoiceID: &invoice.ID, CreatedAt: now,
	}}
	data.CashSession = &cashSession
	data.CashMovements = cashMovements
	data.CashSummary = model.SummarizeCash(cashMovements, cashSession.OpeningAmount, 0)
	data.CashDate = data.AgendaDate
	data.CashIsToday = true
	data.CashHistory = []model.CashSession{cashSession}
	data.AgendaDate = now.Format("2006-01-02")
	data.AgendaView = "day"
	data.AgendaHours = []string{"08:00", "09:00"}
	data.AgendaColumns = []render.AgendaColumn{{AssigneeID: 0, Name: "Sin asignar", Events: []model.Appointment{appt}}}
	data.AgendaDays = []render.AgendaDay{{Date: data.AgendaDate, Label: "lun 27", Weekday: "lunes", Num: 27, InMonth: true, Appointments: []model.Appointment{appt}}}

	pages := []string{
		"dashboard", "agenda", "invoices", "invoice", "settings-apariencia", "settings-facturacion",
		"settings-horario", "settings-equipo", "settings-bloqueos", "services", "patients", "patient", "leads", "caja",
	}
	for _, page := range pages {
		t.Run(page, func(t *testing.T) {
			var buf bytes.Buffer
			if err := renderer.RenderApp(&buf, page, data); err != nil {
				t.Fatalf("render app %s: %v", page, err)
			}
		})
	}
}

func TestAgendaWeekRenders(t *testing.T) {
	renderer, err := render.New(web.Templates)
	if err != nil {
		t.Fatalf("render.New: %v", err)
	}

	now := time.Now()
	appt := model.Appointment{
		ID:          10,
		UserID:      1,
		PatientID:   2,
		PatientName: "Laura García",
		StartsAt:    now,
		EndsAt:      now.Add(30 * time.Minute),
		Status:      "confirmada",
		Services: []model.AppointmentService{
			{ServiceID: 1, Name: "Consulta láser", DurationMinutes: 30, Price: 100, Color: "#e05d5d", Icon: "zap"},
			{ServiceID: 2, Name: "Masaje", DurationMinutes: 15, Price: 40, Color: "#5b8def", Icon: "droplet"},
		},
	}

	data := samplePageData()
	data.User = &model.User{ID: 1, Name: "Clínica Demo", Email: "demo@klowna.com", Role: "manager"}
	data.AgendaView = "week"
	data.AgendaDate = now.Format("2006-01-02")
	data.AgendaHours = []string{"08:00", "09:00"}
	data.Staff = []model.User{{ID: 2, Name: "Ana Beltrán", Color: "#e05d5d", Icon: "sparkles"}}
	data.AgendaWeekColumns = []render.AgendaWeekColumn{
		{AssigneeID: 0, Name: "Sin asignar"},
		{AssigneeID: 2, Name: "Ana Beltrán", Color: "#e05d5d", Icon: "sparkles"},
	}
	slot := render.AgendaWeekSlot{
		Date: data.AgendaDate, AssigneeID: 2, Name: "Ana Beltrán",
		Color: "#e05d5d", Icon: "sparkles", Appointments: []model.Appointment{appt},
	}
	data.AgendaWeekDays = []render.AgendaWeekDay{{
		Date: data.AgendaDate, Weekday: "lunes", Label: "lun 27", IsToday: true,
		Slots: []render.AgendaWeekSlot{slot},
	}, {
		Date: "2026-12-25", Weekday: "viernes", Label: "vie 25", Closed: true, Holiday: "Navidad",
		Slots: []render.AgendaWeekSlot{{Date: "2026-12-25", AssigneeID: 2, Name: "Ana Beltrán", Color: "#e05d5d", Icon: "num-1"}},
	}}

	var buf bytes.Buffer
	if err := renderer.RenderApp(&buf, "agenda", data); err != nil {
		t.Fatalf("render agenda week: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "data-add-slot") {
		t.Fatalf("la vista de semana no incluye botones para crear cita por hora")
	}
	if !strings.Contains(out, `data-time="08:00"`) {
		t.Fatalf("la vista de semana no incluye la franja horaria 08:00")
	}
	if !strings.Contains(out, `data-closed="1"`) || !strings.Contains(out, "Navidad") {
		t.Fatalf("la vista de semana no marca el festivo como cerrado")
	}
	if !strings.Contains(out, "agenda-seg-week") || !strings.Contains(out, "flex-grow: 30") {
		t.Fatalf("la card de la vista de semana no divide el tiempo por servicio")
	}
	if !strings.Contains(out, "--svc: #e05d5d") || !strings.Contains(out, "--svc: #5b8def") {
		t.Fatalf("la card de la vista de semana no usa el color de cada servicio")
	}
}

func TestAgendaDayClosedHidesSlots(t *testing.T) {
	renderer, err := render.New(web.Templates)
	if err != nil {
		t.Fatalf("render.New: %v", err)
	}

	start := time.Date(2026, 12, 25, 8, 0, 0, 0, time.UTC)
	data := samplePageData()
	data.User = &model.User{ID: 1, Name: "Clínica Demo", Role: "manager"}
	data.AgendaView = "day"
	data.AgendaDate = "2026-12-25"
	data.AgendaClosed = true
	data.AgendaHoliday = "Navidad"
	data.ClinicClosed = []model.ClosedRange{{Start: start, End: start.Add(12 * time.Hour), Label: "Navidad"}}
	data.AgendaColumns = []render.AgendaColumn{{AssigneeID: 0, Name: "Sin asignar"}}

	var buf bytes.Buffer
	if err := renderer.RenderApp(&buf, "agenda", data); err != nil {
		t.Fatalf("render agenda day: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "data-add-slot") {
		t.Fatalf("un día cerrado no debería ofrecer botones para crear citas")
	}
	if !strings.Contains(out, "is-closed") {
		t.Fatalf("el día cerrado debería marcarse con is-closed")
	}
}
