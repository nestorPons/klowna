-- 011_user_preferences.sql — Preferencias de tema por usuario.

CREATE TABLE IF NOT EXISTS user_preferences (
  user_id BIGINT UNSIGNED NOT NULL PRIMARY KEY,
  theme_preset VARCHAR(32) NOT NULL DEFAULT 'bordeaux',
  theme_json TEXT NOT NULL,
  updated_at DATETIME NOT NULL,
  CONSTRAINT fk_user_preferences_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
