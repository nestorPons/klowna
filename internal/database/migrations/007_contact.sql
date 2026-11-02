-- 007_contact.sql — Datos oficiales de contacto y identidad del sitio.

INSERT INTO site_settings (setting_key, setting_value) VALUES
  ('site_name', 'Klowna'),
  ('tagline', 'Agenda electrónica y software de gestión de citas a medida'),
  ('phone', '+34 660 291 797'),
  ('email', 'nestorpons@gmail.com'),
  ('whatsapp', '34660291797'),
  ('address', 'España · Atención online'),
  ('legal_note', 'Software de agenda electrónica y gestión de citas a medida.')
ON DUPLICATE KEY UPDATE setting_value = VALUES(setting_value);
