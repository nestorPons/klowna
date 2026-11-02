-- 016_worker_style.sql — Icono y color por trabajador (equipo).

ALTER TABLE users
  ADD COLUMN color CHAR(7) NOT NULL DEFAULT '' AFTER name,
  ADD COLUMN icon VARCHAR(32) NOT NULL DEFAULT '' AFTER color;
