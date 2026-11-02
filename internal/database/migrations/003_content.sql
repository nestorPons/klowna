-- 003_content.sql — Contenido real de flowww.es adaptado a Klowna
-- Reemplaza precios y FAQ por las licencias y preguntas reales.

DELETE FROM plans;
DELETE FROM faqs;

INSERT INTO plans (name, price, period, description, features_json, highlighted, sort_order) VALUES
  ('Boss', '59 €', '/mes', 'Gestión completa de tu clínica. 590 €/año.',
   '["Gestión de clientes","Agenda inteligente","Facturación y ventas","Stock de productos","Campañas de marketing","Control de personal","Análisis de datos","Almacenamiento: 100 GB"]', 0, 1),
  ('Star', '119 €', '/mes', 'Haz crecer tu negocio con herramientas avanzadas. 1190 €/año.',
   '["Todo lo de Boss","Almacenamiento: 250 GB","Anticipos","Workflows","Presupuestos"]', 1, 2),
  ('Ultimate', '249 €', '/mes', 'Personaliza Klowna al máximo y expande sus posibilidades. 2490 €/año.',
   '["Todo lo de Star","Almacenamiento: 500 GB","Workflows en tiempo real","API","Acceso a SQL","Integración con HubSpot"]', 0, 3);

INSERT INTO faqs (question, answer, sort_order) VALUES
  ('¿Cómo contrato Klowna?',
   'Tienes dos vías: pide una consultoría personalizada en la que analizamos tu negocio y te recomendamos la solución que mejor encaja, o prueba Klowna gratis durante 10 días. Un consultor te acompañará durante todo el proceso de contratación.', 1),
  ('¿Cuál es el precio de Klowna?',
   'Ofrecemos 3 licencias: Boss por 59 €/mes, Star por 119 €/mes y Ultimate por 249 €/mes. Además puedes sumar herramientas a tu SaaS para potenciar el crecimiento de tu negocio.', 2),
  ('Tengo más de una clínica. ¿Cuántas licencias debo contratar?',
   'Deberás contratar una licencia para cada clínica. Según tus necesidades y las herramientas que uses a diario, podrás elegir una versión u otra.', 3),
  ('Si contrato Klowna, ¿qué cambios debo aplicar en mi negocio?',
   'Ninguno. Un consultor especializado revisará tu operativa diaria, entenderá tus procesos y verá cómo Klowna puede ayudarte a optimizarlos. Buscaremos juntos la mejor solución para ti.', 4),
  ('Ya tengo otro software. ¿Cómo hago la transición?',
   'Un agente de implantación te explicará cómo exportar tus datos para migrarlos a Klowna con todas las garantías. Cuando llegue el momento, tendrás todo organizado y preparado.', 5),
  ('¿Puedo integrar Klowna con otras plataformas?',
   'Sí. Con el plan Ultimate tienes APIs para conectar e integrar otras plataformas y herramientas externas, ganando flexibilidad y funcionalidad al sincronizar distintas soluciones.', 6);
