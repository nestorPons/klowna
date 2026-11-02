-- 014_services_style.sql — Color e icono por servicio.

ALTER TABLE services
  ADD COLUMN color CHAR(7) NOT NULL DEFAULT '' AFTER price,
  ADD COLUMN icon VARCHAR(32) NOT NULL DEFAULT '' AFTER color;
