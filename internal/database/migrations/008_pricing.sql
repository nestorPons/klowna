-- 008_pricing.sql — Precios reales: app gratuita, desarrollo, alojamiento y soporte.

DELETE FROM plans;
DELETE FROM faqs;

INSERT INTO plans (name, price, period, description, features_json, highlighted, sort_order) VALUES
  ('App gratuita', '0 €', 'gratis', 'La aplicación tal y como está, sin coste y sin permanencia.',
   '["Agenda electrónica","Gestión de pacientes","Facturación","Servicios y profesionales","Acceso desde móvil y ordenador"]', 1, 1),
  ('Desarrollo a medida', 'Desde 45 €', '', 'Adaptamos la aplicación a los procesos de tu negocio.',
   '["Análisis de tus procesos","Funciones personalizadas","Integraciones con otras herramientas","Presupuesto cerrado"]', 0, 2),
  ('Alojamiento', '10 €', '/mes · anual', 'En mis servidores. Contratación anual.',
   '["Servidor gestionado","Copias de seguridad","Certificado SSL","Actualizaciones incluidas"]', 0, 3),
  ('Servicio técnico', '5 €', '/mes · anual', 'Soporte y mantenimiento. Contratación anual.',
   '["Soporte por WhatsApp y email","Resolución de incidencias","Mantenimiento de la aplicación","Contratación anual"]', 0, 4);

INSERT INTO faqs (question, answer, sort_order) VALUES
  ('¿Cuánto cuesta la aplicación?',
   'La aplicación tal y como está es gratis. Puedes usarla sin coste y sin permanencia.', 1),
  ('¿Qué incluye la aplicación gratuita?',
   'Incluye agenda electrónica, gestión de pacientes, facturación, servicios y acceso desde móvil y ordenador.', 2),
  ('¿Cuánto cuesta un desarrollo a medida?',
   'Los desarrollos a medida parten de 45 €. El precio final depende del alcance y se cierra con un presupuesto sin sorpresas.', 3),
  ('¿Cuánto cuesta el alojamiento?',
   'El alojamiento en mis servidores cuesta 10 €/mes con contratación anual.', 4),
  ('¿Cuánto cuesta el servicio técnico?',
   'El servicio técnico cuesta 5 €/mes con contratación anual.', 5),
  ('¿Hay permanencia?',
   'La aplicación gratuita no tiene permanencia. El alojamiento y el servicio técnico se contratan de forma anual.', 6);
