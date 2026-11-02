-- 018_schedule_block_series.sql — Agrupa bloques creados para todos los días.

ALTER TABLE schedule_blocks
  ADD COLUMN series_id BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER id;
