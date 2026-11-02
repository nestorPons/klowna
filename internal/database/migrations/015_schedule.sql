-- 015_schedule.sql — Horario de la clínica y bloques cerrados por empleado.

CREATE TABLE IF NOT EXISTS clinic_schedule (
  user_id BIGINT UNSIGNED NOT NULL PRIMARY KEY,
  schedule_json TEXT NOT NULL,
  updated_at DATETIME NOT NULL,
  CONSTRAINT fk_clinic_schedule_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS schedule_blocks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  staff_id BIGINT UNSIGNED NOT NULL,
  weekday INT NOT NULL,
  start_time TIME NOT NULL,
  end_time TIME NOT NULL,
  reason VARCHAR(120) NOT NULL DEFAULT '',
  KEY idx_blocks_user (user_id),
  KEY idx_blocks_staff (staff_id),
  CONSTRAINT fk_schedule_blocks_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_schedule_blocks_staff FOREIGN KEY (staff_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
