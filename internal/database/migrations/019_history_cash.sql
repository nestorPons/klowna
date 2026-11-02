-- 019_history_cash.sql — Historial de clientes, caja diaria y forma de pago.

-- 1. Forma de pago de las facturas cobradas.
ALTER TABLE invoices
  ADD COLUMN payment_method VARCHAR(24) NOT NULL DEFAULT '' AFTER paid_at;

-- 2. Estados de cita: pendiente, confirmada, realizada, no_asistida, facturada y abonada.
--    El estado se guarda como texto para poder evolucionar sin migraciones.
ALTER TABLE appointments
  ADD KEY idx_appointments_status (user_id, status);

-- 3. Historial de cliente: notas, comentarios y seguimientos editables.
CREATE TABLE IF NOT EXISTS patient_history (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  patient_id BIGINT UNSIGNED NOT NULL,
  author_id BIGINT UNSIGNED NULL,
  entry_type VARCHAR(24) NOT NULL DEFAULT 'nota',
  content TEXT NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  KEY idx_history_patient (user_id, patient_id, created_at),
  KEY idx_history_author (author_id),
  CONSTRAINT fk_history_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_history_patient FOREIGN KEY (patient_id) REFERENCES patients(id) ON DELETE CASCADE,
  CONSTRAINT fk_history_author FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 4. Caja diaria: una sesión por día y gestor.
CREATE TABLE IF NOT EXISTS cash_sessions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  session_date DATE NOT NULL,
  opening_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  counted_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  expected_cash DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  difference DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  denoms_json TEXT NOT NULL,
  notes VARCHAR(500) NOT NULL DEFAULT '',
  status VARCHAR(16) NOT NULL DEFAULT 'abierta',
  opened_at DATETIME NOT NULL,
  closed_at DATETIME NULL,
  updated_at DATETIME NOT NULL,
  UNIQUE KEY uq_cash_user_date (user_id, session_date),
  KEY idx_cash_user_date (user_id, session_date),
  CONSTRAINT fk_cash_session_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 5. Movimientos de caja: cobros y gastos, con vínculo opcional a factura/cita.
CREATE TABLE IF NOT EXISTS cash_movements (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  session_id BIGINT UNSIGNED NOT NULL,
  movement_type VARCHAR(16) NOT NULL DEFAULT 'ingreso',
  payment_method VARCHAR(24) NOT NULL DEFAULT 'efectivo',
  amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  concept VARCHAR(255) NOT NULL DEFAULT '',
  invoice_id BIGINT UNSIGNED NULL,
  appointment_id BIGINT UNSIGNED NULL,
  created_at DATETIME NOT NULL,
  KEY idx_cash_mov_session (session_id, created_at),
  KEY idx_cash_mov_user (user_id, created_at),
  KEY idx_cash_mov_invoice (invoice_id),
  CONSTRAINT fk_cash_mov_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_cash_mov_session FOREIGN KEY (session_id) REFERENCES cash_sessions(id) ON DELETE CASCADE,
  CONSTRAINT fk_cash_mov_invoice FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE SET NULL,
  CONSTRAINT fk_cash_mov_appointment FOREIGN KEY (appointment_id) REFERENCES appointments(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
