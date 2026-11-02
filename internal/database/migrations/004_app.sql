-- 004_app.sql — Esquema del SaaS de gestión (autenticación, agenda, pacientes y facturación).

CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  email VARCHAR(160) NOT NULL UNIQUE,
  name VARCHAR(120) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  created_at DATETIME NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS sessions (
  token_hash CHAR(64) NOT NULL PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  expires_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL,
  KEY idx_sessions_user (user_id),
  CONSTRAINT fk_sessions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS patients (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  name VARCHAR(120) NOT NULL,
  email VARCHAR(160) NOT NULL DEFAULT '',
  phone VARCHAR(40) NOT NULL DEFAULT '',
  notes TEXT NOT NULL,
  created_at DATETIME NOT NULL,
  KEY idx_patients_user (user_id),
  CONSTRAINT fk_patients_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS appointments (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  patient_id BIGINT UNSIGNED NOT NULL,
  service VARCHAR(120) NOT NULL DEFAULT '',
  starts_at DATETIME NOT NULL,
  ends_at DATETIME NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pendiente',
  notes TEXT NOT NULL,
  created_at DATETIME NOT NULL,
  KEY idx_appointments_user_start (user_id, starts_at),
  CONSTRAINT fk_appointments_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_appointments_patient FOREIGN KEY (patient_id) REFERENCES patients(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS invoices (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  patient_id BIGINT UNSIGNED NOT NULL,
  number VARCHAR(32) NOT NULL,
  concept VARCHAR(255) NOT NULL DEFAULT '',
  amount DECIMAL(10,2) NOT NULL DEFAULT 0,
  status VARCHAR(32) NOT NULL DEFAULT 'pendiente',
  issued_at DATETIME NOT NULL,
  paid_at DATETIME NULL,
  created_at DATETIME NOT NULL,
  KEY idx_invoices_user (user_id),
  CONSTRAINT fk_invoices_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_invoices_patient FOREIGN KEY (patient_id) REFERENCES patients(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Usuario demo: demo@klowna.com / klowna1234
INSERT INTO users (email, name, password_hash, created_at) VALUES
  ('demo@klowna.com', 'Clínica Demo', '$2a$10$zIi7MZ/zp2oF0Qb.vAth.uy0BwCjxMx4klmeKAFcjmjJRs41ctw6q', NOW());

INSERT INTO patients (user_id, name, email, phone, notes, created_at) VALUES
  ((SELECT id FROM users WHERE email = 'demo@klowna.com'), 'Laura García', 'laura@ejemplo.com', '+34 600 111 222', 'Paciente de medicina estética. Piel sensible.', NOW()),
  ((SELECT id FROM users WHERE email = 'demo@klowna.com'), 'Marta Ruiz', 'marta@ejemplo.com', '+34 600 333 444', 'Seguimiento capilar mensual.', NOW()),
  ((SELECT id FROM users WHERE email = 'demo@klowna.com'), 'Javier Sanz', 'javier@ejemplo.com', '+34 600 555 666', 'Prefiere cita a primera hora.', NOW());

INSERT INTO appointments (user_id, patient_id, service, starts_at, ends_at, status, notes, created_at) VALUES
  ((SELECT id FROM users WHERE email = 'demo@klowna.com'), (SELECT id FROM patients WHERE name = 'Laura García'), 'Consulta láser', CONCAT(CURDATE(), ' 10:00:00'), CONCAT(CURDATE(), ' 10:30:00'), 'confirmada', 'Cabina 2', NOW()),
  ((SELECT id FROM users WHERE email = 'demo@klowna.com'), (SELECT id FROM patients WHERE name = 'Marta Ruiz'), 'Revisión capilar', CONCAT(CURDATE(), ' 11:30:00'), CONCAT(CURDATE(), ' 12:00:00'), 'pendiente', 'Consulta 1', NOW()),
  ((SELECT id FROM users WHERE email = 'demo@klowna.com'), (SELECT id FROM patients WHERE name = 'Javier Sanz'), 'Primera visita', CONCAT(CURDATE(), ' 17:15:00'), CONCAT(CURDATE(), ' 18:00:00'), 'pendiente', 'Documentación inicial', NOW());

INSERT INTO invoices (user_id, patient_id, number, concept, amount, status, issued_at, paid_at, created_at) VALUES
  ((SELECT id FROM users WHERE email = 'demo@klowna.com'), (SELECT id FROM patients WHERE name = 'Laura García'), 'F-2026-0001', 'Tratamiento láser (3 sesiones)', 240.00, 'pagada', CONCAT(CURDATE(), ' 09:00:00'), CONCAT(CURDATE(), ' 09:05:00'), NOW()),
  ((SELECT id FROM users WHERE email = 'demo@klowna.com'), (SELECT id FROM patients WHERE name = 'Marta Ruiz'), 'F-2026-0002', 'Consulta y diagnóstico capilar', 90.00, 'pendiente', CONCAT(CURDATE(), ' 09:00:00'), NULL, NOW()),
  ((SELECT id FROM users WHERE email = 'demo@klowna.com'), (SELECT id FROM patients WHERE name = 'Javier Sanz'), 'F-2026-0003', 'Bono 5 sesiones de mantenimiento', 320.00, 'pendiente', CONCAT(CURDATE(), ' 09:00:00'), NULL, NOW());
