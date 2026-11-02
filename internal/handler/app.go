package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"klowna/internal/model"
	"klowna/internal/pdf"
	"klowna/internal/render"
	"klowna/internal/store"
)

const dateLayout = "2006-01-02"

var weekdays = [...]string{"domingo", "lunes", "martes", "miércoles", "jueves", "viernes", "sábado"}
var weekdaysShort = [...]string{"dom", "lun", "mar", "mié", "jue", "vie", "sáb"}

func isHX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func todayStr() string {
	return time.Now().UTC().Format(dateLayout)
}

func dayBounds(dateStr string) (time.Time, time.Time, error) {
	t, err := time.Parse(dateLayout, dateStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	start := t.UTC()
	return start, start.Add(24 * time.Hour), nil
}

func parseDate(s string) (time.Time, bool) {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return time.Time{}, false
	}
	return t.UTC(), true
}

func shiftDate(dateStr string, days int) string {
	t, err := time.Parse(dateLayout, dateStr)
	if err != nil {
		return dateStr
	}
	return t.AddDate(0, 0, days).Format(dateLayout)
}

func startOfWeek(t time.Time) time.Time {
	days := (int(t.Weekday()) + 6) % 7
	d := t.AddDate(0, 0, -days)
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
}

func startOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func (h *Handler) appData(r *http.Request, path string) (render.PageData, error) {
	site, err := h.store.Settings(r.Context())
	if err != nil {
		return render.PageData{}, err
	}
	user, _ := userFromContext(r.Context())
	theme, err := h.store.Preferences(r.Context(), user.ID)
	if err != nil {
		slog.Error("load preferences", "err", err)
		theme = model.DefaultTheme()
	}
	connected, err := h.store.ConnectedUsers(r.Context(), user.ID, time.Now())
	if err != nil {
		slog.Error("load connected users", "err", err)
	}
	return render.PageData{
		Site:           site,
		User:           &user,
		ConnectedUsers: connected,
		Path:           path,
		Year:           time.Now().Year(),
		Theme:          theme,
		Title:          "Panel | " + site.SiteName,
	}, nil
}

func (h *Handler) renderApp(w http.ResponseWriter, page string, d render.PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.render.RenderApp(w, page, d); err != nil {
		slog.Error("render app page", "page", page, "err", err)
	}
}

func (h *Handler) renderFragment(w http.ResponseWriter, name string, d render.PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.render.RenderFragment(w, name, d); err != nil {
		slog.Error("render fragment", "fragment", name, "err", err)
	}
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) managerOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := userFromContext(r.Context())
		if !ok || !user.IsManager() {
			http.Redirect(w, r, "/app/agenda", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

// --- Dashboard ---

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	if !user.IsManager() {
		http.Redirect(w, r, "/app/agenda", http.StatusSeeOther)
		return
	}
	d, err := h.appData(r, "/app")
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	start, end, _ := dayBounds(todayStr())
	stats, err := h.store.AppStats(r.Context(), user.ID, start, end)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	today, err := h.store.AppointmentsRange(r.Context(), user.ID, start, end)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	d.AppStats = &stats
	d.Appointments = today
	d.Title = "Panel | " + d.Site.SiteName
	h.renderApp(w, "dashboard", d)
}

// --- Agenda ---

func (h *Handler) Agenda(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	d, err := h.appData(r, "/app/agenda")
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	view := r.URL.Query().Get("view")
	if view != "week" && view != "month" {
		view = "day"
	}
	date := r.URL.Query().Get("date")
	if _, ok := parseDate(date); !ok {
		date = todayStr()
	}
	h.fillAgenda(r, &d, user, view, date)
	d.Title = "Agenda | " + d.Site.SiteName

	if isHX(r) {
		h.renderFragment(w, "agenda", d)
		return
	}
	h.renderApp(w, "agenda", d)
}

func (h *Handler) fillAgenda(r *http.Request, d *render.PageData, user model.User, view, date string) {
	d.AgendaView = view
	d.AgendaDate = date

	switch view {
	case "week":
		d.PrevDate = shiftDate(date, -7)
		d.NextDate = shiftDate(date, 7)
	case "month":
		d.PrevDate = startOfMonth(timeMust(date)).AddDate(0, -1, 0).Format(dateLayout)
		d.NextDate = startOfMonth(timeMust(date)).AddDate(0, 1, 0).Format(dateLayout)
	default:
		d.PrevDate = shiftDate(date, -1)
		d.NextDate = shiftDate(date, 1)
	}

	anchor := timeMust(date)
	var start, end time.Time
	switch view {
	case "week":
		start = startOfWeek(anchor)
		end = start.AddDate(0, 0, 7)
	case "month":
		start = startOfMonth(anchor)
		end = start.AddDate(0, 1, 0)
	default:
		start = anchor
		end = start.AddDate(0, 0, 1)
	}

	var appts []model.Appointment
	var err error
	if user.IsManager() {
		appts, err = h.store.AppointmentsRange(r.Context(), user.ID, start, end)
	} else {
		appts, err = h.store.StaffAppointmentsRange(r.Context(), user.ID, start, end)
	}
	if err != nil {
		slog.Error("load agenda", "err", err)
		appts = nil
	}

	if user.IsManager() {
		d.Staff, _ = h.store.Staff(r.Context(), user.ID)
		d.Patients, _ = h.store.Patients(r.Context(), user.ID, "")
		d.Services, _ = h.store.Services(r.Context(), user.ID)
		if billing, err := h.store.BillingSettings(r.Context(), user.ID); err == nil {
			d.Billing = billing
		}
		d.AgendaColumns = buildColumns(user, d.Staff, appts)
	} else {
		d.AgendaColumns = buildColumns(user, nil, appts)
	}

	var staffFilter int64
	if user.IsManager() {
		if v, err := strconv.ParseInt(r.FormValue("staff"), 10, 64); err == nil && v > 0 {
			for _, s := range d.Staff {
				if s.ID == v {
					staffFilter = v
					break
				}
			}
		}
	}
	d.AgendaStaffFilter = staffFilter

	ownerID := scheduleOwner(user)
	schedule, err := h.store.ClinicSchedule(r.Context(), ownerID)
	if err != nil {
		schedule = model.DefaultSchedule()
	}
	holidays, _ := h.store.Holidays(r.Context(), ownerID)
	closedWD := closedWeekdays(schedule)
	holMap := holidayMap(holidays)

	if view != "day" {
		d.AgendaDays = buildDays(view, anchor, appts, closedWD, holMap)
	}
	blocks, _ := h.store.ScheduleBlocks(r.Context(), ownerID)
	if view == "week" {
		d.AgendaWeekColumns, d.AgendaWeekDays = buildWeek(user, anchor, d.Staff, appts, closedWD, holMap, staffFilter, blocks)
	}
	if view == "day" {
		holiday := holMap[date]
		d.AgendaHoliday = holiday
		d.AgendaClosed = closedWD[int(anchor.Weekday())] || holiday != ""
		if holiday != "" {
			dayEnd := dayStart(anchor).Add(time.Duration(render.DayEndHour-render.DayStartHour) * time.Hour)
			d.ClinicClosed = []model.ClosedRange{{Start: dayStart(anchor), End: dayEnd, Label: holiday}}
		} else {
			d.ClinicClosed = closedRangesForDay(anchor, schedule)
		}
		for i := range d.AgendaColumns {
			d.AgendaColumns[i].Closed = blocksForColumn(anchor, blocks, d.AgendaColumns[i].AssigneeID)
		}
	}
	d.AgendaHours = buildHours()
}

func scheduleOwner(user model.User) int64 {
	if user.IsManager() {
		return user.ID
	}
	if user.ManagerID != nil {
		return *user.ManagerID
	}
	return user.ID
}

func dayStart(anchor time.Time) time.Time {
	return time.Date(anchor.Year(), anchor.Month(), anchor.Day(), render.DayStartHour, 0, 0, 0, time.UTC)
}

// closedWeekdays devuelve los días de la semana cerrados según el horario del centro.
func closedWeekdays(days []model.ScheduleDay) map[int]bool {
	out := make(map[int]bool, 7)
	for _, d := range days {
		openT, errO := time.Parse("15:04", d.Open)
		closeT, errC := time.Parse("15:04", d.Close)
		if errO != nil || errC != nil || !closeT.After(openT) {
			out[d.Day] = true
		}
	}
	return out
}

// holidayMap indexa los festivos por fecha (YYYY-MM-DD) con su nombre.
func holidayMap(holidays []model.Holiday) map[string]string {
	out := make(map[string]string, len(holidays))
	for _, h := range holidays {
		out[h.Date] = h.Name
	}
	return out
}

func closedRangesForDay(anchor time.Time, days []model.ScheduleDay) []model.ClosedRange {
	wd := int(anchor.Weekday())
	open, close := "", ""
	for _, d := range days {
		if d.Day == wd {
			open, close = d.Open, d.Close
			break
		}
	}
	start := dayStart(anchor)
	end := start.Add(time.Duration(render.DayEndHour-render.DayStartHour) * time.Hour)
	o, errO := time.Parse("15:04", open)
	c, errC := time.Parse("15:04", close)
	if errO != nil || errC != nil || !c.After(o) {
		return []model.ClosedRange{{Start: start, End: end}}
	}
	openT := start.Add(time.Duration(o.Hour()-render.DayStartHour)*time.Hour + time.Duration(o.Minute())*time.Minute)
	closeT := start.Add(time.Duration(c.Hour()-render.DayStartHour)*time.Hour + time.Duration(c.Minute())*time.Minute)
	if openT.Before(start) {
		openT = start
	}
	if closeT.After(end) {
		closeT = end
	}
	out := make([]model.ClosedRange, 0, 2)
	if openT.After(start) {
		out = append(out, model.ClosedRange{Start: start, End: openT})
	}
	if closeT.Before(end) {
		out = append(out, model.ClosedRange{Start: closeT, End: end})
	}
	return out
}

func blocksForColumn(anchor time.Time, blocks []model.ScheduleBlock, assigneeID int64) []model.ClosedRange {
	if assigneeID == 0 {
		return nil
	}
	wd := int(anchor.Weekday())
	start := dayStart(anchor)
	out := make([]model.ClosedRange, 0, len(blocks))
	for _, b := range blocks {
		if b.StaffID != assigneeID || b.Weekday != wd {
			continue
		}
		s, errS := time.Parse("15:04", b.Start)
		e, errE := time.Parse("15:04", b.End)
		if errS != nil || errE != nil || !e.After(s) {
			continue
		}
		blockStart := start.Add(time.Duration(s.Hour()-render.DayStartHour)*time.Hour + time.Duration(s.Minute())*time.Minute)
		blockEnd := start.Add(time.Duration(e.Hour()-render.DayStartHour)*time.Hour + time.Duration(e.Minute())*time.Minute)
		if blockEnd.Before(start.Add(time.Duration(-1)*time.Hour)) || blockStart.After(start.Add(time.Duration(render.DayEndHour-render.DayStartHour)*time.Hour)) {
			continue
		}
		if blockStart.Before(start) {
			blockStart = start
		}
		if blockEnd.After(start.Add(time.Duration(render.DayEndHour-render.DayStartHour) * time.Hour)) {
			blockEnd = start.Add(time.Duration(render.DayEndHour-render.DayStartHour) * time.Hour)
		}
		out = append(out, model.ClosedRange{Start: blockStart, End: blockEnd, Label: b.Reason})
	}
	return out
}

// staffBlockConflict indica si una cita se solapa con un bloque cerrado del
// empleado asignado en ese día de la semana.
func (h *Handler) staffBlockConflict(r *http.Request, user model.User, assigneeID int64, starts, ends time.Time) (string, bool, error) {
	if assigneeID <= 0 {
		return "", false, nil
	}
	blocks, err := h.store.ScheduleBlocks(r.Context(), scheduleOwner(user))
	if err != nil {
		return "", false, err
	}
	wd := int(starts.UTC().Weekday())
	for _, b := range blocks {
		if b.StaffID != assigneeID || b.Weekday != wd {
			continue
		}
		blockStart, okStart := blockDateTime(starts, b.Start)
		blockEnd, okEnd := blockDateTime(starts, b.End)
		if !okStart || !okEnd || !blockEnd.After(blockStart) {
			continue
		}
		if starts.UTC().Before(blockEnd) && ends.UTC().After(blockStart) {
			reason := strings.TrimSpace(b.Reason)
			if reason == "" {
				reason = "horario cerrado"
			}
			return reason, true, nil
		}
	}
	return "", false, nil
}

func blockDateTime(day time.Time, hhmm string) (time.Time, bool) {
	t, err := time.Parse("15:04", hhmm)
	if err != nil {
		return time.Time{}, false
	}
	d := day.UTC()
	return time.Date(d.Year(), d.Month(), d.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC), true
}

func timeMust(s string) time.Time {
	t, _ := time.Parse(dateLayout, s)
	return t.UTC()
}

func buildHours() []string {
	hours := make([]string, 0, render.DayEndHour-render.DayStartHour)
	for i := render.DayStartHour; i < render.DayEndHour; i++ {
		hours = append(hours, fmt.Sprintf("%02d:00", i))
	}
	return hours
}

func buildColumns(user model.User, staff []model.User, appts []model.Appointment) []render.AgendaColumn {
	if !user.IsManager() {
		col := render.AgendaColumn{AssigneeID: user.ID, Name: user.Name, Icon: user.Icon, Color: user.Color}
		for _, a := range appts {
			col.Events = append(col.Events, a)
		}
		return []render.AgendaColumn{col}
	}

	cols := []render.AgendaColumn{{AssigneeID: 0, Name: "Sin asignar"}}
	for _, s := range staff {
		cols = append(cols, render.AgendaColumn{AssigneeID: s.ID, Name: s.Name, Icon: s.Icon, Color: s.Color})
	}
	for _, a := range appts {
		idx := 0
		if a.AssigneeID != nil {
			for i := range cols {
				if cols[i].AssigneeID == *a.AssigneeID {
					idx = i
					break
				}
			}
		}
		cols[idx].Events = append(cols[idx].Events, a)
	}
	return cols
}

func groupByDate(appts []model.Appointment) map[string][]model.Appointment {
	m := make(map[string][]model.Appointment)
	for _, a := range appts {
		key := a.StartsAt.UTC().Format(dateLayout)
		m[key] = append(m[key], a)
	}
	return m
}

func assigneeID(a model.Appointment) int64 {
	if a.AssigneeID == nil {
		return 0
	}
	return *a.AssigneeID
}

// buildWeek construye la vista de semana como una matriz de columnas por día,
// subdividida por empleado, para dibujar un calendario de horas (eje Y) × días (eje X).
func buildWeek(user model.User, anchor time.Time, staff []model.User, appts []model.Appointment, closedWD map[int]bool, holMap map[string]string, staffFilter int64, blocks []model.ScheduleBlock) ([]render.AgendaWeekColumn, []render.AgendaWeekDay) {
	var cols []render.AgendaWeekColumn
	if !user.IsManager() {
		cols = []render.AgendaWeekColumn{{AssigneeID: user.ID, Name: user.Name, Icon: user.Icon, Color: user.Color}}
	} else if staffFilter > 0 {
		for _, s := range staff {
			if s.ID == staffFilter {
				cols = []render.AgendaWeekColumn{{AssigneeID: s.ID, Name: s.Name, Icon: s.Icon, Color: s.Color}}
				break
			}
		}
	}
	if len(cols) == 0 {
		cols = []render.AgendaWeekColumn{{AssigneeID: 0, Name: "Sin asignar"}}
		for _, s := range staff {
			cols = append(cols, render.AgendaWeekColumn{AssigneeID: s.ID, Name: s.Name, Icon: s.Icon, Color: s.Color})
		}
	}

	type key struct {
		date string
		asg  int64
	}
	byCell := make(map[key][]model.Appointment)
	for _, a := range appts {
		k := key{date: a.StartsAt.UTC().Format(dateLayout), asg: assigneeID(a)}
		byCell[k] = append(byCell[k], a)
	}

	start := startOfWeek(anchor)
	today := todayStr()
	days := make([]render.AgendaWeekDay, 0, 7)
	for i := 0; i < 7; i++ {
		day := start.AddDate(0, 0, i)
		dateStr := day.Format(dateLayout)
		if closedWD[int(day.Weekday())] {
			continue
		}
		holiday := holMap[dateStr]
		slots := make([]render.AgendaWeekSlot, 0, len(cols))
		for _, c := range cols {
			slots = append(slots, render.AgendaWeekSlot{
				Date:         dateStr,
				AssigneeID:   c.AssigneeID,
				Name:         c.Name,
				Icon:         c.Icon,
				Color:        c.Color,
				Closed:       blocksForColumn(day, blocks, c.AssigneeID),
				Appointments: byCell[key{date: dateStr, asg: c.AssigneeID}],
			})
		}
		days = append(days, render.AgendaWeekDay{
			Date:    dateStr,
			Weekday: weekdays[day.Weekday()],
			Label:   fmt.Sprintf("%s %d", weekdaysShort[day.Weekday()], day.Day()),
			IsToday: dateStr == today,
			Closed:  holiday != "",
			Holiday: holiday,
			Slots:   slots,
		})
	}
	return cols, days
}

func buildDays(view string, anchor time.Time, appts []model.Appointment, closedWD map[int]bool, holMap map[string]string) []render.AgendaDay {
	byDate := groupByDate(appts)
	var start time.Time
	count := 0
	switch view {
	case "week":
		start = startOfWeek(anchor)
		count = 7
	case "month":
		start = startOfWeek(startOfMonth(anchor))
		count = 42
	}
	today := todayStr()
	days := make([]render.AgendaDay, 0, count)
	for i := 0; i < count; i++ {
		day := start.AddDate(0, 0, i)
		key := day.Format(dateLayout)
		holiday := holMap[key]
		days = append(days, render.AgendaDay{
			Date:         key,
			Label:        fmt.Sprintf("%s %d", weekdaysShort[day.Weekday()], day.Day()),
			Weekday:      weekdays[day.Weekday()],
			Num:          day.Day(),
			IsToday:      key == today,
			InMonth:      view != "month" || day.Month() == anchor.Month(),
			Closed:       closedWD[int(day.Weekday())] || holiday != "",
			Holiday:      holiday,
			Appointments: byDate[key],
		})
	}
	return days
}

// --- Crear cita ---

func (h *Handler) CreateAppointment(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		h.internalError(w, r, err)
		return
	}

	date := r.FormValue("date")
	dateStart, _, err := dayBounds(date)
	if err != nil {
		date = todayStr()
		dateStart, _, _ = dayBounds(date)
	}
	view := r.FormValue("view")
	if view != "week" && view != "month" {
		view = "day"
	}

	patientID, err := strconv.ParseInt(r.FormValue("patient_id"), 10, 64)
	if err != nil || patientID <= 0 {
		h.agendaError(w, r, user, view, date, "Selecciona un cliente.")
		return
	}
	serviceIDs := parseIDs(r.Form["service_id"])
	items, err := h.appointmentServices(r.Context(), user.ID, serviceIDs)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	if len(items) == 0 {
		h.agendaError(w, r, user, view, date, "Selecciona al menos un servicio.")
		return
	}
	startTime, okTime := normalizeClock(r.FormValue("time"))
	if !okTime {
		h.agendaError(w, r, user, view, date, "Hora no válida.")
		return
	}
	tod, err := time.Parse("15:04", startTime)
	if err != nil {
		h.agendaError(w, r, user, view, date, "Hora no válida.")
		return
	}
	duration := serviceTotal(items)

	starts := dateStart.Add(time.Duration(tod.Hour())*time.Hour + time.Duration(tod.Minute())*time.Minute)
	status := trim(r.FormValue("status"), 32)
	if !model.IsAppointmentStatus(status) {
		status = "pendiente"
	}
	appt := model.Appointment{
		UserID:     user.ID,
		PatientID:  patientID,
		Services:   items,
		StartsAt:   starts,
		EndsAt:     starts.Add(time.Duration(duration) * time.Minute),
		Status:     status,
		Notes:      trim(r.FormValue("notes"), 500),
		DepositDue: depositDue(items),
	}
	if v := r.FormValue("assignee_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil && id > 0 {
			appt.AssigneeID = &id
		}
	}

	assignee := int64(0)
	if appt.AssigneeID != nil {
		assignee = *appt.AssigneeID
	}
	if reason, blocked, err := h.staffBlockConflict(r, user, assignee, appt.StartsAt, appt.EndsAt); err != nil {
		h.internalError(w, r, err)
		return
	} else if blocked {
		h.agendaError(w, r, user, view, date, "No se puede asignar la cita: el empleado tiene ese horario cerrado ("+reason+").")
		return
	}

	id, err := h.store.CreateAppointment(r.Context(), appt)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	appt.ID = id
	if depositRequested(r.FormValue("deposit_paid")) {
		if _, err := h.registerDeposit(r.Context(), user.ID, id, r.FormValue("deposit_method")); err != nil {
			h.internalError(w, r, err)
			return
		}
	}
	if _, err := h.settleAppointment(r.Context(), user.ID, appt, status, r.FormValue("payment_method")); err != nil {
		h.internalError(w, r, err)
		return
	}

	h.refreshAgenda(w, r, user, r.FormValue("view"), date)
}

// --- Editar / mover / eliminar cita ---

type appointmentUpdate struct {
	AssigneeID    *int64    `json:"assignee_id"`
	StartsAt      time.Time `json:"starts_at"`
	Status        string    `json:"status"`
	Notes         string    `json:"notes"`
	Services      []int64   `json:"services"`
	PaymentMethod string    `json:"payment_method"`
	DepositPaid   bool      `json:"deposit_paid"`
	DepositMethod string    `json:"deposit_method"`
}

type appointmentMove struct {
	AssigneeID *int64    `json:"assignee_id"`
	StartsAt   time.Time `json:"starts_at"`
	EndsAt     time.Time `json:"ends_at"`
}

func (h *Handler) UpdateAppointment(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "Cita no válida."})
		return
	}
	var req appointmentUpdate
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "Datos no válidos."})
		return
	}

	items, err := h.appointmentServices(r.Context(), user.ID, req.Services)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	if len(items) == 0 {
		// Sin servicios: se elimina la cita, salvo que ya esté cobrada.
		if current, err := h.store.AppointmentByID(r.Context(), user.ID, id); err == nil && current.Status == "abonada" {
			respondJSON(w, http.StatusConflict, map[string]any{
				"ok":    false,
				"error": "La cita está abonada y su cobro consta en caja. No se puede eliminar.",
			})
			return
		}
		if err := h.store.DeleteAppointment(r.Context(), id, user.ID); err != nil {
			respondJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "No se pudo eliminar la cita."})
			return
		}
		respondJSON(w, http.StatusOK, map[string]any{"ok": true, "deleted": true})
		return
	}

	duration := serviceTotal(items)
	status := trim(req.Status, 32)
	if !model.IsAppointmentStatus(status) {
		status = "pendiente"
	}
	current, err := h.store.AppointmentByID(r.Context(), user.ID, id)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "No se pudo actualizar la cita."})
		return
	}
	if current.Status == "abonada" && status != "abonada" {
		respondJSON(w, http.StatusConflict, map[string]any{
			"ok":    false,
			"error": "La cita ya está abonada y su cobro consta en caja. No se puede cambiar el estado.",
		})
		return
	}
	a := model.Appointment{
		ID:         id,
		UserID:     user.ID,
		AssigneeID: req.AssigneeID,
		StartsAt:   req.StartsAt,
		EndsAt:     req.StartsAt.Add(time.Duration(duration) * time.Minute),
		Status:     status,
		Notes:      trim(req.Notes, 500),
		Services:   items,
		DepositDue: depositDue(items),
	}
	assignee := int64(0)
	if a.AssigneeID != nil {
		assignee = *a.AssigneeID
	}
	if reason, blocked, err := h.staffBlockConflict(r, user, assignee, a.StartsAt, a.EndsAt); err != nil {
		h.internalError(w, r, err)
		return
	} else if blocked {
		respondJSON(w, http.StatusConflict, map[string]any{
			"ok":    false,
			"error": "No se puede asignar la cita: el empleado tiene ese horario cerrado (" + reason + ").",
		})
		return
	}
	if err := h.store.UpdateAppointment(r.Context(), a); err != nil {
		respondJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "No se pudo actualizar la cita."})
		return
	}
	if req.DepositPaid {
		if _, err := h.registerDeposit(r.Context(), user.ID, id, req.DepositMethod); err != nil {
			h.internalError(w, r, err)
			return
		}
	}
	invoiceID, err := h.settleAppointment(r.Context(), user.ID, a, status, req.PaymentMethod)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"ok": true, "invoice_id": invoiceID})
}

func (h *Handler) MoveAppointment(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]any{"ok": false})
		return
	}
	var req appointmentMove
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]any{"ok": false})
		return
	}
	assignee := int64(0)
	if req.AssigneeID != nil {
		assignee = *req.AssigneeID
	}
	if reason, blocked, err := h.staffBlockConflict(r, user, assignee, req.StartsAt, req.EndsAt); err != nil {
		h.internalError(w, r, err)
		return
	} else if blocked {
		respondJSON(w, http.StatusConflict, map[string]any{
			"ok":    false,
			"error": "No se puede mover la cita: el empleado tiene ese horario cerrado (" + reason + ").",
		})
		return
	}
	if err := h.store.MoveAppointment(r.Context(), id, user.ID, req.AssigneeID, req.StartsAt, req.EndsAt); err != nil {
		respondJSON(w, http.StatusNotFound, map[string]any{"ok": false})
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) DeleteAppointment(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]any{"ok": false})
		return
	}
	if current, err := h.store.AppointmentByID(r.Context(), user.ID, id); err == nil && current.Status == "abonada" {
		respondJSON(w, http.StatusConflict, map[string]any{
			"ok":    false,
			"error": "La cita está abonada y su cobro consta en caja. No se puede eliminar.",
		})
		return
	}
	if err := h.store.DeleteAppointment(r.Context(), id, user.ID); err != nil {
		respondJSON(w, http.StatusNotFound, map[string]any{"ok": false})
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func parseIDs(in []string) []int64 {
	out := make([]int64, 0, len(in))
	for _, s := range in {
		if id, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64); err == nil && id > 0 {
			out = append(out, id)
		}
	}
	return out
}

func (h *Handler) appointmentServices(ctx context.Context, userID int64, ids []int64) ([]model.AppointmentService, error) {
	services, err := h.store.Services(ctx, userID)
	if err != nil {
		return nil, err
	}
	byID := make(map[int64]model.Service, len(services))
	for _, s := range services {
		byID[s.ID] = s
	}
	seen := make(map[int64]bool, len(ids))
	items := make([]model.AppointmentService, 0, len(ids))
	for _, id := range ids {
		if seen[id] {
			continue
		}
		if s, ok := byID[id]; ok {
			seen[id] = true
			items = append(items, model.AppointmentService{
				ServiceID: s.ID, Name: s.Name, DurationMinutes: s.DurationMinutes,
				Price: s.Price, DepositAmount: serviceDeposit(s), Color: s.Color, Icon: s.Icon,
			})
		}
	}
	return items, nil
}

func serviceTotal(items []model.AppointmentService) int {
	total := 0
	for _, it := range items {
		total += it.DurationMinutes
	}
	if total <= 0 {
		return 30
	}
	return total
}

// serviceDeposit devuelve la señal exigida por un servicio (0 si no exige).
func serviceDeposit(s model.Service) float64 {
	if !s.DepositRequired || s.DepositAmount <= 0 {
		return 0
	}
	return s.DepositAmount
}

// depositDue suma las señales de los servicios de una cita.
func depositDue(items []model.AppointmentService) float64 {
	total := 0.0
	for _, it := range items {
		if it.DepositAmount > 0 {
			total += it.DepositAmount
		}
	}
	return round2(total)
}

func normalizePaymentMethod(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if model.IsPaymentMethod(v) {
		return v
	}
	return "efectivo"
}

// depositRequested interpreta el checkbox de «señal cobrada» de los formularios.
func depositRequested(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "on", "true", "si", "sí", "yes":
		return true
	}
	return false
}

// registerDeposit cobra la señal que quede pendiente de una cita y la registra
// en la caja del día. El importe 0 significa «cobrar lo que falte».
func (h *Handler) registerDeposit(ctx context.Context, userID, appointmentID int64, method string) (float64, error) {
	return h.store.RegisterAppointmentDeposit(ctx, userID, appointmentID, 0, normalizePaymentMethod(method))
}

// settleAppointment aplica el flujo de facturación de una cita: «facturada»
// genera la factura y «abonada» además la cobra y registra el ingreso en caja.
func (h *Handler) settleAppointment(ctx context.Context, userID int64, appt model.Appointment, status, method string) (int64, error) {
	if status != "facturada" && status != "abonada" {
		return 0, nil
	}
	invoiceID, err := h.invoiceForAppointment(ctx, userID, appt)
	if err != nil {
		return 0, err
	}
	if status == "abonada" {
		if _, err := h.store.MarkInvoicePaid(ctx, userID, invoiceID, normalizePaymentMethod(method)); err != nil {
			return 0, err
		}
	}
	return invoiceID, nil
}

// invoiceForAppointment devuelve la factura de la cita y la crea si no existe.
func (h *Handler) invoiceForAppointment(ctx context.Context, userID int64, appt model.Appointment) (int64, error) {
	if existing, err := h.store.InvoiceByAppointment(ctx, userID, appt.ID); err == nil {
		return existing.ID, nil
	} else if !store.IsNotFound(err) {
		return 0, err
	}

	if appt.PatientID == 0 || len(appt.Services) == 0 {
		loaded, err := h.store.AppointmentByID(ctx, userID, appt.ID)
		if err != nil {
			return 0, err
		}
		if len(appt.Services) > 0 {
			loaded.Services = appt.Services
		}
		appt = loaded
	}

	billing, err := h.store.BillingSettings(ctx, userID)
	if err != nil {
		return 0, err
	}

	items := make([]model.InvoiceItem, 0, len(appt.Services))
	names := make([]string, 0, len(appt.Services))
	subtotal := 0.0
	for _, svc := range appt.Services {
		serviceID := svc.ServiceID
		items = append(items, model.InvoiceItem{
			ServiceID:   &serviceID,
			Description: svc.Name,
			Quantity:    1,
			UnitPrice:   svc.Price,
			VATRate:     billing.VATRate,
		})
		subtotal += svc.Price
		names = append(names, svc.Name)
	}
	subtotal = round2(subtotal)
	vatAmount := round2(subtotal * billing.VATRate / 100)
	concept := strings.Join(names, ", ")
	if concept == "" {
		concept = "Cita"
	}
	number, err := h.nextInvoiceNumberFor(ctx, userID)
	if err != nil {
		return 0, err
	}
	appointmentID := appt.ID
	inv := model.Invoice{
		UserID:        userID,
		PatientID:     appt.PatientID,
		AppointmentID: &appointmentID,
		Number:        number,
		Concept:       concept,
		Subtotal:      subtotal,
		VATRate:       billing.VATRate,
		VATAmount:     vatAmount,
		Amount:        round2(subtotal + vatAmount),
		Status:        "pendiente",
		IssuedAt:      time.Now().UTC(),
		Items:         items,
	}
	id, err := h.store.CreateInvoice(ctx, inv)
	if store.IsDuplicate(err) {
		if existing, e := h.store.InvoiceByAppointment(ctx, userID, appt.ID); e == nil {
			return existing.ID, nil
		}
	}
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (h *Handler) refreshAgenda(w http.ResponseWriter, r *http.Request, user model.User, view, date string) {
	if view == "" {
		view = "day"
	}
	if _, ok := parseDate(date); !ok {
		date = todayStr()
	}
	d, err := h.appData(r, "/app/agenda")
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	h.fillAgenda(r, &d, user, view, date)
	h.renderFragment(w, "agenda", d)
}

func (h *Handler) agendaError(w http.ResponseWriter, r *http.Request, user model.User, view, date, msg string) {
	if _, ok := parseDate(date); !ok {
		date = todayStr()
	}
	d, err := h.appData(r, "/app/agenda")
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	h.fillAgenda(r, &d, user, view, date)
	d.FormError = msg
	h.renderFragment(w, "agenda", d)
}

// --- Pacientes ---

func (h *Handler) Patients(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	d, err := h.appData(r, "/app/pacientes")
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	q := r.URL.Query().Get("q")
	patients, err := h.store.Patients(r.Context(), user.ID, q)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	d.Patients = patients
	d.Query = q
	d.Title = "Clientes | " + d.Site.SiteName

	if isHX(r) {
		if r.URL.Query().Get("partial") == "list" || r.Header.Get("HX-Target") == "patients-list" {
			h.renderFragment(w, "patients-list", d)
			return
		}
		h.renderFragment(w, "patients", d)
		return
	}
	h.renderApp(w, "patients", d)
}

func (h *Handler) CreatePatient(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		h.internalError(w, r, err)
		return
	}
	name := trim(r.FormValue("name"), 120)
	if name == "" {
		h.patientsError(w, r, user.ID, "Indica el nombre del cliente.")
		return
	}
	p := model.Patient{
		UserID: user.ID,
		Name:   name,
		Email:  trim(r.FormValue("email"), 160),
		Phone:  trim(r.FormValue("phone"), 40),
		Notes:  trim(r.FormValue("notes"), 2000),
	}
	if _, err := h.store.CreatePatient(r.Context(), p); err != nil {
		h.internalError(w, r, err)
		return
	}
	h.patientsRefresh(w, r, user.ID)
}

func (h *Handler) patientsError(w http.ResponseWriter, r *http.Request, userID int64, msg string) {
	d, err := h.appData(r, "/app/pacientes")
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	if patients, err := h.store.Patients(r.Context(), userID, ""); err == nil {
		d.Patients = patients
	}
	d.FormError = msg
	h.renderFragment(w, "patients", d)
}

func (h *Handler) patientsRefresh(w http.ResponseWriter, r *http.Request, userID int64) {
	d, err := h.appData(r, "/app/pacientes")
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	if patients, err := h.store.Patients(r.Context(), userID, ""); err == nil {
		d.Patients = patients
	}
	h.renderFragment(w, "patients", d)
}

func (h *Handler) PatientDetail(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.NotFound(w, r)
		return
	}
	patient, err := h.store.PatientByID(r.Context(), user.ID, id)
	if store.IsNotFound(err) {
		h.NotFound(w, r)
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	d, err := h.appData(r, "/app/pacientes")
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	d.Patient = &patient
	if appointments, err := h.store.PatientAppointments(r.Context(), user.ID, id, 10); err == nil {
		d.Appointments = appointments
	}
	if history, err := h.store.PatientHistory(r.Context(), user.ID, id); err == nil {
		d.PatientHistory = history
	}
	if services, err := h.store.PatientServicesDone(r.Context(), user.ID, id); err == nil {
		d.PatientServices = services
	}
	if invoices, err := h.store.Invoices(r.Context(), user.ID, ""); err == nil {
		var filtered []model.Invoice
		for _, inv := range invoices {
			if inv.PatientID == id {
				filtered = append(filtered, inv)
			}
		}
		d.Invoices = filtered
	}
	d.Title = patient.Name + " | " + d.Site.SiteName
	h.renderApp(w, "patient", d)
}

// --- Historial del cliente ---

func (h *Handler) CreatePatientHistory(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	patientID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.NotFound(w, r)
		return
	}
	if _, err := h.store.PatientByID(r.Context(), user.ID, patientID); store.IsNotFound(err) {
		h.NotFound(w, r)
		return
	} else if err != nil {
		h.internalError(w, r, err)
		return
	}
	if err := r.ParseForm(); err != nil {
		h.internalError(w, r, err)
		return
	}
	content := trim(r.FormValue("content"), 4000)
	if content == "" {
		http.Redirect(w, r, fmt.Sprintf("/app/pacientes/%d#historial", patientID), http.StatusSeeOther)
		return
	}
	entryType := trim(r.FormValue("entry_type"), 24)
	if !model.IsPatientHistoryType(entryType) {
		entryType = "nota"
	}
	authorID := user.ID
	entry := model.PatientHistoryEntry{
		UserID:    user.ID,
		PatientID: patientID,
		AuthorID:  &authorID,
		EntryType: entryType,
		Content:   content,
	}
	if _, err := h.store.CreatePatientHistory(r.Context(), entry); err != nil {
		h.internalError(w, r, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/app/pacientes/%d#historial", patientID), http.StatusSeeOther)
}

func (h *Handler) UpdatePatientHistory(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	patientID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.NotFound(w, r)
		return
	}
	id, err := strconv.ParseInt(r.PathValue("entryID"), 10, 64)
	if err != nil {
		h.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		h.internalError(w, r, err)
		return
	}
	content := trim(r.FormValue("content"), 4000)
	if content == "" {
		http.Redirect(w, r, fmt.Sprintf("/app/pacientes/%d#historial", patientID), http.StatusSeeOther)
		return
	}
	entryType := trim(r.FormValue("entry_type"), 24)
	if !model.IsPatientHistoryType(entryType) {
		entryType = "nota"
	}
	entry := model.PatientHistoryEntry{
		ID:        id,
		UserID:    user.ID,
		PatientID: patientID,
		EntryType: entryType,
		Content:   content,
	}
	if err := h.store.UpdatePatientHistory(r.Context(), entry); err != nil && !store.IsNotFound(err) {
		h.internalError(w, r, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/app/pacientes/%d#historial", patientID), http.StatusSeeOther)
}

func (h *Handler) DeletePatientHistory(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	patientID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.NotFound(w, r)
		return
	}
	id, err := strconv.ParseInt(r.PathValue("entryID"), 10, 64)
	if err != nil {
		h.NotFound(w, r)
		return
	}
	if err := h.store.DeletePatientHistory(r.Context(), user.ID, patientID, id); err != nil && !store.IsNotFound(err) {
		h.internalError(w, r, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/app/pacientes/%d#historial", patientID), http.StatusSeeOther)
}

func (h *Handler) UpdatePatient(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.NotFound(w, r)
		return
	}
	patient, err := h.store.PatientByID(r.Context(), user.ID, id)
	if store.IsNotFound(err) {
		h.NotFound(w, r)
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	if err := r.ParseForm(); err != nil {
		h.internalError(w, r, err)
		return
	}
	patient.Name = trim(r.FormValue("name"), 120)
	patient.Email = trim(r.FormValue("email"), 160)
	patient.Phone = trim(r.FormValue("phone"), 40)
	patient.Notes = trim(r.FormValue("notes"), 2000)
	if patient.Name == "" {
		patient.Name = "Sin nombre"
	}
	if err := h.store.UpdatePatient(r.Context(), patient); err != nil {
		h.internalError(w, r, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/app/pacientes/%d", id), http.StatusSeeOther)
}

// --- Servicios ---

func (h *Handler) Services(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	d, err := h.appData(r, "/app/servicios")
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	if services, err := h.store.Services(r.Context(), user.ID); err == nil {
		d.Services = services
	}
	d.Title = "Servicios | " + d.Site.SiteName

	if isHX(r) {
		h.renderFragment(w, "services", d)
		return
	}
	h.renderApp(w, "services", d)
}

func (h *Handler) CreateService(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		h.internalError(w, r, err)
		return
	}
	name := trim(r.FormValue("name"), 120)
	if name == "" {
		h.servicesError(w, r, user.ID, "Indica el nombre del servicio.")
		return
	}
	s := model.Service{
		UserID:          user.ID,
		Name:            name,
		DurationMinutes: parseDuration(r.FormValue("duration_minutes")),
		Price:           parsePrice(r.FormValue("price")),
		DepositRequired: depositRequested(r.FormValue("deposit_required")),
		DepositAmount:   parsePrice(r.FormValue("deposit_amount")),
		Color:           parseColor(r.FormValue("color")),
		Icon:            trim(r.FormValue("icon"), 32),
	}
	if !s.DepositRequired {
		s.DepositAmount = 0
	}
	if _, err := h.store.CreateService(r.Context(), s); err != nil {
		h.internalError(w, r, err)
		return
	}
	h.servicesRefresh(w, r, user.ID)
}

func (h *Handler) UpdateService(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.NotFound(w, r)
		return
	}
	s, err := h.store.ServiceByID(r.Context(), user.ID, id)
	if store.IsNotFound(err) {
		h.NotFound(w, r)
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	if err := r.ParseForm(); err != nil {
		h.internalError(w, r, err)
		return
	}
	name := trim(r.FormValue("name"), 120)
	if name == "" {
		h.servicesError(w, r, user.ID, "Indica el nombre del servicio.")
		return
	}
	s.Name = name
	s.DurationMinutes = parseDuration(r.FormValue("duration_minutes"))
	s.Price = parsePrice(r.FormValue("price"))
	s.DepositRequired = depositRequested(r.FormValue("deposit_required"))
	s.DepositAmount = parsePrice(r.FormValue("deposit_amount"))
	if !s.DepositRequired {
		s.DepositAmount = 0
	}
	s.Color = parseColor(r.FormValue("color"))
	s.Icon = trim(r.FormValue("icon"), 32)
	if err := h.store.UpdateService(r.Context(), s); err != nil {
		h.internalError(w, r, err)
		return
	}
	h.servicesRefresh(w, r, user.ID)
}

func (h *Handler) DeleteService(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.NotFound(w, r)
		return
	}
	if err := h.store.DeleteService(r.Context(), user.ID, id); err != nil && !store.IsNotFound(err) {
		h.internalError(w, r, err)
		return
	}
	h.servicesRefresh(w, r, user.ID)
}

func parseDuration(v string) int {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil || n < 5 || n > 480 {
		return 30
	}
	return n
}

func parsePrice(v string) float64 {
	v = strings.Replace(strings.TrimSpace(v), ",", ".", 1)
	n, err := strconv.ParseFloat(v, 64)
	if err != nil || n < 0 {
		return 0
	}
	return round2(n)
}

// parseColor normaliza un color hexadecimal (#RRGGBB); si es inválido devuelve "".
func parseColor(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if len(v) != 7 || v[0] != '#' {
		return ""
	}
	for _, c := range v[1:] {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return ""
		}
	}
	return v
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func (h *Handler) servicesError(w http.ResponseWriter, r *http.Request, userID int64, msg string) {
	d, err := h.appData(r, "/app/servicios")
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	if services, err := h.store.Services(r.Context(), userID); err == nil {
		d.Services = services
	}
	d.FormError = msg
	h.renderFragment(w, "services", d)
}

func (h *Handler) servicesRefresh(w http.ResponseWriter, r *http.Request, userID int64) {
	d, err := h.appData(r, "/app/servicios")
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	if services, err := h.store.Services(r.Context(), userID); err == nil {
		d.Services = services
	}
	h.renderFragment(w, "services", d)
}

// --- Facturación ---

func (h *Handler) Invoices(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	d, err := h.appData(r, "/app/facturacion")
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	status := r.URL.Query().Get("status")
	invoices, err := h.store.Invoices(r.Context(), user.ID, status)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	d.Invoices = invoices
	d.Query = status
	d.Patients, _ = h.store.Patients(r.Context(), user.ID, "")
	d.Title = "Facturación | " + d.Site.SiteName

	if isHX(r) {
		h.renderFragment(w, "invoices", d)
		return
	}
	h.renderApp(w, "invoices", d)
}

func (h *Handler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		h.internalError(w, r, err)
		return
	}
	patientID, err := strconv.ParseInt(r.FormValue("patient_id"), 10, 64)
	if err != nil || patientID <= 0 {
		h.invoicesError(w, r, user.ID, "", "Selecciona un cliente.")
		return
	}
	concept := trim(r.FormValue("concept"), 255)
	if concept == "" {
		concept = "Servicios prestados"
	}
	subtotal, err := strconv.ParseFloat(strings.Replace(strings.TrimSpace(r.FormValue("amount")), ",", ".", 1), 64)
	if err != nil || subtotal < 0 {
		h.invoicesError(w, r, user.ID, "", "Importe no válido.")
		return
	}
	subtotal = round2(subtotal)

	billing, err := h.store.BillingSettings(r.Context(), user.ID)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	vatAmount := round2(subtotal * billing.VATRate / 100)
	total := round2(subtotal + vatAmount)

	number, err := h.nextInvoiceNumber(r)
	if err != nil {
		h.internalError(w, r, err)
		return
	}

	inv := model.Invoice{
		UserID:    user.ID,
		PatientID: patientID,
		Number:    number,
		Concept:   concept,
		Subtotal:  subtotal,
		VATRate:   billing.VATRate,
		VATAmount: vatAmount,
		Amount:    total,
		Status:    "pendiente",
		IssuedAt:  time.Now().UTC(),
		Items: []model.InvoiceItem{{
			Description: concept,
			Quantity:    1,
			UnitPrice:   subtotal,
			VATRate:     billing.VATRate,
		}},
	}
	if _, err := h.store.CreateInvoice(r.Context(), inv); err != nil {
		h.internalError(w, r, err)
		return
	}
	h.invoicesRefresh(w, r, user.ID)
}

func (h *Handler) nextInvoiceNumber(r *http.Request) (string, error) {
	user, _ := userFromContext(r.Context())
	return h.nextInvoiceNumberFor(r.Context(), user.ID)
}

func (h *Handler) nextInvoiceNumberFor(ctx context.Context, userID int64) (string, error) {
	count, err := h.store.InvoiceCount(ctx, userID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("F-%d-%04d", time.Now().Year(), count+1), nil
}

// CreateInvoiceFromAppointment genera una factura con los servicios de una cita.
func (h *Handler) CreateInvoiceFromAppointment(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	appointmentID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "Cita no válida."})
		return
	}
	if existing, err := h.store.InvoiceByAppointment(r.Context(), user.ID, appointmentID); err == nil {
		_ = h.store.SetAppointmentStatus(r.Context(), user.ID, appointmentID, "facturada")
		respondJSON(w, http.StatusOK, map[string]any{"ok": true, "invoice_id": existing.ID, "existing": true})
		return
	} else if !store.IsNotFound(err) {
		h.internalError(w, r, err)
		return
	}

	appt, err := h.store.AppointmentByID(r.Context(), user.ID, appointmentID)
	if store.IsNotFound(err) {
		respondJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "Cita no encontrada."})
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}

	id, err := h.invoiceForAppointment(r.Context(), user.ID, appt)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	if err := h.store.SetAppointmentStatus(r.Context(), user.ID, appointmentID, "facturada"); err != nil {
		h.internalError(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"ok": true, "invoice_id": id})
}

// InvoiceDetail muestra la factura completa (página o fragmento para el modal).
func (h *Handler) InvoiceDetail(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.NotFound(w, r)
		return
	}
	invoice, err := h.store.InvoiceByID(r.Context(), user.ID, id)
	if store.IsNotFound(err) {
		h.NotFound(w, r)
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	billing, err := h.store.BillingSettings(r.Context(), user.ID)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	d, err := h.appData(r, "/app/facturacion")
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	d.Invoice = &invoice
	d.Billing = billing
	d.Title = "Factura " + invoice.Number + " | " + d.Site.SiteName
	if isHX(r) || r.URL.Query().Get("modal") == "1" {
		h.renderFragment(w, "invoice-detail", d)
		return
	}
	h.renderApp(w, "invoice", d)
}

// InvoicePDF devuelve la factura como PDF descargable.
func (h *Handler) InvoicePDF(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.NotFound(w, r)
		return
	}
	invoice, err := h.store.InvoiceByID(r.Context(), user.ID, id)
	if store.IsNotFound(err) {
		h.NotFound(w, r)
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	billing, err := h.store.BillingSettings(r.Context(), user.ID)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	data, err := pdf.Generate(billing, invoice)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", invoice.Number+".pdf"))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// InvoiceEdit devuelve el formulario de edición de una factura pendiente.
func (h *Handler) InvoiceEdit(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.NotFound(w, r)
		return
	}
	invoice, err := h.store.InvoiceByID(r.Context(), user.ID, id)
	if store.IsNotFound(err) {
		h.NotFound(w, r)
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	d, err := h.appData(r, "/app/facturacion")
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	d.Invoice = &invoice
	h.renderFragment(w, "invoice-edit", d)
}

// UpdateInvoice guarda las líneas e importes editados de una factura pendiente.
func (h *Handler) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		h.internalError(w, r, err)
		return
	}

	existing, err := h.store.InvoiceByID(r.Context(), user.ID, id)
	if store.IsNotFound(err) {
		h.NotFound(w, r)
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}

	vatRate := parsePrice(r.FormValue("vat_rate"))
	if vatRate < 0 || vatRate > 100 {
		vatRate = existing.VATRate
	}

	descriptions := r.Form["line_description"]
	quantities := r.Form["line_qty"]
	prices := r.Form["line_price"]
	items := make([]model.InvoiceItem, 0, len(descriptions))
	subtotal := 0.0
	for i, desc := range descriptions {
		desc = trim(desc, 255)
		if desc == "" {
			continue
		}
		qty := 1
		if i < len(quantities) {
			if n, err := strconv.Atoi(quantities[i]); err == nil && n > 0 {
				qty = n
			}
		}
		price := 0.0
		if i < len(prices) {
			price = parsePrice(prices[i])
		}
		items = append(items, model.InvoiceItem{
			Description: desc,
			Quantity:    qty,
			UnitPrice:   price,
			VATRate:     vatRate,
		})
		subtotal += round2(float64(qty) * price)
	}
	if len(items) == 0 {
		d, _ := h.appData(r, "/app/facturacion")
		d.Invoice = &existing
		d.FormError = "La factura necesita al menos una línea."
		h.renderFragment(w, "invoice-edit", d)
		return
	}

	subtotal = round2(subtotal)
	vatAmount := round2(subtotal * vatRate / 100)
	total := round2(subtotal + vatAmount)

	concept := trim(r.FormValue("concept"), 255)
	if concept == "" {
		names := make([]string, 0, len(items))
		for _, it := range items {
			names = append(names, it.Description)
		}
		concept = strings.Join(names, ", ")
	}

	inv := model.Invoice{
		ID:        id,
		UserID:    user.ID,
		Concept:   concept,
		Subtotal:  subtotal,
		VATRate:   vatRate,
		VATAmount: vatAmount,
		Amount:    total,
		Items:     items,
	}
	if err := h.store.UpdateInvoice(r.Context(), inv); err != nil {
		if store.IsNotFound(err) {
			h.NotFound(w, r)
			return
		}
		h.internalError(w, r, err)
		return
	}

	updated, err := h.store.InvoiceByID(r.Context(), user.ID, id)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	billing, err := h.store.BillingSettings(r.Context(), user.ID)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	d, err := h.appData(r, "/app/facturacion")
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	d.Invoice = &updated
	d.Billing = billing
	h.renderFragment(w, "invoice-detail", d)
}

func (h *Handler) invoicesError(w http.ResponseWriter, r *http.Request, userID int64, status, msg string) {
	d, err := h.appData(r, "/app/facturacion")
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	if invoices, err := h.store.Invoices(r.Context(), userID, status); err == nil {
		d.Invoices = invoices
	}
	d.Patients, _ = h.store.Patients(r.Context(), userID, "")
	d.Query = status
	d.FormError = msg
	h.renderFragment(w, "invoices", d)
}

func (h *Handler) invoicesRefresh(w http.ResponseWriter, r *http.Request, userID int64) {
	d, err := h.appData(r, "/app/facturacion")
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	if invoices, err := h.store.Invoices(r.Context(), userID, ""); err == nil {
		d.Invoices = invoices
	}
	d.Patients, _ = h.store.Patients(r.Context(), userID, "")
	h.renderFragment(w, "invoices", d)
}

func (h *Handler) PayInvoice(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		h.internalError(w, r, err)
		return
	}
	method := normalizePaymentMethod(r.FormValue("payment_method"))
	if _, err := h.store.MarkInvoicePaid(r.Context(), user.ID, id, method); err != nil {
		h.internalError(w, r, err)
		return
	}
	invoice, err := h.store.InvoiceByID(r.Context(), user.ID, id)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	d, err := h.appData(r, "/app/facturacion")
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	d.Invoice = &invoice
	if r.FormValue("detail") == "1" {
		if billing, err := h.store.BillingSettings(r.Context(), user.ID); err == nil {
			d.Billing = billing
		}
		h.renderFragment(w, "invoice-detail", d)
		return
	}
	h.renderFragment(w, "invoice-row", d)
}

// --- Caja diaria ---

func (h *Handler) Caja(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	d, err := h.appData(r, "/app/caja")
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	date := r.URL.Query().Get("date")
	if _, ok := parseDate(date); !ok {
		date = todayStr()
	}
	d.CashDate = date
	d.CashPrevDate = shiftDate(date, -1)
	d.CashNextDate = shiftDate(date, 1)
	d.CashIsToday = date == todayStr()
	d.FormOK = r.URL.Query().Get("ok") == "1"
	d.FormError = r.URL.Query().Get("error")

	session, err := h.store.CashSessionByDate(r.Context(), user.ID, date)
	if err == nil {
		d.CashSession = &session
		movements, err := h.store.CashMovements(r.Context(), user.ID, session.ID)
		if err != nil {
			h.internalError(w, r, err)
			return
		}
		d.CashMovements = movements
		d.CashSummary = model.SummarizeCash(movements, session.OpeningAmount, session.CountedAmount)
	} else if !store.IsNotFound(err) {
		h.internalError(w, r, err)
		return
	}

	if history, err := h.store.CashSessions(r.Context(), user.ID, 30); err == nil {
		d.CashHistory = history
	}
	d.Title = "Caja | " + d.Site.SiteName
	h.renderApp(w, "caja", d)
}

func (h *Handler) OpenCash(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		h.internalError(w, r, err)
		return
	}
	date := trim(r.FormValue("date"), 10)
	if _, ok := parseDate(date); !ok {
		date = todayStr()
	}
	opening := parsePrice(r.FormValue("opening_amount"))
	if _, err := h.store.CashSessionByDate(r.Context(), user.ID, date); err == nil {
		h.cajaRedirect(w, r, date, "Ya existe una caja para ese día.")
		return
	} else if !store.IsNotFound(err) {
		h.internalError(w, r, err)
		return
	}
	if _, err := h.store.CreateCashSession(r.Context(), user.ID, date, opening); err != nil {
		if store.IsDuplicate(err) {
			h.cajaRedirect(w, r, date, "Ya existe una caja para ese día.")
			return
		}
		h.internalError(w, r, err)
		return
	}
	http.Redirect(w, r, "/app/caja?date="+date+"&ok=1", http.StatusSeeOther)
}

func (h *Handler) CreateCashMovement(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		h.internalError(w, r, err)
		return
	}
	sessionID, err := strconv.ParseInt(r.FormValue("session_id"), 10, 64)
	if err != nil {
		h.cajaRedirect(w, r, todayStr(), "Caja no válida.")
		return
	}
	session, err := h.store.CashSessionByID(r.Context(), user.ID, sessionID)
	if store.IsNotFound(err) {
		h.cajaRedirect(w, r, todayStr(), "Caja no encontrada.")
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	if !session.IsOpen() {
		h.cajaRedirect(w, r, session.Date, "La caja está cerrada. Reábrela antes de añadir movimientos.")
		return
	}
	amount := parsePrice(r.FormValue("amount"))
	if amount <= 0 {
		h.cajaRedirect(w, r, session.Date, "Indica un importe mayor que cero.")
		return
	}
	movementType := trim(r.FormValue("movement_type"), 16)
	if movementType != "gasto" {
		movementType = "ingreso"
	}
	concept := trim(r.FormValue("concept"), 255)
	if concept == "" {
		if movementType == "gasto" {
			concept = "Gasto"
		} else {
			concept = "Ingreso"
		}
	}
	movement := model.CashMovement{
		UserID:        user.ID,
		SessionID:     session.ID,
		MovementType:  movementType,
		PaymentMethod: normalizePaymentMethod(r.FormValue("payment_method")),
		Amount:        amount,
		Concept:       concept,
	}
	if _, err := h.store.CreateCashMovement(r.Context(), movement); err != nil {
		h.internalError(w, r, err)
		return
	}
	http.Redirect(w, r, "/app/caja?date="+session.Date+"&ok=1", http.StatusSeeOther)
}

func (h *Handler) DeleteCashMovement(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		h.internalError(w, r, err)
		return
	}
	date := trim(r.FormValue("date"), 10)
	if _, ok := parseDate(date); !ok {
		date = todayStr()
	}
	if err := h.store.DeleteCashMovement(r.Context(), user.ID, id); err != nil {
		if store.IsNotFound(err) {
			h.cajaRedirect(w, r, date, "No se pudo eliminar el movimiento.")
			return
		}
		h.internalError(w, r, err)
		return
	}
	http.Redirect(w, r, "/app/caja?date="+date+"&ok=1", http.StatusSeeOther)
}

func (h *Handler) CloseCash(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		h.internalError(w, r, err)
		return
	}
	sessionID, err := strconv.ParseInt(r.FormValue("session_id"), 10, 64)
	if err != nil {
		h.NotFound(w, r)
		return
	}
	session, err := h.store.CashSessionByID(r.Context(), user.ID, sessionID)
	if store.IsNotFound(err) {
		h.NotFound(w, r)
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	if !session.IsOpen() {
		h.cajaRedirect(w, r, session.Date, "La caja ya está cerrada.")
		return
	}

	counted := parsePrice(r.FormValue("counted_amount"))
	movements, err := h.store.CashMovements(r.Context(), user.ID, session.ID)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	summary := model.SummarizeCash(movements, session.OpeningAmount, counted)
	session.CountedAmount = counted
	session.ExpectedCash = summary.ExpectedCash
	session.Difference = summary.Difference
	session.Denoms = []model.CashDenomination{}
	session.Notes = trim(r.FormValue("notes"), 500)
	if err := h.store.CloseCashSession(r.Context(), session); err != nil {
		if store.IsNotFound(err) {
			h.cajaRedirect(w, r, session.Date, "La caja ya está cerrada.")
			return
		}
		h.internalError(w, r, err)
		return
	}
	http.Redirect(w, r, "/app/caja?date="+session.Date+"&ok=1", http.StatusSeeOther)
}

func (h *Handler) cajaRedirect(w http.ResponseWriter, r *http.Request, date, msg string) {
	http.Redirect(w, r, "/app/caja?date="+url.QueryEscape(date)+"&error="+url.QueryEscape(msg), http.StatusSeeOther)
}

// --- Leads (dentro de la app) ---

func (h *Handler) AppLeads(w http.ResponseWriter, r *http.Request) {
	d, err := h.appData(r, "/app/leads")
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
	h.renderApp(w, "leads", d)
}

// --- Configuración ---

var settingsSections = map[string]string{
	"apariencia":  "Apariencia",
	"facturacion": "Facturación",
	"horario":     "Horario del centro",
	"equipo":      "Equipo",
	"bloqueos":    "Bloqueos y festivos",
}

func (h *Handler) SettingsRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/app/configuracion/apariencia", http.StatusSeeOther)
}

func (h *Handler) Settings(w http.ResponseWriter, r *http.Request) {
	section := r.PathValue("section")
	title, ok := settingsSections[section]
	if !ok {
		h.NotFound(w, r)
		return
	}
	user, _ := userFromContext(r.Context())
	d, err := h.appData(r, "/app/configuracion/"+section)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	d.SettingsSection = section
	d.SettingsTitle = title

	switch section {
	case "apariencia":
		if d.Theme, err = h.store.Preferences(r.Context(), user.ID); err != nil {
			h.internalError(w, r, err)
			return
		}
		d.ThemePresets = model.ThemePresets()
	case "facturacion":
		if d.Billing, err = h.store.BillingSettings(r.Context(), user.ID); err != nil {
			h.internalError(w, r, err)
			return
		}
	case "horario":
		if d.ClinicSchedule, err = h.store.ClinicSchedule(r.Context(), user.ID); err != nil {
			d.ClinicSchedule = model.DefaultSchedule()
		}
	case "equipo":
		d.Staff, _ = h.store.Staff(r.Context(), user.ID)
	case "bloqueos":
		if d.ScheduleBlocks, err = h.store.ScheduleBlocks(r.Context(), user.ID); err != nil {
			d.ScheduleBlocks = nil
		}
		if d.Holidays, err = h.store.Holidays(r.Context(), user.ID); err != nil {
			d.Holidays = nil
		}
		d.Staff, _ = h.store.Staff(r.Context(), user.ID)
	}

	d.FormOK = r.URL.Query().Get("ok") == "1"
	d.FormError = r.URL.Query().Get("error")
	d.Title = "Configuración · " + title + " | " + d.Site.SiteName
	h.renderApp(w, "settings-"+section, d)
}

func (h *Handler) SaveSettings(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		h.internalError(w, r, err)
		return
	}

	preset := trim(r.FormValue("preset"), 32)
	var theme model.Theme
	if preset != "" && preset != "custom" {
		if p, ok := model.PresetBySlug(preset); ok {
			theme = p
		} else {
			theme = model.DefaultTheme()
		}
	} else {
		current, _ := h.store.Preferences(r.Context(), user.ID)
		theme = model.Theme{
			Preset:        "custom",
			Bg:            hexOr(r.FormValue("bg"), current.Bg),
			Surface:       hexOr(r.FormValue("surface"), current.Surface),
			Surface2:      hexOr(r.FormValue("surface2"), current.Surface2),
			Text:          hexOr(r.FormValue("text"), current.Text),
			Muted:         hexOr(r.FormValue("muted"), current.Muted),
			Accent:        hexOr(r.FormValue("accent"), current.Accent),
			AccentPressed: hexOr(r.FormValue("accentPressed"), current.AccentPressed),
			AccentSoft:    hexOr(r.FormValue("accentSoft"), current.AccentSoft),
			Gold:          hexOr(r.FormValue("gold"), current.Gold),
			Sage:          hexOr(r.FormValue("sage"), current.Sage),
		}
	}
	if r.FormValue("reset") == "1" {
		theme = model.DefaultTheme()
	}
	if err := h.store.SavePreferences(r.Context(), user.ID, theme); err != nil {
		h.internalError(w, r, err)
		return
	}
	http.Redirect(w, r, "/app/configuracion/apariencia?ok=1", http.StatusSeeOther)
}

func (h *Handler) SaveBillingSettings(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		h.internalError(w, r, err)
		return
	}
	rate := parsePrice(r.FormValue("vat_rate"))
	if rate < 0 || rate > 100 {
		rate = 21
	}
	b := model.BillingSettings{
		UserID:      user.ID,
		CompanyName: trim(r.FormValue("company_name"), 160),
		TaxID:       trim(r.FormValue("tax_id"), 40),
		Address:     trim(r.FormValue("address"), 255),
		Phone:       trim(r.FormValue("phone"), 40),
		Email:       trim(r.FormValue("email"), 160),
		VATRate:     rate,
	}
	if err := h.store.SaveBillingSettings(r.Context(), b); err != nil {
		h.internalError(w, r, err)
		return
	}
	http.Redirect(w, r, "/app/configuracion/facturacion?ok=1", http.StatusSeeOther)
}

func (h *Handler) SaveClinicSchedule(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		h.internalError(w, r, err)
		return
	}
	days := make([]model.ScheduleDay, 7)
	for i := 0; i < 7; i++ {
		open, okOpen := normalizeClock(r.FormValue(fmt.Sprintf("open_%d", i)))
		close, okClose := normalizeClock(r.FormValue(fmt.Sprintf("close_%d", i)))
		if !okOpen {
			open = ""
		}
		if !okClose {
			close = ""
		}
		if r.FormValue(fmt.Sprintf("closed_%d", i)) == "1" {
			open, close = "", ""
		}
		days[i] = model.ScheduleDay{Day: i, Open: open, Close: close}
	}
	if err := h.store.SaveClinicSchedule(r.Context(), user.ID, days); err != nil {
		h.internalError(w, r, err)
		return
	}
	http.Redirect(w, r, "/app/configuracion/horario?ok=1", http.StatusSeeOther)
}

func (h *Handler) CreateScheduleBlock(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		h.internalError(w, r, err)
		return
	}
	staffID, err := strconv.ParseInt(r.FormValue("staff_id"), 10, 64)
	if err != nil || staffID <= 0 {
		http.Redirect(w, r, "/app/configuracion/bloqueos?error="+url.QueryEscape("Selecciona un empleado."), http.StatusSeeOther)
		return
	}
	start, okStart := normalizeClock(r.FormValue("start"))
	end, okEnd := normalizeClock(r.FormValue("end"))
	ts, errS := time.Parse("15:04", start)
	te, errE := time.Parse("15:04", end)
	if !okStart || !okEnd || errS != nil || errE != nil || !te.After(ts) {
		http.Redirect(w, r, "/app/configuracion/bloqueos?error="+url.QueryEscape("Horas no válidas."), http.StatusSeeOther)
		return
	}
	reason := trim(r.FormValue("reason"), 120)
	b := model.ScheduleBlock{
		UserID:  user.ID,
		StaffID: staffID,
		Start:   start,
		End:     end,
		Reason:  reason,
	}

	if r.FormValue("all_days") == "1" {
		existing, err := h.store.ScheduleBlocks(r.Context(), user.ID)
		if err != nil {
			h.internalError(w, r, err)
			return
		}
		b.SeriesID = time.Now().UnixNano()
		for weekday := 0; weekday < 7; weekday++ {
			if hasExactBlock(existing, staffID, weekday, start, end, reason) {
				continue
			}
			b.Weekday = weekday
			if _, err := h.store.CreateScheduleBlock(r.Context(), b); err != nil {
				h.internalError(w, r, err)
				return
			}
		}
		http.Redirect(w, r, "/app/configuracion/bloqueos?ok=1", http.StatusSeeOther)
		return
	}

	weekday, err := strconv.Atoi(r.FormValue("weekday"))
	if err != nil || weekday < 0 || weekday > 6 {
		http.Redirect(w, r, "/app/configuracion/bloqueos?error="+url.QueryEscape("Día de la semana no válido."), http.StatusSeeOther)
		return
	}
	b.Weekday = weekday
	if _, err := h.store.CreateScheduleBlock(r.Context(), b); err != nil {
		h.internalError(w, r, err)
		return
	}
	http.Redirect(w, r, "/app/configuracion/bloqueos?ok=1", http.StatusSeeOther)
}

func (h *Handler) UpdateScheduleBlock(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		h.internalError(w, r, err)
		return
	}
	blocks, err := h.store.ScheduleBlocks(r.Context(), user.ID)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	var current *model.ScheduleBlock
	for i := range blocks {
		if blocks[i].ID == id {
			current = &blocks[i]
			break
		}
	}
	if current == nil {
		http.Redirect(w, r, "/app/configuracion/bloqueos?error="+url.QueryEscape("Bloque no encontrado."), http.StatusSeeOther)
		return
	}
	staffID, err := strconv.ParseInt(r.FormValue("staff_id"), 10, 64)
	if err != nil || staffID <= 0 {
		http.Redirect(w, r, "/app/configuracion/bloqueos?error="+url.QueryEscape("Selecciona un empleado."), http.StatusSeeOther)
		return
	}
	start, okStart := normalizeClock(r.FormValue("start"))
	end, okEnd := normalizeClock(r.FormValue("end"))
	ts, errS := time.Parse("15:04", start)
	te, errE := time.Parse("15:04", end)
	if !okStart || !okEnd || errS != nil || errE != nil || !te.After(ts) {
		http.Redirect(w, r, "/app/configuracion/bloqueos?error="+url.QueryEscape("Horas no válidas."), http.StatusSeeOther)
		return
	}
	b := model.ScheduleBlock{
		ID:      id,
		UserID:  user.ID,
		StaffID: staffID,
		Weekday: current.Weekday,
		Start:   start,
		End:     end,
		Reason:  trim(r.FormValue("reason"), 120),
	}

	if r.FormValue("all_days") == "1" {
		if current.SeriesID > 0 {
			b.SeriesID = current.SeriesID
			if err := h.store.UpdateScheduleSeries(r.Context(), user.ID, b.SeriesID, b); err != nil {
				h.internalError(w, r, err)
				return
			}
		} else {
			b.SeriesID = time.Now().UnixNano()
			if err := h.store.UpdateScheduleBlock(r.Context(), b); err != nil {
				h.internalError(w, r, err)
				return
			}
			for weekday := 0; weekday < 7; weekday++ {
				if weekday == current.Weekday {
					continue
				}
				next := b
				next.ID = 0
				next.Weekday = weekday
				if _, err := h.store.CreateScheduleBlock(r.Context(), next); err != nil {
					h.internalError(w, r, err)
					return
				}
			}
		}
		http.Redirect(w, r, "/app/configuracion/bloqueos?ok=1", http.StatusSeeOther)
		return
	}

	weekday, err := strconv.Atoi(r.FormValue("weekday"))
	if err != nil || weekday < 0 || weekday > 6 {
		http.Redirect(w, r, "/app/configuracion/bloqueos?error="+url.QueryEscape("Día de la semana no válido."), http.StatusSeeOther)
		return
	}
	b.Weekday = weekday
	b.SeriesID = 0
	if err := h.store.UpdateScheduleBlock(r.Context(), b); err != nil {
		h.internalError(w, r, err)
		return
	}
	http.Redirect(w, r, "/app/configuracion/bloqueos?ok=1", http.StatusSeeOther)
}

func normalizeClock(s string) (string, bool) {
	parts := strings.Split(strings.TrimSpace(s), ":")
	if len(parts) != 2 {
		return "", false
	}
	hour, errH := strconv.Atoi(strings.TrimSpace(parts[0]))
	minute, errM := strconv.Atoi(strings.TrimSpace(parts[1]))
	if errH != nil || errM != nil || hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return "", false
	}
	return fmt.Sprintf("%02d:%02d", hour, minute), true
}

func hasExactBlock(blocks []model.ScheduleBlock, staffID int64, weekday int, start, end, reason string) bool {
	for _, b := range blocks {
		if b.StaffID == staffID && b.Weekday == weekday && b.Start == start && b.End == end && b.Reason == reason {
			return true
		}
	}
	return false
}

func (h *Handler) CreateHoliday(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		h.internalError(w, r, err)
		return
	}
	date := trim(r.FormValue("date"), 10)
	if _, err := time.Parse(dateLayout, date); err != nil {
		http.Redirect(w, r, "/app/configuracion/bloqueos?error="+url.QueryEscape("Fecha no válida."), http.StatusSeeOther)
		return
	}
	holiday := model.Holiday{
		UserID: user.ID,
		Date:   date,
		Name:   trim(r.FormValue("name"), 120),
	}
	if err := h.store.CreateHoliday(r.Context(), holiday); err != nil {
		h.internalError(w, r, err)
		return
	}
	http.Redirect(w, r, "/app/configuracion/bloqueos?ok=1", http.StatusSeeOther)
}

func (h *Handler) DeleteHoliday(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.NotFound(w, r)
		return
	}
	if err := h.store.DeleteHoliday(r.Context(), user.ID, id); err != nil && !store.IsNotFound(err) {
		h.internalError(w, r, err)
		return
	}
	http.Redirect(w, r, "/app/configuracion/bloqueos?ok=1", http.StatusSeeOther)
}

func (h *Handler) DeleteScheduleBlock(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.NotFound(w, r)
		return
	}
	if err := h.store.DeleteScheduleBlock(r.Context(), user.ID, id); err != nil && !store.IsNotFound(err) {
		h.internalError(w, r, err)
		return
	}
	http.Redirect(w, r, "/app/configuracion/bloqueos?ok=1", http.StatusSeeOther)
}

// UpdateStaff guarda el icono y color de un empleado del equipo.
func (h *Handler) UpdateStaff(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	staffID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		h.internalError(w, r, err)
		return
	}
	color := parseColor(r.FormValue("color"))
	icon := trim(r.FormValue("icon"), 32)
	if err := h.store.UpdateStaffStyle(r.Context(), user.ID, staffID, color, icon); err != nil {
		if store.IsNotFound(err) {
			h.NotFound(w, r)
			return
		}
		h.internalError(w, r, err)
		return
	}
	if isHX(r) {
		d, err := h.appData(r, "/app/configuracion/equipo")
		if err != nil {
			h.internalError(w, r, err)
			return
		}
		d.SettingsSection = "equipo"
		d.SettingsTitle = "Equipo"
		d.FormOK = true
		d.Staff, _ = h.store.Staff(r.Context(), user.ID)
		h.renderFragment(w, "settings-equipo", d)
		return
	}
	http.Redirect(w, r, "/app/configuracion/equipo?ok=1", http.StatusSeeOther)
}

func hexOr(v, fallback string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if len(v) != 7 || v[0] != '#' {
		return fallback
	}
	for _, c := range v[1:] {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return fallback
		}
	}
	return v
}

// --- Perfil de usuario y conectados ---

type userProfileUpdate struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Color           string `json:"color"`
	Icon            string `json:"icon"`
	Avatar          string `json:"avatar"`
	Password        string `json:"password"`
	CurrentPassword string `json:"current_password"`
}

// UserProfile devuelve los datos editables de un usuario del equipo.
func (h *Handler) UserProfile(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "Usuario no válido."})
		return
	}
	target, err := h.store.UserForOwner(r.Context(), user.ID, id)
	if store.IsNotFound(err) {
		respondJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "Usuario no encontrado."})
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"ok":         true,
		"id":         target.ID,
		"name":       target.Name,
		"email":      target.Email,
		"color":      target.Color,
		"icon":       target.Icon,
		"avatar":     target.Avatar,
		"role":       target.Role,
		"is_manager": target.IsManager(),
		"is_self":    target.ID == user.ID,
	})
}

// UpdateUserProfile guarda los datos del perfil, el avatar y, opcionalmente,
// una nueva contraseña.
func (h *Handler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "Usuario no válido."})
		return
	}
	var req userProfileUpdate
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "Datos no válidos."})
		return
	}
	target, err := h.store.UserForOwner(r.Context(), user.ID, id)
	if store.IsNotFound(err) {
		respondJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "Usuario no encontrado."})
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}

	name := trim(req.Name, 120)
	if name == "" {
		respondJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "Indica el nombre."})
		return
	}
	email := strings.ToLower(trim(req.Email, 160))
	if _, err := mail.ParseAddress(email); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "El email no es válido."})
		return
	}
	taken, err := h.store.EmailTaken(r.Context(), email, target.ID)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	if taken {
		respondJSON(w, http.StatusConflict, map[string]any{"ok": false, "error": "Ese email ya está en uso."})
		return
	}
	avatar, ok := normalizeAvatar(req.Avatar)
	if !ok {
		respondJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "El avatar no es válido o es demasiado grande."})
		return
	}

	target.Name = name
	target.Email = email
	target.Color = parseColor(req.Color)
	target.Icon = trim(req.Icon, 32)
	target.Avatar = avatar
	if err := h.store.UpdateUserProfile(r.Context(), user.ID, target); err != nil {
		h.internalError(w, r, err)
		return
	}

	if req.Password != "" {
		if len([]rune(req.Password)) < 8 {
			respondJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "La nueva contraseña debe tener al menos 8 caracteres."})
			return
		}
		if target.ID == user.ID {
			if req.CurrentPassword == "" ||
				bcrypt.CompareHashAndPassword([]byte(target.PasswordHash), []byte(req.CurrentPassword)) != nil {
				respondJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "La contraseña actual no es correcta."})
				return
			}
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			h.internalError(w, r, err)
			return
		}
		if err := h.store.UpdateUserPassword(r.Context(), target.ID, string(hash)); err != nil {
			h.internalError(w, r, err)
			return
		}
	}

	respondJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// UserAvatar sirve el avatar de un usuario del equipo.
func (h *Handler) UserAvatar(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.NotFound(w, r)
		return
	}
	target, err := h.store.UserForOwner(r.Context(), user.ID, id)
	if store.IsNotFound(err) {
		h.NotFound(w, r)
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	if target.Avatar == "" {
		h.NotFound(w, r)
		return
	}
	contentType, data, ok := decodeAvatar(target.Avatar)
	if !ok {
		h.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "private, no-store")
	_, _ = w.Write(data)
}

// normalizeAvatar valida un avatar en data URI de imagen.
func normalizeAvatar(v string) (string, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", true
	}
	if len(v) > 400_000 {
		return "", false
	}
	for _, prefix := range []string{
		"data:image/png;base64,", "data:image/jpeg;base64,",
		"data:image/jpg;base64,", "data:image/webp;base64,",
		"data:image/gif;base64,",
	} {
		if strings.HasPrefix(v, prefix) {
			return v, true
		}
	}
	return "", false
}

// decodeAvatar extrae el tipo MIME y los bytes de un data URI de imagen.
func decodeAvatar(v string) (string, []byte, bool) {
	if !strings.HasPrefix(v, "data:image/") {
		return "", nil, false
	}
	marker := ";base64,"
	i := strings.Index(v, marker)
	if i <= len("data:image/") {
		return "", nil, false
	}
	contentType := v[len("data:"):i]
	data, err := base64.StdEncoding.DecodeString(v[i+len(marker):])
	if err != nil {
		return "", nil, false
	}
	return contentType, data, true
}
