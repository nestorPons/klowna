package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"klowna/internal/model"
)

// --- Autenticación ---

func (st *Store) UserByEmail(ctx context.Context, email string) (model.User, error) {
	var u model.User
	var managerID sql.NullInt64
	err := st.db.QueryRowContext(ctx,
		`SELECT id, email, name, color, icon, avatar, password_hash, role, manager_id FROM users WHERE email = ?`, email).
		Scan(&u.ID, &u.Email, &u.Name, &u.Color, &u.Icon, &u.Avatar, &u.PasswordHash, &u.Role, &managerID)
	if managerID.Valid {
		u.ManagerID = &managerID.Int64
	}
	return u, err
}

func (st *Store) CreateSession(ctx context.Context, tokenHash string, userID int64, expiresAt time.Time) error {
	_, err := st.db.ExecContext(ctx,
		`INSERT INTO sessions (token_hash, user_id, expires_at, created_at) VALUES (?, ?, ?, ?)`,
		tokenHash, userID, expiresAt.UTC(), time.Now().UTC())
	return err
}

func (st *Store) UserBySession(ctx context.Context, tokenHash string) (model.User, error) {
	var u model.User
	var managerID sql.NullInt64
	err := st.db.QueryRowContext(ctx,
		`SELECT u.id, u.email, u.name, u.color, u.icon, u.avatar, u.password_hash, u.role, u.manager_id
		 FROM sessions s JOIN users u ON u.id = s.user_id
		 WHERE s.token_hash = ? AND s.expires_at > ?`, tokenHash, time.Now().UTC()).
		Scan(&u.ID, &u.Email, &u.Name, &u.Color, &u.Icon, &u.Avatar, &u.PasswordHash, &u.Role, &managerID)
	if managerID.Valid {
		u.ManagerID = &managerID.Int64
	}
	return u, err
}

func (st *Store) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := st.db.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash = ?`, tokenHash)
	return err
}

// --- Equipo ---

// Staff devuelve los empleados asignables de un gestor, ordenados por nombre.
func (st *Store) Staff(ctx context.Context, managerID int64) ([]model.User, error) {
	rows, err := st.db.QueryContext(ctx,
		`SELECT id, email, name, color, icon, avatar, role FROM users WHERE manager_id = ? AND role = 'staff' ORDER BY name ASC`, managerID)
	if err != nil {
		return nil, fmt.Errorf("staff: %w", err)
	}
	defer rows.Close()

	out := make([]model.User, 0, 8)
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Color, &u.Icon, &u.Avatar, &u.Role); err != nil {
			return nil, fmt.Errorf("scan staff: %w", err)
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// ConnectedUsers devuelve los usuarios con sesión activa que puede ver el
// usuario actual: él mismo y, si es gestor, su equipo.
func (st *Store) ConnectedUsers(ctx context.Context, ownerID int64, now time.Time) ([]model.ConnectedUser, error) {
	rows, err := st.db.QueryContext(ctx,
		`SELECT u.id, u.email, u.name, u.color, u.icon, MAX(u.avatar), u.role, u.manager_id,
			MAX(s.created_at), COUNT(*)
		 FROM sessions s
		 JOIN users u ON u.id = s.user_id
		 WHERE s.expires_at > ? AND (u.id = ? OR u.manager_id = ?)
		 GROUP BY u.id, u.email, u.name, u.color, u.icon, u.role, u.manager_id
		 ORDER BY u.name ASC`, now.UTC(), ownerID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("connected users: %w", err)
	}
	defer rows.Close()

	out := make([]model.ConnectedUser, 0, 8)
	for rows.Next() {
		var c model.ConnectedUser
		var managerID sql.NullInt64
		if err := rows.Scan(&c.ID, &c.Email, &c.Name, &c.Color, &c.Icon, &c.Avatar, &c.Role,
			&managerID, &c.LastLogin, &c.Sessions); err != nil {
			return nil, fmt.Errorf("scan connected user: %w", err)
		}
		if managerID.Valid {
			c.ManagerID = &managerID.Int64
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// UserForOwner devuelve un usuario si es el propio usuario o un empleado suyo.
func (st *Store) UserForOwner(ctx context.Context, ownerID, id int64) (model.User, error) {
	var u model.User
	var managerID sql.NullInt64
	err := st.db.QueryRowContext(ctx,
		`SELECT id, email, name, color, icon, avatar, password_hash, role, manager_id
		 FROM users WHERE id = ? AND (id = ? OR manager_id = ?)`,
		id, ownerID, ownerID).
		Scan(&u.ID, &u.Email, &u.Name, &u.Color, &u.Icon, &u.Avatar, &u.PasswordHash, &u.Role, &managerID)
	if managerID.Valid {
		u.ManagerID = &managerID.Int64
	}
	return u, err
}

// UpdateUserProfile guarda los datos editables del perfil.
func (st *Store) UpdateUserProfile(ctx context.Context, ownerID int64, u model.User) error {
	res, err := st.db.ExecContext(ctx,
		`UPDATE users SET name = ?, email = ?, color = ?, icon = ?, avatar = ?
		 WHERE id = ? AND (id = ? OR manager_id = ?)`,
		u.Name, u.Email, u.Color, u.Icon, u.Avatar, u.ID, ownerID, ownerID)
	if err != nil {
		return fmt.Errorf("update user profile: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		var exists int
		if err := st.db.QueryRowContext(ctx,
			`SELECT 1 FROM users WHERE id = ? AND (id = ? OR manager_id = ?)`,
			u.ID, ownerID, ownerID).Scan(&exists); err != nil {
			return err
		}
	}
	return nil
}

// UpdateUserPassword cambia la contraseña de un usuario ya validado.
func (st *Store) UpdateUserPassword(ctx context.Context, id int64, hash string) error {
	_, err := st.db.ExecContext(ctx, `UPDATE users SET password_hash = ? WHERE id = ?`, hash, id)
	if err != nil {
		return fmt.Errorf("update user password: %w", err)
	}
	return nil
}

// EmailTaken indica si otro usuario usa ya ese correo.
func (st *Store) EmailTaken(ctx context.Context, email string, exceptID int64) (bool, error) {
	var n int
	err := st.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM users WHERE email = ? AND id <> ?`, email, exceptID).Scan(&n)
	if err != nil {
		return false, fmt.Errorf("check email: %w", err)
	}
	return n > 0, nil
}

// UpdateStaffStyle actualiza el icono y color de un empleado del gestor.
func (st *Store) UpdateStaffStyle(ctx context.Context, managerID, staffID int64, color, icon string) error {
	res, err := st.db.ExecContext(ctx,
		`UPDATE users SET color = ?, icon = ? WHERE id = ? AND manager_id = ? AND role = 'staff'`,
		color, icon, staffID, managerID)
	if err != nil {
		return fmt.Errorf("update staff style: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		var exists int
		err := st.db.QueryRowContext(ctx,
			`SELECT 1 FROM users WHERE id = ? AND manager_id = ? AND role = 'staff'`,
			staffID, managerID).Scan(&exists)
		if err == sql.ErrNoRows {
			return sql.ErrNoRows
		}
		if err != nil {
			return fmt.Errorf("check staff: %w", err)
		}
	}
	return nil
}

// --- Pacientes ---

func (st *Store) Patients(ctx context.Context, userID int64, query string) ([]model.Patient, error) {
	base := `SELECT id, user_id, name, email, phone, notes, created_at FROM patients WHERE user_id = ?`
	args := []any{userID}
	if q := strings.TrimSpace(query); q != "" {
		like := "%" + q + "%"
		base += ` AND (name LIKE ? OR email LIKE ? OR phone LIKE ?)`
		args = append(args, like, like, like)
	}
	base += ` ORDER BY name ASC`

	rows, err := st.db.QueryContext(ctx, base, args...)
	if err != nil {
		return nil, fmt.Errorf("patients: %w", err)
	}
	defer rows.Close()

	out := make([]model.Patient, 0, 32)
	for rows.Next() {
		var p model.Patient
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Email, &p.Phone, &p.Notes, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan patient: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (st *Store) PatientByID(ctx context.Context, userID, id int64) (model.Patient, error) {
	var p model.Patient
	err := st.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, email, phone, notes, created_at FROM patients WHERE id = ? AND user_id = ?`, id, userID).
		Scan(&p.ID, &p.UserID, &p.Name, &p.Email, &p.Phone, &p.Notes, &p.CreatedAt)
	return p, err
}

func (st *Store) CreatePatient(ctx context.Context, p model.Patient) (int64, error) {
	res, err := st.db.ExecContext(ctx,
		`INSERT INTO patients (user_id, name, email, phone, notes, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		p.UserID, p.Name, p.Email, p.Phone, p.Notes, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("create patient: %w", err)
	}
	return res.LastInsertId()
}

func (st *Store) UpdatePatient(ctx context.Context, p model.Patient) error {
	_, err := st.db.ExecContext(ctx,
		`UPDATE patients SET name = ?, email = ?, phone = ?, notes = ? WHERE id = ? AND user_id = ?`,
		p.Name, p.Email, p.Phone, p.Notes, p.ID, p.UserID)
	return err
}

// --- Historial del cliente ---

func (st *Store) PatientHistory(ctx context.Context, userID, patientID int64) ([]model.PatientHistoryEntry, error) {
	rows, err := st.db.QueryContext(ctx,
		`SELECT h.id, h.user_id, h.patient_id, h.author_id, COALESCE(u.name, ''),
			h.entry_type, h.content, h.created_at, h.updated_at
		 FROM patient_history h
		 LEFT JOIN users u ON u.id = h.author_id
		 WHERE h.user_id = ? AND h.patient_id = ?
		 ORDER BY h.created_at DESC, h.id DESC`, userID, patientID)
	if err != nil {
		return nil, fmt.Errorf("patient history: %w", err)
	}
	defer rows.Close()

	out := make([]model.PatientHistoryEntry, 0, 32)
	for rows.Next() {
		var e model.PatientHistoryEntry
		var authorID sql.NullInt64
		if err := rows.Scan(&e.ID, &e.UserID, &e.PatientID, &authorID, &e.AuthorName,
			&e.EntryType, &e.Content, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan patient history: %w", err)
		}
		if authorID.Valid {
			e.AuthorID = &authorID.Int64
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (st *Store) CreatePatientHistory(ctx context.Context, e model.PatientHistoryEntry) (int64, error) {
	now := time.Now().UTC()
	res, err := st.db.ExecContext(ctx,
		`INSERT INTO patient_history (user_id, patient_id, author_id, entry_type, content, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		e.UserID, e.PatientID, e.AuthorID, e.EntryType, e.Content, now, now)
	if err != nil {
		return 0, fmt.Errorf("create patient history: %w", err)
	}
	return res.LastInsertId()
}

func (st *Store) UpdatePatientHistory(ctx context.Context, e model.PatientHistoryEntry) error {
	res, err := st.db.ExecContext(ctx,
		`UPDATE patient_history SET entry_type = ?, content = ?, updated_at = ?
		 WHERE id = ? AND user_id = ? AND patient_id = ?`,
		e.EntryType, e.Content, time.Now().UTC(), e.ID, e.UserID, e.PatientID)
	if err != nil {
		return fmt.Errorf("update patient history: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (st *Store) DeletePatientHistory(ctx context.Context, userID, patientID, id int64) error {
	res, err := st.db.ExecContext(ctx,
		`DELETE FROM patient_history WHERE id = ? AND user_id = ? AND patient_id = ?`,
		id, userID, patientID)
	if err != nil {
		return fmt.Errorf("delete patient history: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// PatientServicesDone devuelve los servicios prestados al cliente en citas
// realizadas, facturadas o abonadas.
func (st *Store) PatientServicesDone(ctx context.Context, userID, patientID int64) ([]model.PatientServiceDone, error) {
	rows, err := st.db.QueryContext(ctx,
		`SELECT a.id, a.starts_at, a.status, COALESCE(s.name, ''), ai.duration_minutes, COALESCE(ai.price, 0)
		 FROM appointments a
		 JOIN appointment_items ai ON ai.appointment_id = a.id
		 LEFT JOIN services s ON s.id = ai.service_id
		 WHERE a.user_id = ? AND a.patient_id = ?
		   AND a.status IN ('realizada', 'facturada', 'abonada')
		 ORDER BY a.starts_at DESC, ai.sort_order, ai.id`, userID, patientID)
	if err != nil {
		return nil, fmt.Errorf("patient services: %w", err)
	}
	defer rows.Close()

	out := make([]model.PatientServiceDone, 0, 16)
	for rows.Next() {
		var s model.PatientServiceDone
		if err := rows.Scan(&s.AppointmentID, &s.StartsAt, &s.Status, &s.ServiceName,
			&s.DurationMinutes, &s.Price); err != nil {
			return nil, fmt.Errorf("scan patient service: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// --- Citas ---

const appointmentCols = `a.id, a.user_id, a.patient_id, p.name, a.assignee_id,
	COALESCE(asg.name, ''), a.starts_at, a.ends_at, a.status, a.notes,
	a.deposit_due, a.deposit_paid, a.deposit_payment_method, a.deposit_paid_at, a.created_at`

const appointmentFrom = ` FROM appointments a
	JOIN patients p ON p.id = a.patient_id
	LEFT JOIN users asg ON asg.id = a.assignee_id`

func scanAppointment(scan func(dest ...any) error) (model.Appointment, error) {
	var a model.Appointment
	var assigneeID sql.NullInt64
	var depositPaidAt sql.NullTime
	err := scan(&a.ID, &a.UserID, &a.PatientID, &a.PatientName, &assigneeID, &a.AssigneeName,
		&a.StartsAt, &a.EndsAt, &a.Status, &a.Notes,
		&a.DepositDue, &a.DepositPaid, &a.DepositMethod, &depositPaidAt, &a.CreatedAt)
	if err != nil {
		return a, err
	}
	if assigneeID.Valid {
		a.AssigneeID = &assigneeID.Int64
	}
	if depositPaidAt.Valid {
		a.DepositPaidAt = &depositPaidAt.Time
	}
	a.Services = []model.AppointmentService{}
	return a, nil
}

func (st *Store) AppointmentsRange(ctx context.Context, userID int64, start, end time.Time) ([]model.Appointment, error) {
	return st.appointmentsQuery(ctx,
		`WHERE a.user_id = ? AND a.starts_at >= ? AND a.starts_at < ?`,
		`ORDER BY a.starts_at ASC`,
		userID, start.UTC(), end.UTC())
}

func (st *Store) StaffAppointmentsRange(ctx context.Context, assigneeID int64, start, end time.Time) ([]model.Appointment, error) {
	return st.appointmentsQuery(ctx,
		`WHERE a.assignee_id = ? AND a.starts_at >= ? AND a.starts_at < ?`,
		`ORDER BY a.starts_at ASC`,
		assigneeID, start.UTC(), end.UTC())
}

func (st *Store) appointmentsQuery(ctx context.Context, where, order string, args ...any) ([]model.Appointment, error) {
	rows, err := st.db.QueryContext(ctx,
		`SELECT `+appointmentCols+appointmentFrom+` `+where+` `+order, args...)
	if err != nil {
		return nil, fmt.Errorf("appointments: %w", err)
	}
	defer rows.Close()

	out := make([]model.Appointment, 0, 32)
	for rows.Next() {
		a, err := scanAppointment(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("scan appointment: %w", err)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := st.attachItems(ctx, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (st *Store) attachItems(ctx context.Context, appts []model.Appointment) error {
	if len(appts) == 0 {
		return nil
	}
	ids := make([]any, 0, len(appts))
	placeholders := make([]string, 0, len(appts))
	for i := range appts {
		ids = append(ids, appts[i].ID)
		placeholders = append(placeholders, "?")
	}
	query := `SELECT ai.appointment_id, COALESCE(ai.service_id, 0), COALESCE(s.name, ''),
			ai.duration_minutes, COALESCE(ai.price, s.price, 0),
			CASE WHEN s.deposit_required = 1 THEN COALESCE(s.deposit_amount, 0) ELSE 0 END,
			COALESCE(s.color, ''), COALESCE(s.icon, '')
		FROM appointment_items ai
		LEFT JOIN services s ON s.id = ai.service_id
		WHERE ai.appointment_id IN (` + strings.Join(placeholders, ",") + `)
		ORDER BY ai.sort_order, ai.id`
	rows, err := st.db.QueryContext(ctx, query, ids...)
	if err != nil {
		return fmt.Errorf("appointment items: %w", err)
	}
	defer rows.Close()

	index := make(map[int64]int, len(appts))
	for i := range appts {
		index[appts[i].ID] = i
	}
	for rows.Next() {
		var apptID int64
		var it model.AppointmentService
		if err := rows.Scan(&apptID, &it.ServiceID, &it.Name, &it.DurationMinutes, &it.Price, &it.DepositAmount, &it.Color, &it.Icon); err != nil {
			return fmt.Errorf("scan item: %w", err)
		}
		if i, ok := index[apptID]; ok {
			appts[i].Services = append(appts[i].Services, it)
		}
	}
	return rows.Err()
}

func (st *Store) AppointmentByID(ctx context.Context, userID, id int64) (model.Appointment, error) {
	list, err := st.appointmentsQuery(ctx,
		`WHERE a.user_id = ? AND a.id = ?`,
		`LIMIT 1`, userID, id)
	if err != nil {
		return model.Appointment{}, err
	}
	if len(list) == 0 {
		return model.Appointment{}, sql.ErrNoRows
	}
	return list[0], nil
}

func (st *Store) CreateAppointment(ctx context.Context, a model.Appointment) (int64, error) {
	res, err := st.db.ExecContext(ctx,
		`INSERT INTO appointments (user_id, patient_id, assignee_id, starts_at, ends_at, status, notes, deposit_due, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.UserID, a.PatientID, a.AssigneeID, a.StartsAt.UTC(), a.EndsAt.UTC(), a.Status, a.Notes, a.DepositDue, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("create appointment: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := st.replaceItems(ctx, id, a.Services); err != nil {
		return 0, err
	}
	return id, nil
}

func (st *Store) UpdateAppointment(ctx context.Context, a model.Appointment) error {
	res, err := st.db.ExecContext(ctx,
		`UPDATE appointments SET assignee_id = ?, starts_at = ?, ends_at = ?, status = ?, notes = ?, deposit_due = ?
		 WHERE id = ? AND user_id = ?`,
		a.AssigneeID, a.StartsAt.UTC(), a.EndsAt.UTC(), a.Status, a.Notes, a.DepositDue, a.ID, a.UserID)
	if err != nil {
		return fmt.Errorf("update appointment: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		// Sin cambios en los campos: comprobar que la cita existe para no
		// impedir guardados que solo afectan a la señal o a los servicios.
		var exists int
		if err := st.db.QueryRowContext(ctx,
			`SELECT 1 FROM appointments WHERE id = ? AND user_id = ?`,
			a.ID, a.UserID).Scan(&exists); err != nil {
			return err
		}
	}
	return st.replaceItems(ctx, a.ID, a.Services)
}

func (st *Store) MoveAppointment(ctx context.Context, id, userID int64, assigneeID *int64, startsAt, endsAt time.Time) error {
	res, err := st.db.ExecContext(ctx,
		`UPDATE appointments SET assignee_id = ?, starts_at = ?, ends_at = ? WHERE id = ? AND user_id = ?`,
		assigneeID, startsAt.UTC(), endsAt.UTC(), id, userID)
	if err != nil {
		return fmt.Errorf("move appointment: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (st *Store) DeleteAppointment(ctx context.Context, id, userID int64) error {
	res, err := st.db.ExecContext(ctx,
		`DELETE FROM appointments WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return fmt.Errorf("delete appointment: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (st *Store) replaceItems(ctx context.Context, appointmentID int64, items []model.AppointmentService) error {
	if _, err := st.db.ExecContext(ctx,
		`DELETE FROM appointment_items WHERE appointment_id = ?`, appointmentID); err != nil {
		return fmt.Errorf("clear items: %w", err)
	}
	for i, it := range items {
		if it.ServiceID <= 0 {
			continue
		}
		if _, err := st.db.ExecContext(ctx,
			`INSERT INTO appointment_items (appointment_id, service_id, duration_minutes, price, sort_order) VALUES (?, ?, ?, ?, ?)`,
			appointmentID, it.ServiceID, it.DurationMinutes, it.Price, i); err != nil {
			return fmt.Errorf("insert item: %w", err)
		}
	}
	return nil
}

func (st *Store) PatientAppointments(ctx context.Context, userID, patientID int64, limit int) ([]model.Appointment, error) {
	if limit <= 0 {
		limit = 10
	}
	return st.appointmentsQuery(ctx,
		`WHERE a.user_id = ? AND a.patient_id = ?`,
		`ORDER BY a.starts_at DESC LIMIT ?`,
		userID, patientID, limit)
}

// --- Servicios ---

func (st *Store) Services(ctx context.Context, userID int64) ([]model.Service, error) {
	rows, err := st.db.QueryContext(ctx,
		`SELECT id, user_id, name, duration_minutes, price, deposit_required, deposit_amount, color, icon, sort_order, created_at
		 FROM services WHERE user_id = ? ORDER BY sort_order, name`, userID)
	if err != nil {
		return nil, fmt.Errorf("services: %w", err)
	}
	defer rows.Close()

	out := make([]model.Service, 0, 32)
	for rows.Next() {
		var s model.Service
		if err := rows.Scan(&s.ID, &s.UserID, &s.Name, &s.DurationMinutes, &s.Price, &s.DepositRequired, &s.DepositAmount, &s.Color, &s.Icon, &s.SortOrder, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan service: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (st *Store) ServiceByID(ctx context.Context, userID, id int64) (model.Service, error) {
	var s model.Service
	err := st.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, duration_minutes, price, deposit_required, deposit_amount, color, icon, sort_order, created_at
		 FROM services WHERE id = ? AND user_id = ?`, id, userID).
		Scan(&s.ID, &s.UserID, &s.Name, &s.DurationMinutes, &s.Price, &s.DepositRequired, &s.DepositAmount, &s.Color, &s.Icon, &s.SortOrder, &s.CreatedAt)
	return s, err
}

func (st *Store) CreateService(ctx context.Context, s model.Service) (int64, error) {
	res, err := st.db.ExecContext(ctx,
		`INSERT INTO services (user_id, name, duration_minutes, price, deposit_required, deposit_amount, color, icon, sort_order, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.UserID, s.Name, s.DurationMinutes, s.Price, s.DepositRequired, s.DepositAmount, s.Color, s.Icon, s.SortOrder, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("create service: %w", err)
	}
	return res.LastInsertId()
}

func (st *Store) UpdateService(ctx context.Context, s model.Service) error {
	res, err := st.db.ExecContext(ctx,
		`UPDATE services SET name = ?, duration_minutes = ?, price = ?, deposit_required = ?, deposit_amount = ?, color = ?, icon = ?
		 WHERE id = ? AND user_id = ?`,
		s.Name, s.DurationMinutes, s.Price, s.DepositRequired, s.DepositAmount, s.Color, s.Icon, s.ID, s.UserID)
	if err != nil {
		return fmt.Errorf("update service: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		var exists int
		err := st.db.QueryRowContext(ctx,
			`SELECT 1 FROM services WHERE id = ? AND user_id = ?`, s.ID, s.UserID).Scan(&exists)
		if err == sql.ErrNoRows {
			return sql.ErrNoRows
		}
		if err != nil {
			return fmt.Errorf("check service: %w", err)
		}
	}
	return nil
}

func (st *Store) DeleteService(ctx context.Context, userID, id int64) error {
	res, err := st.db.ExecContext(ctx,
		`DELETE FROM services WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return fmt.Errorf("delete service: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// --- Facturas ---

const invoiceCols = `i.id, i.user_id, i.patient_id, p.name, i.appointment_id, i.number, i.concept,
	i.subtotal, i.vat_rate, i.vat_amount, i.amount, i.status, i.issued_at, i.paid_at, i.payment_method, i.created_at`

const invoiceFrom = ` FROM invoices i JOIN patients p ON p.id = i.patient_id`

func scanInvoice(scan func(dest ...any) error) (model.Invoice, error) {
	var i model.Invoice
	var appointmentID sql.NullInt64
	err := scan(&i.ID, &i.UserID, &i.PatientID, &i.PatientName, &appointmentID, &i.Number, &i.Concept,
		&i.Subtotal, &i.VATRate, &i.VATAmount, &i.Amount, &i.Status, &i.IssuedAt, &i.PaidAt, &i.PaymentMethod, &i.CreatedAt)
	if err != nil {
		return i, err
	}
	if appointmentID.Valid {
		i.AppointmentID = &appointmentID.Int64
	}
	i.Items = []model.InvoiceItem{}
	return i, nil
}

func (st *Store) Invoices(ctx context.Context, userID int64, status string) ([]model.Invoice, error) {
	base := `SELECT ` + invoiceCols + invoiceFrom + ` WHERE i.user_id = ?`
	args := []any{userID}
	if s := strings.TrimSpace(status); s != "" {
		base += ` AND i.status = ?`
		args = append(args, s)
	}
	base += ` ORDER BY i.issued_at DESC, i.id DESC`

	rows, err := st.db.QueryContext(ctx, base, args...)
	if err != nil {
		return nil, fmt.Errorf("invoices: %w", err)
	}
	defer rows.Close()

	out := make([]model.Invoice, 0, 32)
	for rows.Next() {
		i, err := scanInvoice(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("scan invoice: %w", err)
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

func (st *Store) InvoiceByID(ctx context.Context, userID, id int64) (model.Invoice, error) {
	i, err := scanInvoice(st.db.QueryRowContext(ctx,
		`SELECT `+invoiceCols+invoiceFrom+` WHERE i.id = ? AND i.user_id = ?`, id, userID).Scan)
	if err != nil {
		return i, err
	}
	if i.Items, err = st.invoiceItems(ctx, i.ID); err != nil {
		return i, err
	}
	return i, nil
}

func (st *Store) InvoiceByAppointment(ctx context.Context, userID, appointmentID int64) (model.Invoice, error) {
	return scanInvoice(st.db.QueryRowContext(ctx,
		`SELECT `+invoiceCols+invoiceFrom+` WHERE i.user_id = ? AND i.appointment_id = ? ORDER BY i.id DESC LIMIT 1`,
		userID, appointmentID).Scan)
}

func (st *Store) invoiceItems(ctx context.Context, invoiceID int64) ([]model.InvoiceItem, error) {
	rows, err := st.db.QueryContext(ctx,
		`SELECT id, invoice_id, service_id, description, quantity, unit_price, vat_rate
		 FROM invoice_items WHERE invoice_id = ? ORDER BY sort_order, id`, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("invoice items: %w", err)
	}
	defer rows.Close()

	out := make([]model.InvoiceItem, 0, 8)
	for rows.Next() {
		var it model.InvoiceItem
		var serviceID sql.NullInt64
		if err := rows.Scan(&it.ID, &it.InvoiceID, &serviceID, &it.Description,
			&it.Quantity, &it.UnitPrice, &it.VATRate); err != nil {
			return nil, fmt.Errorf("scan invoice item: %w", err)
		}
		if serviceID.Valid {
			it.ServiceID = &serviceID.Int64
		}
		it.LineTotal = float64(it.Quantity) * it.UnitPrice
		out = append(out, it)
	}
	return out, rows.Err()
}

func (st *Store) CreateInvoice(ctx context.Context, i model.Invoice) (int64, error) {
	tx, err := st.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin invoice: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`INSERT INTO invoices (user_id, patient_id, appointment_id, number, concept, subtotal, vat_rate,
			vat_amount, amount, status, issued_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		i.UserID, i.PatientID, i.AppointmentID, i.Number, i.Concept, i.Subtotal, i.VATRate,
		i.VATAmount, i.Amount, i.Status, i.IssuedAt.UTC(), time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("create invoice: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	for n, it := range i.Items {
		qty := it.Quantity
		if qty <= 0 {
			qty = 1
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO invoice_items (invoice_id, service_id, description, quantity, unit_price, vat_rate, sort_order)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			id, it.ServiceID, it.Description, qty, it.UnitPrice, it.VATRate, n); err != nil {
			return 0, fmt.Errorf("insert invoice item: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit invoice: %w", err)
	}
	return id, nil
}

// UpdateInvoice reemplaza el concepto, importes y líneas de una factura pendiente.
func (st *Store) UpdateInvoice(ctx context.Context, i model.Invoice) error {
	tx, err := st.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin invoice update: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE invoices SET concept = ?, subtotal = ?, vat_rate = ?, vat_amount = ?, amount = ?
		 WHERE id = ? AND user_id = ? AND status = 'pendiente'`,
		i.Concept, i.Subtotal, i.VATRate, i.VATAmount, i.Amount, i.ID, i.UserID)
	if err != nil {
		return fmt.Errorf("update invoice: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM invoice_items WHERE invoice_id = ?`, i.ID); err != nil {
		return fmt.Errorf("clear invoice items: %w", err)
	}
	for n, it := range i.Items {
		qty := it.Quantity
		if qty <= 0 {
			qty = 1
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO invoice_items (invoice_id, service_id, description, quantity, unit_price, vat_rate, sort_order)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			i.ID, it.ServiceID, it.Description, qty, it.UnitPrice, it.VATRate, n); err != nil {
			return fmt.Errorf("insert invoice item: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit invoice update: %w", err)
	}
	return nil
}

// SetAppointmentStatus cambia el estado sin degradar una cita ya abonada (su
// cobro permanece registrado en caja).
func (st *Store) SetAppointmentStatus(ctx context.Context, userID, id int64, status string) error {
	_, err := st.db.ExecContext(ctx,
		`UPDATE appointments SET status = ? WHERE id = ? AND user_id = ? AND status <> 'abonada'`,
		status, id, userID)
	return err
}

// MarkInvoicePaid marca la factura como cobrada, deja constancia de la forma de
// pago, actualiza la cita a «abonada» y registra el ingreso en la caja del día.
// Devuelve la cita asociada (0 si la factura no nace de una cita).
func (st *Store) MarkInvoicePaid(ctx context.Context, userID, id int64, method string) (int64, error) {
	tx, err := st.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin invoice payment: %w", err)
	}
	defer tx.Rollback()

	var amount float64
	var number, status string
	var appointmentID sql.NullInt64
	if err := tx.QueryRowContext(ctx,
		`SELECT number, amount, status, appointment_id FROM invoices WHERE id = ? AND user_id = ? FOR UPDATE`,
		id, userID).Scan(&number, &amount, &status, &appointmentID); err != nil {
		return 0, err
	}
	if status == "pagada" {
		if appointmentID.Valid {
			return appointmentID.Int64, nil
		}
		return 0, nil
	}

	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx,
		`UPDATE invoices SET status = 'pagada', paid_at = ?, payment_method = ? WHERE id = ? AND user_id = ?`,
		now, method, id, userID); err != nil {
		return 0, fmt.Errorf("mark invoice paid: %w", err)
	}

	var linkedAppointment int64
	var depositPaid float64
	if appointmentID.Valid {
		linkedAppointment = appointmentID.Int64
		if _, err := tx.ExecContext(ctx,
			`UPDATE appointments SET status = 'abonada' WHERE id = ? AND user_id = ?`,
			linkedAppointment, userID); err != nil {
			return 0, fmt.Errorf("update appointment status: %w", err)
		}
		if err := tx.QueryRowContext(ctx,
			`SELECT deposit_paid FROM appointments WHERE id = ? AND user_id = ?`,
			linkedAppointment, userID).Scan(&depositPaid); err != nil {
			return 0, fmt.Errorf("read appointment deposit: %w", err)
		}
	}

	sessionID, err := ensureCashSessionTx(ctx, tx, userID, now)
	if err != nil {
		return 0, err
	}
	// La señal ya entró en caja al reservarse: aquí solo se registra lo que
	// falta por cobrar para no duplicar el ingreso.
	income := roundMoney(amount - depositPaid)
	if income > 0 {
		concept := "Cobro factura " + number
		if depositPaid > 0 {
			concept = "Cobro factura " + number + " (señal descontada)"
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO cash_movements (user_id, session_id, movement_type, payment_method, amount, concept, invoice_id, appointment_id, created_at)
			 VALUES (?, ?, 'ingreso', ?, ?, ?, ?, ?, ?)`,
			userID, sessionID, method, income, concept, id, nullableID(linkedAppointment), now); err != nil {
			return 0, fmt.Errorf("register cash income: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit invoice payment: %w", err)
	}
	return linkedAppointment, nil
}

// RegisterAppointmentDeposit cobra la señal pendiente de una cita: acumula el
// importe en la cita y lo registra como ingreso en la caja del día. Devuelve lo
// efectivamente cobrado (0 si la señal ya estaba completa).
func (st *Store) RegisterAppointmentDeposit(ctx context.Context, userID, appointmentID int64, amount float64, method string) (float64, error) {
	tx, err := st.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin deposit: %w", err)
	}
	defer tx.Rollback()

	var due, paid float64
	var patientName string
	if err := tx.QueryRowContext(ctx,
		`SELECT a.deposit_due, a.deposit_paid, p.name
		 FROM appointments a JOIN patients p ON p.id = a.patient_id
		 WHERE a.id = ? AND a.user_id = ? FOR UPDATE`,
		appointmentID, userID).Scan(&due, &paid, &patientName); err != nil {
		return 0, err
	}
	remaining := roundMoney(due - paid)
	if remaining <= 0 {
		return 0, nil
	}
	if amount <= 0 || amount > remaining {
		amount = remaining
	}
	amount = roundMoney(amount)

	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx,
		`UPDATE appointments SET deposit_paid = deposit_paid + ?, deposit_payment_method = ?, deposit_paid_at = ?
		 WHERE id = ? AND user_id = ?`,
		amount, method, now, appointmentID, userID); err != nil {
		return 0, fmt.Errorf("update appointment deposit: %w", err)
	}

	sessionID, err := ensureCashSessionTx(ctx, tx, userID, now)
	if err != nil {
		return 0, err
	}
	concept := "Señal cita · " + patientName
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO cash_movements (user_id, session_id, movement_type, payment_method, amount, concept, appointment_id, created_at)
		 VALUES (?, ?, 'ingreso', ?, ?, ?, ?, ?)`,
		userID, sessionID, method, amount, concept, appointmentID, now); err != nil {
		return 0, fmt.Errorf("register deposit income: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit deposit: %w", err)
	}
	return amount, nil
}

// ensureCashSessionTx devuelve la caja del día, creándola si no existe. Si estaba
// cerrada, la reabre para no perder el cobro recién registrado.
func ensureCashSessionTx(ctx context.Context, tx *sql.Tx, userID int64, now time.Time) (int64, error) {
	date := now.Format("2006-01-02")
	var id int64
	var status string
	err := tx.QueryRowContext(ctx,
		`SELECT id, status FROM cash_sessions WHERE user_id = ? AND session_date = ? FOR UPDATE`,
		userID, date).Scan(&id, &status)
	if err == nil {
		if status == "cerrada" {
			if _, err := tx.ExecContext(ctx,
				`UPDATE cash_sessions SET status = 'abierta', closed_at = NULL, difference = 0, updated_at = ? WHERE id = ?`,
				now, id); err != nil {
				return 0, fmt.Errorf("reopen cash session: %w", err)
			}
		}
		return id, nil
	}
	if !IsNotFound(err) {
		return 0, fmt.Errorf("find cash session: %w", err)
	}
	res, err := tx.ExecContext(ctx,
		`INSERT INTO cash_sessions (user_id, session_date, opening_amount, counted_amount, expected_cash,
		 difference, denoms_json, notes, status, opened_at, updated_at)
		 VALUES (?, ?, 0, 0, 0, 0, '[]', '', 'abierta', ?, ?)`,
		userID, date, now, now)
	if err != nil {
		return 0, fmt.Errorf("create cash session: %w", err)
	}
	return res.LastInsertId()
}

func nullableID(id int64) any {
	if id <= 0 {
		return nil
	}
	return id
}

func roundMoney(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}

func (st *Store) InvoiceCount(ctx context.Context, userID int64) (int, error) {
	var n int
	err := st.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM invoices WHERE user_id = ?`, userID).Scan(&n)
	return n, err
}

// IsDuplicate indica si el error es una violación de clave única (p. ej., una
// factura por cita que ya existe).
func IsDuplicate(err error) bool {
	var me *mysql.MySQLError
	return errors.As(err, &me) && me.Number == 1062
}

// --- Configuración de facturación ---

func (st *Store) BillingSettings(ctx context.Context, userID int64) (model.BillingSettings, error) {
	var b model.BillingSettings
	err := st.db.QueryRowContext(ctx,
		`SELECT user_id, company_name, tax_id, address, phone, email, vat_rate
		 FROM billing_settings WHERE user_id = ?`, userID).
		Scan(&b.UserID, &b.CompanyName, &b.TaxID, &b.Address, &b.Phone, &b.Email, &b.VATRate)
	if IsNotFound(err) {
		return model.BillingSettings{UserID: userID, VATRate: 21}, nil
	}
	if err != nil {
		return b, fmt.Errorf("billing settings: %w", err)
	}
	return b, nil
}

func (st *Store) SaveBillingSettings(ctx context.Context, b model.BillingSettings) error {
	_, err := st.db.ExecContext(ctx,
		`INSERT INTO billing_settings (user_id, company_name, tax_id, address, phone, email, vat_rate, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE company_name = VALUES(company_name), tax_id = VALUES(tax_id),
		 address = VALUES(address), phone = VALUES(phone), email = VALUES(email),
		 vat_rate = VALUES(vat_rate), updated_at = VALUES(updated_at)`,
		b.UserID, b.CompanyName, b.TaxID, b.Address, b.Phone, b.Email, b.VATRate, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("save billing settings: %w", err)
	}
	return nil
}

// --- Caja diaria ---

func (st *Store) CashSessionByDate(ctx context.Context, userID int64, date string) (model.CashSession, error) {
	var s model.CashSession
	var denoms string
	var closed sql.NullTime
	err := st.db.QueryRowContext(ctx,
		`SELECT id, user_id, DATE_FORMAT(session_date, '%Y-%m-%d'), opening_amount,
			counted_amount, expected_cash, difference, denoms_json, notes, status,
			opened_at, closed_at
		 FROM cash_sessions WHERE user_id = ? AND session_date = ?`, userID, date).
		Scan(&s.ID, &s.UserID, &s.Date, &s.OpeningAmount, &s.CountedAmount,
			&s.ExpectedCash, &s.Difference, &denoms, &s.Notes, &s.Status,
			&s.OpenedAt, &closed)
	if err != nil {
		return s, err
	}
	if closed.Valid {
		s.ClosedAt = &closed.Time
	}
	s.Denoms = decodeDenoms(denoms)
	return s, nil
}

func (st *Store) CashSessionByID(ctx context.Context, userID, id int64) (model.CashSession, error) {
	var s model.CashSession
	var denoms string
	var closed sql.NullTime
	err := st.db.QueryRowContext(ctx,
		`SELECT id, user_id, DATE_FORMAT(session_date, '%Y-%m-%d'), opening_amount,
			counted_amount, expected_cash, difference, denoms_json, notes, status,
			opened_at, closed_at
		 FROM cash_sessions WHERE user_id = ? AND id = ?`, userID, id).
		Scan(&s.ID, &s.UserID, &s.Date, &s.OpeningAmount, &s.CountedAmount,
			&s.ExpectedCash, &s.Difference, &denoms, &s.Notes, &s.Status,
			&s.OpenedAt, &closed)
	if err != nil {
		return s, err
	}
	if closed.Valid {
		s.ClosedAt = &closed.Time
	}
	s.Denoms = decodeDenoms(denoms)
	return s, nil
}

func (st *Store) CreateCashSession(ctx context.Context, userID int64, date string, opening float64) (int64, error) {
	now := time.Now().UTC()
	res, err := st.db.ExecContext(ctx,
		`INSERT INTO cash_sessions (user_id, session_date, opening_amount, counted_amount, expected_cash,
		 difference, denoms_json, notes, status, opened_at, updated_at)
		 VALUES (?, ?, ?, 0, 0, 0, '[]', '', 'abierta', ?, ?)`,
		userID, date, opening, now, now)
	if err != nil {
		return 0, fmt.Errorf("create cash session: %w", err)
	}
	return res.LastInsertId()
}

func (st *Store) CashMovements(ctx context.Context, userID, sessionID int64) ([]model.CashMovement, error) {
	rows, err := st.db.QueryContext(ctx,
		`SELECT id, user_id, session_id, movement_type, payment_method, amount,
			concept, invoice_id, appointment_id, created_at
		 FROM cash_movements WHERE user_id = ? AND session_id = ?
		 ORDER BY created_at ASC, id ASC`, userID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("cash movements: %w", err)
	}
	defer rows.Close()

	out := make([]model.CashMovement, 0, 32)
	for rows.Next() {
		var m model.CashMovement
		var invoiceID, appointmentID sql.NullInt64
		if err := rows.Scan(&m.ID, &m.UserID, &m.SessionID, &m.MovementType, &m.PaymentMethod,
			&m.Amount, &m.Concept, &invoiceID, &appointmentID, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan cash movement: %w", err)
		}
		if invoiceID.Valid {
			m.InvoiceID = &invoiceID.Int64
		}
		if appointmentID.Valid {
			m.AppointmentID = &appointmentID.Int64
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (st *Store) CreateCashMovement(ctx context.Context, m model.CashMovement) (int64, error) {
	res, err := st.db.ExecContext(ctx,
		`INSERT INTO cash_movements (user_id, session_id, movement_type, payment_method, amount,
			concept, invoice_id, appointment_id, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.UserID, m.SessionID, m.MovementType, m.PaymentMethod, m.Amount,
		m.Concept, m.InvoiceID, m.AppointmentID, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("create cash movement: %w", err)
	}
	return res.LastInsertId()
}

func (st *Store) DeleteCashMovement(ctx context.Context, userID, id int64) error {
	tx, err := st.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete movement: %w", err)
	}
	defer tx.Rollback()

	var amount float64
	var invoiceID, appointmentID sql.NullInt64
	if err := tx.QueryRowContext(ctx,
		`SELECT m.amount, m.invoice_id, m.appointment_id
		 FROM cash_movements m
		 JOIN cash_sessions s ON s.id = m.session_id
		 WHERE m.id = ? AND m.user_id = ? AND s.status = 'abierta' FOR UPDATE`,
		id, userID).Scan(&amount, &invoiceID, &appointmentID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM cash_movements WHERE id = ? AND user_id = ?`, id, userID); err != nil {
		return fmt.Errorf("delete cash movement: %w", err)
	}
	// Si se elimina el cobro de una señal, la cita vuelve a quedar pendiente.
	if appointmentID.Valid && !invoiceID.Valid {
		if _, err := tx.ExecContext(ctx,
			`UPDATE appointments SET deposit_paid = GREATEST(deposit_paid - ?, 0) WHERE id = ? AND user_id = ?`,
			amount, appointmentID.Int64, userID); err != nil {
			return fmt.Errorf("restore appointment deposit: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete movement: %w", err)
	}
	return nil
}

func (st *Store) CloseCashSession(ctx context.Context, s model.CashSession) error {
	payload, err := json.Marshal(s.Denoms)
	if err != nil {
		return fmt.Errorf("encode denominations: %w", err)
	}
	now := time.Now().UTC()
	res, err := st.db.ExecContext(ctx,
		`UPDATE cash_sessions SET counted_amount = ?, expected_cash = ?, difference = ?,
			denoms_json = ?, notes = ?, status = 'cerrada', closed_at = ?, updated_at = ?
		 WHERE id = ? AND user_id = ? AND status = 'abierta'`,
		s.CountedAmount, s.ExpectedCash, s.Difference, string(payload), s.Notes, now, now, s.ID, s.UserID)
	if err != nil {
		return fmt.Errorf("close cash session: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (st *Store) CashSessions(ctx context.Context, userID int64, limit int) ([]model.CashSession, error) {
	if limit <= 0 {
		limit = 30
	}
	rows, err := st.db.QueryContext(ctx,
		`SELECT s.id, s.user_id, DATE_FORMAT(s.session_date, '%Y-%m-%d'), s.opening_amount,
			s.counted_amount, s.expected_cash, s.difference, s.denoms_json, s.notes, s.status,
			s.opened_at, s.closed_at,
			COALESCE((SELECT SUM(amount) FROM cash_movements m WHERE m.session_id = s.id AND m.movement_type = 'ingreso'), 0),
			COALESCE((SELECT SUM(amount) FROM cash_movements m WHERE m.session_id = s.id AND m.movement_type = 'gasto'), 0)
		 FROM cash_sessions s WHERE s.user_id = ?
		 ORDER BY s.session_date DESC LIMIT ?`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("cash sessions: %w", err)
	}
	defer rows.Close()

	out := make([]model.CashSession, 0, limit)
	for rows.Next() {
		var s model.CashSession
		var denoms string
		var closed sql.NullTime
		if err := rows.Scan(&s.ID, &s.UserID, &s.Date, &s.OpeningAmount, &s.CountedAmount,
			&s.ExpectedCash, &s.Difference, &denoms, &s.Notes, &s.Status,
			&s.OpenedAt, &closed, &s.Income, &s.Expenses); err != nil {
			return nil, fmt.Errorf("scan cash session: %w", err)
		}
		if closed.Valid {
			s.ClosedAt = &closed.Time
		}
		s.Denoms = decodeDenoms(denoms)
		out = append(out, s)
	}
	return out, rows.Err()
}

func decodeDenoms(payload string) []model.CashDenomination {
	list := model.CashDenominations()
	if strings.TrimSpace(payload) == "" {
		return list
	}
	var saved []model.CashDenomination
	if err := json.Unmarshal([]byte(payload), &saved); err != nil {
		return list
	}
	counts := make(map[string]int, len(saved))
	for _, d := range saved {
		counts[d.Key] = d.Count
	}
	for i := range list {
		list[i].Count = counts[list[i].Key]
	}
	return list
}

// --- Dashboard ---

func (st *Store) AppStats(ctx context.Context, userID int64, start, end time.Time) (model.AppStats, error) {
	var s model.AppStats
	if err := st.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM patients WHERE user_id = ?`, userID).Scan(&s.Patients); err != nil {
		return s, err
	}
	if err := st.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM appointments WHERE user_id = ? AND starts_at >= ? AND starts_at < ?`,
		userID, start.UTC(), end.UTC()).Scan(&s.AppointmentsDay); err != nil {
		return s, err
	}
	if err := st.db.QueryRowContext(ctx,
		`SELECT COUNT(*), COALESCE(SUM(amount), 0) FROM invoices WHERE user_id = ? AND status = 'pendiente'`,
		userID).Scan(&s.PendingInvoices, &s.PendingAmount); err != nil {
		return s, err
	}
	if err := st.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM invoices WHERE user_id = ? AND status = 'pagada'`,
		userID).Scan(&s.CollectedAmount); err != nil {
		return s, err
	}
	var opening float64
	var sessionID int64
	err := st.db.QueryRowContext(ctx,
		`SELECT id, opening_amount FROM cash_sessions WHERE user_id = ? AND session_date = ?`,
		userID, start.UTC().Format("2006-01-02")).Scan(&sessionID, &opening)
	if err == nil {
		s.CashToday = true
		var net float64
		if err := st.db.QueryRowContext(ctx,
			`SELECT COALESCE(SUM(CASE WHEN movement_type = 'gasto' THEN -amount ELSE amount END), 0)
			 FROM cash_movements WHERE session_id = ? AND payment_method = 'efectivo'`, sessionID).Scan(&net); err != nil {
			return s, err
		}
		s.CashExpected = opening + net
	} else if !IsNotFound(err) {
		return s, err
	}
	return s, nil
}

// --- Preferencias de usuario ---

func (st *Store) Preferences(ctx context.Context, userID int64) (model.Theme, error) {
	var preset, payload string
	err := st.db.QueryRowContext(ctx,
		`SELECT theme_preset, theme_json FROM user_preferences WHERE user_id = ?`, userID).
		Scan(&preset, &payload)
	if IsNotFound(err) {
		return model.DefaultTheme(), nil
	}
	if err != nil {
		return model.DefaultTheme(), fmt.Errorf("preferences: %w", err)
	}
	var theme model.Theme
	if err := json.Unmarshal([]byte(payload), &theme); err != nil {
		return model.DefaultTheme(), nil
	}
	if theme.Preset == "" {
		theme.Preset = preset
	}
	return theme.Normalize(), nil
}

func (st *Store) SavePreferences(ctx context.Context, userID int64, theme model.Theme) error {
	theme = theme.Normalize()
	payload, err := json.Marshal(theme)
	if err != nil {
		return fmt.Errorf("encode preferences: %w", err)
	}
	_, err = st.db.ExecContext(ctx,
		`INSERT INTO user_preferences (user_id, theme_preset, theme_json, updated_at)
		 VALUES (?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE theme_preset = VALUES(theme_preset),
		 theme_json = VALUES(theme_json), updated_at = VALUES(updated_at)`,
		userID, theme.Preset, string(payload), time.Now().UTC())
	if err != nil {
		return fmt.Errorf("save preferences: %w", err)
	}
	return nil
}

// --- Horario de la clínica ---

func (st *Store) ClinicSchedule(ctx context.Context, userID int64) ([]model.ScheduleDay, error) {
	var payload string
	err := st.db.QueryRowContext(ctx,
		`SELECT schedule_json FROM clinic_schedule WHERE user_id = ?`, userID).Scan(&payload)
	if IsNotFound(err) {
		return model.DefaultSchedule(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("clinic schedule: %w", err)
	}
	days := make([]model.ScheduleDay, 0, 7)
	if err := json.Unmarshal([]byte(payload), &days); err != nil {
		return model.DefaultSchedule(), nil
	}
	if len(days) != 7 {
		return model.DefaultSchedule(), nil
	}
	return days, nil
}

func (st *Store) SaveClinicSchedule(ctx context.Context, userID int64, days []model.ScheduleDay) error {
	payload, err := json.Marshal(days)
	if err != nil {
		return fmt.Errorf("encode schedule: %w", err)
	}
	_, err = st.db.ExecContext(ctx,
		`INSERT INTO clinic_schedule (user_id, schedule_json, updated_at) VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE schedule_json = VALUES(schedule_json), updated_at = VALUES(updated_at)`,
		userID, string(payload), time.Now().UTC())
	if err != nil {
		return fmt.Errorf("save clinic schedule: %w", err)
	}
	return nil
}

// --- Bloques cerrados por empleado ---

func (st *Store) ScheduleBlocks(ctx context.Context, ownerID int64) ([]model.ScheduleBlock, error) {
	rows, err := st.db.QueryContext(ctx,
		`SELECT b.id, b.series_id, b.user_id, b.staff_id, COALESCE(u.name, ''), b.weekday,
			DATE_FORMAT(b.start_time, '%H:%i'), DATE_FORMAT(b.end_time, '%H:%i'), b.reason
		 FROM schedule_blocks b
		 LEFT JOIN users u ON u.id = b.staff_id
		 WHERE b.user_id = ?
		 ORDER BY b.weekday, b.start_time`, ownerID)
	if err != nil {
		return nil, fmt.Errorf("schedule blocks: %w", err)
	}
	defer rows.Close()

	out := make([]model.ScheduleBlock, 0, 16)
	for rows.Next() {
		var b model.ScheduleBlock
		if err := rows.Scan(&b.ID, &b.SeriesID, &b.UserID, &b.StaffID, &b.StaffName, &b.Weekday, &b.Start, &b.End, &b.Reason); err != nil {
			return nil, fmt.Errorf("scan schedule block: %w", err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (st *Store) CreateScheduleBlock(ctx context.Context, b model.ScheduleBlock) (int64, error) {
	res, err := st.db.ExecContext(ctx,
		`INSERT INTO schedule_blocks (series_id, user_id, staff_id, weekday, start_time, end_time, reason) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		b.SeriesID, b.UserID, b.StaffID, b.Weekday, b.Start, b.End, b.Reason)
	if err != nil {
		return 0, fmt.Errorf("create schedule block: %w", err)
	}
	return res.LastInsertId()
}

func (st *Store) UpdateScheduleBlock(ctx context.Context, b model.ScheduleBlock) error {
	_, err := st.db.ExecContext(ctx,
		`UPDATE schedule_blocks SET series_id = ?, staff_id = ?, weekday = ?, start_time = ?, end_time = ?, reason = ?
		 WHERE id = ? AND user_id = ?`,
		b.SeriesID, b.StaffID, b.Weekday, b.Start, b.End, b.Reason, b.ID, b.UserID)
	if err != nil {
		return fmt.Errorf("update schedule block: %w", err)
	}
	return nil
}

func (st *Store) UpdateScheduleSeries(ctx context.Context, ownerID, seriesID int64, b model.ScheduleBlock) error {
	_, err := st.db.ExecContext(ctx,
		`UPDATE schedule_blocks SET staff_id = ?, start_time = ?, end_time = ?, reason = ?
		 WHERE series_id = ? AND user_id = ?`,
		b.StaffID, b.Start, b.End, b.Reason, seriesID, ownerID)
	if err != nil {
		return fmt.Errorf("update schedule series: %w", err)
	}
	return nil
}

// --- Festivos ---

func (st *Store) Holidays(ctx context.Context, ownerID int64) ([]model.Holiday, error) {
	rows, err := st.db.QueryContext(ctx,
		`SELECT id, user_id, DATE_FORMAT(holiday_date, '%Y-%m-%d'), name
		 FROM holidays WHERE user_id = ? ORDER BY holiday_date`, ownerID)
	if err != nil {
		return nil, fmt.Errorf("holidays: %w", err)
	}
	defer rows.Close()

	out := make([]model.Holiday, 0, 8)
	for rows.Next() {
		var h model.Holiday
		if err := rows.Scan(&h.ID, &h.UserID, &h.Date, &h.Name); err != nil {
			return nil, fmt.Errorf("scan holiday: %w", err)
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func (st *Store) CreateHoliday(ctx context.Context, h model.Holiday) error {
	_, err := st.db.ExecContext(ctx,
		`INSERT INTO holidays (user_id, holiday_date, name) VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE name = VALUES(name)`,
		h.UserID, h.Date, h.Name)
	if err != nil {
		return fmt.Errorf("create holiday: %w", err)
	}
	return nil
}

func (st *Store) DeleteHoliday(ctx context.Context, ownerID, id int64) error {
	res, err := st.db.ExecContext(ctx,
		`DELETE FROM holidays WHERE id = ? AND user_id = ?`, id, ownerID)
	if err != nil {
		return fmt.Errorf("delete holiday: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (st *Store) DeleteScheduleBlock(ctx context.Context, ownerID, id int64) error {
	res, err := st.db.ExecContext(ctx,
		`DELETE FROM schedule_blocks WHERE id = ? AND user_id = ?`, id, ownerID)
	if err != nil {
		return fmt.Errorf("delete schedule block: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
