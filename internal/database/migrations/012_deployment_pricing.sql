-- 012_deployment_pricing.sql — Precio de despliegue: 25 € en servidor propio, desde 60 € en servidor del cliente.

UPDATE plans SET sort_order = sort_order + 1 WHERE sort_order >= 3;

INSERT INTO plans (name, price, period, description, features_json, highlighted, sort_order) VALUES
  ('Despliegue', 'Desde 25 €', 'pago único', 'Instalamos la aplicación en mi servidor o en el tuyo, con dominio y SSL.',
   '["Instalación, base de datos y migraciones","Dominio, certificado SSL y puesta en marcha","En mi servidor: 25 €","En tu servidor: desde 60 € (panel) o 90 € (VPS/SSH)","Pruebas y entrega documentada"]', 0, 3);

INSERT INTO faqs (question, answer, sort_order) VALUES
  ('¿Cuánto cuesta el despliegue?',
   'El despliegue en mi servidor cuesta 25 € (pago único). En tu servidor parte de 60 € si tienes panel de control y de 90 € si es un VPS con SSH y Docker. El precio final se cierra antes de empezar.', 7),
  ('¿Qué influye en el precio del despliegue en mi servidor?',
   'El tipo de acceso (panel, FTP o SSH), el sistema de la máquina, la configuración de dominio, DNS y SSL, la migración de datos existentes y las copias de seguridad. Los extras se presupuestan aparte.', 8);
