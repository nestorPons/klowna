-- 017_holidays.sql — Festivos del centro (fechas concretas cerradas).

CREATE TABLE IF NOT EXISTS holidays (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  holiday_date DATE NOT NULL,
  name VARCHAR(120) NOT NULL DEFAULT '',
  KEY idx_holidays_user (user_id),
  UNIQUE KEY uniq_holidays_user_date (user_id, holiday_date),
  CONSTRAINT fk_holidays_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
