-- 005_agenda.sql — Agenda por empleado: roles, asignación de citas y servicios múltiples.

-- 1. Roles y jerarquía de usuarios (gestor -> empleados)
ALTER TABLE users
  ADD COLUMN role VARCHAR(16) NOT NULL DEFAULT 'staff' AFTER password_hash,
  ADD COLUMN manager_id BIGINT UNSIGNED NULL AFTER role,
  ADD KEY idx_users_manager (manager_id),
  ADD CONSTRAINT fk_users_manager FOREIGN KEY (manager_id) REFERENCES users(id) ON DELETE SET NULL;

-- 2. Asignación de cada cita a un empleado
ALTER TABLE appointments
  ADD COLUMN assignee_id BIGINT UNSIGNED NULL AFTER patient_id,
  ADD KEY idx_appointments_assignee (assignee_id),
  ADD CONSTRAINT fk_appointments_assignee FOREIGN KEY (assignee_id) REFERENCES users(id) ON DELETE SET NULL;

-- 3. Líneas de servicio por cita
CREATE TABLE IF NOT EXISTS appointment_items (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  appointment_id BIGINT UNSIGNED NOT NULL,
  service VARCHAR(120) NOT NULL,
  sort_order INT NOT NULL DEFAULT 0,
  KEY idx_items_appointment (appointment_id),
  CONSTRAINT fk_items_appointment FOREIGN KEY (appointment_id) REFERENCES appointments(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 4. Migrar el servicio único existente a una línea de servicio
INSERT INTO appointment_items (appointment_id, service, sort_order)
SELECT id, service, 0 FROM appointments WHERE service <> '';

-- 5. demo pasa a gestor
UPDATE users SET role = 'manager' WHERE email = 'demo@klowna.com';

-- 6. Tres empleados asignables (contraseña: klowna1234)
INSERT INTO users (email, name, password_hash, role, manager_id, created_at) VALUES
  ('ana@klowna.com',    'Ana Beltrán',    '$2a$10$zIi7MZ/zp2oF0Qb.vAth.uy0BwCjxMx4klmeKAFcjmjJRs41ctw6q', 'staff', (SELECT id FROM users WHERE email = 'demo@klowna.com'), NOW()),
  ('carlos@klowna.com', 'Carlos Mendoza', '$2a$10$zIi7MZ/zp2oF0Qb.vAth.uy0BwCjxMx4klmeKAFcjmjJRs41ctw6q', 'staff', (SELECT id FROM users WHERE email = 'demo@klowna.com'), NOW()),
  ('lucia@klowna.com',  'Lucía Ferrer',   '$2a$10$zIi7MZ/zp2oF0Qb.vAth.uy0BwCjxMx4klmeKAFcjmjJRs41ctw6q', 'staff', (SELECT id FROM users WHERE email = 'demo@klowna.com'), NOW());

-- 7. Eliminar la columna de servicio único
ALTER TABLE appointments DROP COLUMN service;
