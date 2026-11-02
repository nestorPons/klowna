package model

import "time"

type User struct {
	ID           int64
	Email        string
	Name         string
	Color        string
	Icon         string
	Avatar       string
	PasswordHash string
	Role         string
	ManagerID    *int64
}

func (u User) IsManager() bool { return u.Role == "manager" }

// ConnectedUser es un usuario con al menos una sesión activa.
type ConnectedUser struct {
	User
	LastLogin time.Time
	Sessions  int
}

type Service struct {
	ID              int64
	UserID          int64
	Name            string
	DurationMinutes int
	Price           float64
	DepositRequired bool
	DepositAmount   float64
	Color           string
	Icon            string
	SortOrder       int
	CreatedAt       time.Time
}

type AppointmentService struct {
	ServiceID       int64   `json:"service_id"`
	Name            string  `json:"name"`
	DurationMinutes int     `json:"duration_minutes"`
	Price           float64 `json:"price"`
	DepositAmount   float64 `json:"deposit_amount"`
	Color           string  `json:"color"`
	Icon            string  `json:"icon"`
}

type Patient struct {
	ID        int64
	UserID    int64
	Name      string
	Email     string
	Phone     string
	Notes     string
	CreatedAt time.Time
}

type Appointment struct {
	ID            int64
	UserID        int64
	PatientID     int64
	PatientName   string
	AssigneeID    *int64
	AssigneeName  string
	Services      []AppointmentService
	StartsAt      time.Time
	EndsAt        time.Time
	Status        string
	Notes         string
	DepositDue    float64
	DepositPaid   float64
	DepositMethod string
	DepositPaidAt *time.Time
	CreatedAt     time.Time
}

// DepositPending indica que la cita exige una señal que todavía no se ha
// cobrado por completo.
func (a Appointment) DepositPending() bool { return a.DepositRemaining() > 0 }

// DepositRemaining devuelve la parte de la señal que queda por cobrar.
func (a Appointment) DepositRemaining() float64 {
	if a.DepositDue <= 0 || a.DepositPaid >= a.DepositDue {
		return 0
	}
	return roundMoney(a.DepositDue - a.DepositPaid)
}

// DepositLabel resume el estado de la señal para mostrarlo en la interfaz.
func (a Appointment) DepositLabel() string {
	if a.DepositDue <= 0 {
		return ""
	}
	if a.DepositPending() {
		return "Señal pendiente"
	}
	return "Señal cobrada"
}

type Invoice struct {
	ID            int64
	UserID        int64
	PatientID     int64
	PatientName   string
	AppointmentID *int64
	Number        string
	Concept       string
	Subtotal      float64
	VATRate       float64
	VATAmount     float64
	Amount        float64
	Status        string
	IssuedAt      time.Time
	PaidAt        *time.Time
	PaymentMethod string
	CreatedAt     time.Time
	Items         []InvoiceItem
}

type InvoiceItem struct {
	ID          int64
	InvoiceID   int64
	ServiceID   *int64
	Description string
	Quantity    int
	UnitPrice   float64
	VATRate     float64
	LineTotal   float64
}

type BillingSettings struct {
	UserID      int64
	CompanyName string
	TaxID       string
	Address     string
	Phone       string
	Email       string
	VATRate     float64
}

// ScheduleDay representa el horario de un día de la semana (Day usa time.Weekday:
// 0=domingo … 6=sábado). Open/Close vacíos indican día cerrado.
type ScheduleDay struct {
	Day   int    `json:"day"`
	Open  string `json:"open"`
	Close string `json:"close"`
}

// ScheduleBlock es un bloque de horas cerrado de un empleado, recurrente por día
// de la semana.
type ScheduleBlock struct {
	ID        int64
	SeriesID  int64
	UserID    int64
	StaffID   int64
	StaffName string
	Weekday   int
	Start     string
	End       string
	Reason    string
}

// Holiday es una fecha concreta en la que el centro permanece cerrado.
type Holiday struct {
	ID     int64
	UserID int64
	Date   string
	Name   string
}

// ClosedRange delimita un tramo cerrado dentro del día de la agenda.
type ClosedRange struct {
	Start time.Time
	End   time.Time
	Label string
}

// DefaultSchedule devuelve un horario abierto de 08:00 a 20:00 toda la semana.
func DefaultSchedule() []ScheduleDay {
	days := make([]ScheduleDay, 7)
	for i := 0; i < 7; i++ {
		days[i] = ScheduleDay{Day: i, Open: "08:00", Close: "20:00"}
	}
	return days
}

type AppStats struct {
	Patients        int
	AppointmentsDay int
	PendingInvoices int
	PendingAmount   float64
	CollectedAmount float64
	CashExpected    float64
	CashToday       bool
}

// --- Estados de cita ---

type LabelOption struct {
	Value string
	Label string
}

// AppointmentStatuses son los estados por los que pasa una cita, en orden
// operativo: pendiente → confirmada → realizada → no asistida → facturada → abonada.
var AppointmentStatuses = []LabelOption{
	{Value: "pendiente", Label: "Pendiente"},
	{Value: "confirmada", Label: "Confirmada"},
	{Value: "realizada", Label: "Realizada"},
	{Value: "no_asistida", Label: "No asistida"},
	{Value: "facturada", Label: "Facturada"},
	{Value: "abonada", Label: "Abonada"},
}

func AppointmentStatusLabel(value string) string {
	for _, s := range AppointmentStatuses {
		if s.Value == value {
			return s.Label
		}
	}
	return value
}

func IsAppointmentStatus(value string) bool {
	for _, s := range AppointmentStatuses {
		if s.Value == value {
			return true
		}
	}
	return false
}

// --- Formas de pago ---

var PaymentMethods = []LabelOption{
	{Value: "efectivo", Label: "Efectivo"},
	{Value: "tarjeta", Label: "Tarjeta"},
	{Value: "transferencia", Label: "Transferencia bancaria"},
	{Value: "bizum", Label: "Bizum"},
	{Value: "otros", Label: "Otros"},
}

func PaymentMethodLabel(value string) string {
	for _, m := range PaymentMethods {
		if m.Value == value {
			return m.Label
		}
	}
	return value
}

func IsPaymentMethod(value string) bool {
	for _, m := range PaymentMethods {
		if m.Value == value {
			return true
		}
	}
	return false
}

// --- Historial del cliente ---

type PatientHistoryEntry struct {
	ID         int64
	UserID     int64
	PatientID  int64
	AuthorID   *int64
	AuthorName string
	EntryType  string
	Content    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

var PatientHistoryTypes = []LabelOption{
	{Value: "nota", Label: "Nota"},
	{Value: "comentario", Label: "Comentario"},
	{Value: "seguimiento", Label: "Seguimiento"},
	{Value: "incidencia", Label: "Incidencia"},
}

func PatientHistoryTypeLabel(value string) string {
	for _, t := range PatientHistoryTypes {
		if t.Value == value {
			return t.Label
		}
	}
	return value
}

func IsPatientHistoryType(value string) bool {
	for _, t := range PatientHistoryTypes {
		if t.Value == value {
			return true
		}
	}
	return false
}

// PatientServiceDone es un servicio ya prestado a un cliente, extraído de las
// citas realizadas, facturadas o abonadas.
type PatientServiceDone struct {
	AppointmentID   int64
	StartsAt        time.Time
	Status          string
	ServiceName     string
	DurationMinutes int
	Price           float64
}

// --- Caja diaria ---

type CashDenomination struct {
	Key   string  `json:"key"`
	Value float64 `json:"value"`
	Label string  `json:"label"`
	Kind  string  `json:"kind"`
	Count int     `json:"count"`
}

// CashDenominations devuelve los billetes y monedas con los que se cuenta la
// caja. El orden va de mayor a menor valor.
func CashDenominations() []CashDenomination {
	return []CashDenomination{
		{Key: "b500", Value: 500, Label: "500 €", Kind: "billete"},
		{Key: "b200", Value: 200, Label: "200 €", Kind: "billete"},
		{Key: "b100", Value: 100, Label: "100 €", Kind: "billete"},
		{Key: "b50", Value: 50, Label: "50 €", Kind: "billete"},
		{Key: "b20", Value: 20, Label: "20 €", Kind: "billete"},
		{Key: "b10", Value: 10, Label: "10 €", Kind: "billete"},
		{Key: "b5", Value: 5, Label: "5 €", Kind: "billete"},
		{Key: "m2", Value: 2, Label: "2 €", Kind: "moneda"},
		{Key: "m1", Value: 1, Label: "1 €", Kind: "moneda"},
		{Key: "m050", Value: 0.50, Label: "0,50 €", Kind: "moneda"},
		{Key: "m020", Value: 0.20, Label: "0,20 €", Kind: "moneda"},
		{Key: "m010", Value: 0.10, Label: "0,10 €", Kind: "moneda"},
		{Key: "m005", Value: 0.05, Label: "0,05 €", Kind: "moneda"},
		{Key: "m002", Value: 0.02, Label: "0,02 €", Kind: "moneda"},
		{Key: "m001", Value: 0.01, Label: "0,01 €", Kind: "moneda"},
	}
}

func CashTotal(denoms []CashDenomination) float64 {
	total := 0.0
	for _, d := range denoms {
		if d.Count > 0 {
			total += d.Value * float64(d.Count)
		}
	}
	return float64(int64(total*100+0.5)) / 100
}

type CashSession struct {
	ID            int64
	UserID        int64
	Date          string
	OpeningAmount float64
	CountedAmount float64
	ExpectedCash  float64
	Difference    float64
	Denoms        []CashDenomination
	Notes         string
	Status        string
	OpenedAt      time.Time
	ClosedAt      *time.Time
	Income        float64
	Expenses      float64
}

func (s CashSession) IsOpen() bool { return s.Status == "abierta" }

type CashMovement struct {
	ID            int64
	UserID        int64
	SessionID     int64
	MovementType  string
	PaymentMethod string
	Amount        float64
	Concept       string
	InvoiceID     *int64
	AppointmentID *int64
	CreatedAt     time.Time
}

type CashMethodTotal struct {
	Method   string
	Income   float64
	Expenses float64
}

type CashSummary struct {
	Opening       float64
	CashIncome    float64
	OtherIncome   float64
	TotalIncome   float64
	CashExpenses  float64
	OtherExpenses float64
	TotalExpenses float64
	ExpectedCash  float64
	Counted       float64
	Difference    float64
	Methods       []CashMethodTotal
}

// SummarizeCash calcula el resumen diario a partir de los movimientos de caja.
func SummarizeCash(movements []CashMovement, opening, counted float64) CashSummary {
	s := CashSummary{Opening: opening, Counted: counted}
	methodIndex := make(map[string]int, len(PaymentMethods))
	for _, m := range PaymentMethods {
		methodIndex[m.Value] = len(s.Methods)
		s.Methods = append(s.Methods, CashMethodTotal{Method: m.Value})
	}
	for _, mv := range movements {
		idx, ok := methodIndex[mv.PaymentMethod]
		if !ok {
			idx = len(s.Methods)
			methodIndex[mv.PaymentMethod] = idx
			s.Methods = append(s.Methods, CashMethodTotal{Method: mv.PaymentMethod})
		}
		if mv.MovementType == "gasto" {
			s.TotalExpenses += mv.Amount
			s.Methods[idx].Expenses += mv.Amount
			if mv.PaymentMethod == "efectivo" {
				s.CashExpenses += mv.Amount
			} else {
				s.OtherExpenses += mv.Amount
			}
			continue
		}
		s.TotalIncome += mv.Amount
		s.Methods[idx].Income += mv.Amount
		if mv.PaymentMethod == "efectivo" {
			s.CashIncome += mv.Amount
		} else {
			s.OtherIncome += mv.Amount
		}
	}
	s.ExpectedCash = roundMoney(opening + s.CashIncome - s.CashExpenses)
	s.Difference = roundMoney(counted - s.ExpectedCash)
	return s
}

func roundMoney(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}
