package render

import "klowna/internal/model"

type AgendaColumn struct {
	AssigneeID int64 // 0 = sin asignar
	Name       string
	Icon       string
	Color      string
	Events     []model.Appointment
	Closed     []model.ClosedRange
}

type AgendaWeekColumn struct {
	AssigneeID int64
	Name       string
	Icon       string
	Color      string
}

type AgendaWeekDay struct {
	Date    string
	Weekday string
	Label   string
	IsToday bool
	Closed  bool
	Holiday string
	Slots   []AgendaWeekSlot
}

// AgendaWeekSlot es una celda de la vista de semana: un día × un empleado.
type AgendaWeekSlot struct {
	Date         string
	AssigneeID   int64
	Name         string
	Icon         string
	Color        string
	Closed       []model.ClosedRange
	Appointments []model.Appointment
}

// HourSlot es una franja horaria de una columna de la agenda (día o semana).
type HourSlot struct {
	Time   string
	Top    float64
	Height float64
}

type AgendaDay struct {
	Date         string
	Label        string
	Weekday      string
	Num          int
	IsToday      bool
	InMonth      bool
	Closed       bool
	Holiday      string
	Appointments []model.Appointment
}

type PageData struct {
	Title       string
	Description string
	Path        string
	Year        int

	Site      model.Settings
	Sectors   []model.Sector
	Solutions []model.Solution

	Challenges   []model.Challenge
	Stats        []model.Stat
	Testimonials []model.Testimonial
	Plans        []model.Plan
	FAQs         []model.FAQ

	Sector      *model.Sector
	Specialty   *model.Specialty
	Specialties []model.Specialty
	Leads       []model.Lead

	// Aplicación SaaS
	User           *model.User
	ConnectedUsers []model.ConnectedUser
	Patients       []model.Patient
	Patient        *model.Patient
	Appointments   []model.Appointment
	Invoices       []model.Invoice
	Invoice        *model.Invoice
	AppStats       *model.AppStats
	AgendaDate     string
	PrevDate       string
	NextDate       string
	AgendaView     string
	Query          string

	// Agenda por empleado
	Staff             []model.User
	AgendaColumns     []AgendaColumn
	AgendaDays        []AgendaDay
	AgendaHours       []string
	AgendaWeekColumns []AgendaWeekColumn
	AgendaWeekDays    []AgendaWeekDay
	AgendaStaffFilter int64
	AgendaClosed      bool
	AgendaHoliday     string
	Holidays          []model.Holiday

	// Servicios
	Services []model.Service

	// Historial del cliente
	PatientHistory  []model.PatientHistoryEntry
	PatientServices []model.PatientServiceDone

	// Caja diaria
	CashSession   *model.CashSession
	CashMovements []model.CashMovement
	CashSummary   model.CashSummary
	CashHistory   []model.CashSession
	CashDate      string
	CashPrevDate  string
	CashNextDate  string
	CashIsToday   bool

	// Facturación
	Billing model.BillingSettings

	// Horario de la clínica y bloques cerrados
	ClinicSchedule []model.ScheduleDay
	ScheduleBlocks []model.ScheduleBlock
	ClinicClosed   []model.ClosedRange

	// Tema del panel
	Theme        model.Theme
	ThemePresets []model.ThemePreset

	// Configuración
	SettingsSection string
	SettingsTitle   string

	FormOK    bool
	FormError string

	// TurnstileSiteKey es la clave pública del captcha del formulario.
	TurnstileSiteKey string
}
