-- 006_services.sql — Catálogo de servicios con duración y asignación a citas.

-- 1. Catálogo de servicios por clínica
CREATE TABLE IF NOT EXISTS services (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  name VARCHAR(120) NOT NULL,
  duration_minutes INT NOT NULL DEFAULT 30,
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL,
  KEY idx_services_user (user_id),
  CONSTRAINT fk_services_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. Migrar las líneas de servicio libres a un catálogo (una entrada por nombre)
INSERT INTO services (user_id, name, duration_minutes, sort_order, created_at)
SELECT DISTINCT a.user_id, ai.service, 30, 0, NOW()
FROM appointment_items ai
JOIN appointments a ON a.id = ai.appointment_id
WHERE ai.service <> '';

-- 3. Las líneas de cita pasan a referenciar el catálogo y guardan su duración
ALTER TABLE appointment_items
  ADD COLUMN service_id BIGINT UNSIGNED NULL AFTER appointment_id,
  ADD COLUMN duration_minutes INT NOT NULL DEFAULT 30 AFTER service_id,
  ADD KEY idx_items_service (service_id),
  ADD CONSTRAINT fk_items_service FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE SET NULL;

UPDATE appointment_items ai
JOIN appointments a ON a.id = ai.appointment_id
JOIN services s ON s.user_id = a.user_id AND s.name = ai.service
SET ai.service_id = s.id, ai.duration_minutes = s.duration_minutes;

ALTER TABLE appointment_items DROP COLUMN service;
