INSERT INTO site_settings (setting_key, setting_value) VALUES
  ('site_name', 'Klowna'),
  ('tagline', 'Agenda electrónica y software de gestión de citas a medida'),
  ('phone', '+34 660 291 797'),
  ('email', 'nestorpons@gmail.com'),
  ('whatsapp', '34660291797'),
  ('address', 'España · Atención online'),
  ('legal_note', 'Software de agenda electrónica y gestión de citas a medida.');

INSERT INTO sectors (id, slug, name, icon, tagline, description, sort_order) VALUES
  (1, 'medicina-estetica', 'Medicina estética', 'sparkles',
   'Gestiona tu clínica con historia clínica, fotografías y firma digital.',
   'Centraliza la historia clínica, las fotografías antes y después, los consentimientos y el seguimiento de tus pacientes de medicina estética en una sola plataforma. Automatiza la captación y fideliza con campañas personalizadas.',
   1),
  (2, 'belleza-bienestar', 'Belleza y bienestar', 'leaf',
   'Agenda, equipo y clientes bajo control, sin esfuerzo.',
   'Deja la gestión y la operativa diaria en manos de un software pensado para centros de belleza. Conecta con tu público en sus canales favoritos, atrae y fideliza clientes y controla tus datos en tiempo real.',
   2),
  (3, 'salud', 'Salud', 'heart',
   'Cumple la RGPD y dedica más tiempo a tus pacientes.',
   'Asegura una gestión eficaz y confiable de tu clínica: firma digital de consentimientos, documentación médica segura y trazabilidad de consumibles para preservar la seguridad del paciente.',
   3);

INSERT INTO specialties (sector_id, slug, name, description, sort_order) VALUES
  (1, 'clinica-medicina-estetica', 'Clínica medicina estética', 'Historia clínica, fotos y protocolos de tratamiento.', 1),
  (1, 'clinica-medicina-capilar', 'Clínica medicina capilar', 'Seguimiento capilar con evolución fotográfica.', 2),
  (2, 'centro-estetica', 'Centro de estética', 'Agenda inteligente, bonos y fidelización.', 1),
  (2, 'cejas-pestanas', 'Cejas y pestañas', 'Citas rápidas y recordatorios automáticos.', 2),
  (2, 'centro-masajes', 'Centro de masajes', 'Cabinas, terapeutas y disponibilidad real.', 3),
  (2, 'depilacion-laser', 'Depilación láser', 'Bonos, sesiones y control de equipos.', 4),
  (2, 'peluqueria-barberia', 'Peluquería y barbería', 'Turnos, servicios y venta de productos.', 5),
  (2, 'medicina-alternativa', 'Medicina alternativa', 'Historiales y planes de sesiones.', 6),
  (2, 'salon-unas', 'Salón de uñas', 'Reservas online y gestión de cabinas.', 7),
  (2, 'spa', 'Spa', 'Experiencias, packs y ocupación por sala.', 8),
  (3, 'nutricion', 'Nutrición', 'Seguimiento de medidas, planes y evolución.', 1),
  (3, 'odontologia', 'Odontología', 'Presupuestos, producción y planes dentales.', 2),
  (3, 'obstetricia', 'Obstetricia', 'Seguimiento del embarazo y pruebas.', 3),
  (3, 'fisioterapia', 'Fisioterapia', 'Sesiones, rehabilitación y bonos.', 4),
  (3, 'clinicas-medicas', 'Clínicas médicas', 'Historiales, pruebas y derivaciones.', 5),
  (3, 'psicologia', 'Psicología', 'Sesiones, notas privadas y seguimiento.', 6),
  (3, 'dermatologia', 'Dermatología', 'Lesiones, fotos clínicas y tratamientos.', 7),
  (3, 'ginecologia', 'Ginecología', 'Historial ginecológico y pruebas.', 8),
  (3, 'laboratorios-clinicos', 'Laboratorios clínicos', 'Muestras, resultados y trazabilidad.', 9),
  (3, 'podologia', 'Podología', 'Exploraciones, plantillas y seguimiento.', 10);

INSERT INTO solutions (slug, name, description, icon, sort_order) VALUES
  ('saas', 'Klowna SaaS', 'Agenda, historia clínica, facturación y marketing en una sola plataforma.', 'cloud', 1),
  ('app', 'Klowna app', 'Tu clínica en el bolsillo de tu equipo y de tus pacientes.', 'smartphone', 2),
  ('whatsapp', 'Klowna whatsapp', 'Responde, reserva y recuerda citas por WhatsApp.', 'whatsapp', 3),
  ('doctor', 'Klowna doctor', 'Herramientas digitales diseñadas para el médico.', 'stethoscope', 4),
  ('web', 'Klowna web', 'Una experiencia omnicanal que facilita la reserva y la compra.', 'globe', 5),
  ('crm', 'Klowna crm', 'Capta y construye relaciones duraderas con tu público.', 'users', 6),
  ('stock', 'Klowna stock', 'Gestiona el stock de forma centralizada e inteligente.', 'box', 7),
  ('book', 'Klowna book', 'Reduce las ausencias con la posibilidad de reagendar.', 'calendar', 8),
  ('social', 'Klowna social', 'Centraliza tus redes sociales, ahorra tiempo y consigue resultados.', 'share', 9);

INSERT INTO challenges (title, description, icon, sort_order) VALUES
  ('Humanización de la experiencia', 'Delega las tareas repetitivas en la IA y la automatización para que tu equipo se centre en escuchar, diagnosticar y priorizar al paciente.', 'heart', 1),
  ('Captación omnicanal y cita online', 'Abre tu clínica 24/7 en WhatsApp o redes sociales y permite reservar según tu disponibilidad real.', 'calendar', 2),
  ('Datos clínicos y digitalización total', 'Centraliza historiales, evoluciones y consentimientos con firma digital en la nube, de forma segura.', 'file', 3),
  ('Marketing automatizado', 'Activa campañas post-tratamiento, fidelización y seguimiento de presupuestos pendientes.', 'megaphone', 4),
  ('Presupuestos y pagos sin fricciones', 'Diseña planes de tratamiento personalizados, presupuestos detallados y pagos integrados.', 'card', 5),
  ('Control de datos estratégicos', 'Toma mejores decisiones con estadísticas e informes en tiempo real de tu clínica.', 'chart', 6);

INSERT INTO stats (value, label, sort_order) VALUES
  ('+64 millones', 'Citas validadas', 1),
  ('+10 millones', 'Citas por app', 2),
  ('+95.000', 'Profesionales trabajando', 3);

INSERT INTO testimonials (sector, company, quote, url, sort_order) VALUES
  ('Belleza', 'Clínicas Tulaser', 'Con Klowna somos un 20% más eficientes gracias a la gestión centralizada de nuestras clínicas.', '', 1),
  ('Medicina estética', 'Clínicas Dra. Yasmin', 'Klowna nos ayuda a alcanzar nuestros objetivos de eficiencia y rentabilidad.', '', 2),
  ('Podología', 'Raúl Llorca', 'Me ayuda a gestionar mis clínicas y a cuidar la relación con mis pacientes.', '', 3);

INSERT INTO posts (slug, title, excerpt, category, body, image_url, published_at) VALUES
  ('webhooks-el-motor-que-avisa-al-instante-de-lo-que-pasa-en-tu-clinica',
   'Webhooks: el motor que avisa al instante de lo que pasa en tu clínica',
   'Conecta tus herramientas y recibe avisos automáticos cada vez que ocurre algo relevante.',
   'Gestión',
   '<p>Los webhooks son mensajes automáticos que una aplicación envía a otra cuando ocurre un evento. En una clínica permiten sincronizar la agenda, el CRM y la facturación sin intervención manual.</p><h2>Qué puedes automatizar</h2><ul><li>Avisar a tu CRM cuando se crea una cita.</li><li>Enviar una encuesta de satisfacción tras el tratamiento.</li><li>Actualizar el stock al cerrar una sesión.</li></ul><p>Con una configuración sencilla, tu equipo deja de copiar datos entre herramientas y se centra en el paciente.</p>',
   '', '2026-10-18 09:00:00'),
  ('facilita-la-colaboracion-con-profesionales-itinerantes-en-tu-clinica',
   'Facilita la colaboración con profesionales itinerantes en tu clínica',
   'Permisos, agendas y comisiones claras para equipos que trabajan en varias sedes.',
   'Gestión',
   '<p>Cada vez más clínicas trabajan con profesionales que reparten su tiempo entre varias sedes. Gestionar su agenda y sus permisos sin fricción es clave para mantener la calidad del servicio.</p><h2>Claves para organizarlo</h2><ul><li>Agendas independientes por profesional y sede.</li><li>Permisos de acceso ajustados a su rol.</li><li>Informes de actividad y comisiones automatizados.</li></ul><p>Un software centralizado evita solapamientos y da tranquilidad a todo el equipo.</p>',
   '', '2026-10-10 09:00:00'),
  ('estrategias-para-proteger-tu-agenda-ante-las-cancelaciones-de-cita',
   'Estrategias para proteger tu agenda ante las cancelaciones de cita',
   'Recordatorios, listas de espera y reagendado automático para reducir huecos.',
   'Gestión',
   '<p>Una cancelación de última hora puede arruinar la rentabilidad de una jornada. La buena noticia es que gran parte de estas ausencias se pueden prevenir.</p><h2>Tres medidas que funcionan</h2><ul><li>Recordatorios por WhatsApp y email 24 horas antes.</li><li>Confirmación con un clic desde el móvil.</li><li>Lista de espera que cubre automáticamente los huecos.</li></ul><p>Con estas rutinas, la ocupación de tu agenda mejora de forma sostenida.</p>',
   '', '2026-09-28 09:00:00'),
  ('ajusta-el-paciente-ideal-por-tratamiento-y-multiplica-tus-ventas',
   'Ajusta el paciente ideal por tratamiento y multiplica tus ventas',
   'Segmenta tus campañas con datos reales y habla con la persona adecuada.',
   'Marketing',
   '<p>No todos los tratamientos atraen al mismo perfil de paciente. Definir el paciente ideal para cada servicio permite invertir mejor en marketing y aumentar la conversión.</p><h2>Cómo empezar</h2><ul><li>Analiza qué servicios generan más recurrencia.</li><li>Identifica edad, intereses y canal de cada segmento.</li><li>Diseña mensajes específicos para cada perfil.</li></ul><p>El CRM de Klowna te ayuda a medir y ajustar cada campaña con datos reales.</p>',
   '', '2026-09-15 09:00:00'),
  ('como-activar-tu-protocolo-de-recurrencia-en-medicina-estetica',
   'Cómo activar tu protocolo de recurrencia en medicina estética',
   'Automatiza el seguimiento para que tus pacientes vuelvan en el momento justo.',
   'Medicina estética, Marketing',
   '<p>En medicina estética, la recurrencia es la base de un negocio sostenible. Un protocolo bien diseñado avisa al paciente cuando toca su siguiente sesión.</p><h2>Pasos del protocolo</h2><ul><li>Define la cadencia de cada tratamiento.</li><li>Programa recordatorios personalizados.</li><li>Mide la tasa de retorno y optimiza.</li></ul><p>La automatización hace que tu equipo no tenga que recordar cada caso manualmente.</p>',
   '', '2026-09-02 09:00:00'),
  ('klowna-pay-la-pasarela-de-pago-inteligente-que-simplifica-tus-cobros',
   'Klowna pay: la pasarela de pago inteligente que simplifica tus cobros',
   'Cobra desde el presupuesto, sin fricciones y con conciliación automática.',
   'Gestión',
   '<p>El momento del pago es parte de la experiencia del paciente. Una pasarela integrada permite cobrar desde el mismo presupuesto y conciliar los ingresos sin trabajo extra.</p><h2>Ventajas</h2><ul><li>Pagos con tarjeta, Bizum o transferencia.</li><li>Planes de pago fraccionados.</li><li>Conciliación automática con la facturación.</li></ul><p>Menos pasos manuales significa menos errores y más tiempo para atender.</p>',
   '', '2026-08-20 09:00:00');

INSERT INTO plans (name, price, period, description, features_json, highlighted, sort_order) VALUES
  ('Essential', '39 €', '/mes', 'Para profesionales que empiezan a digitalizar su consulta.',
   '["Agenda online 24/7","Ficha de paciente","Recordatorios por SMS y email","Informes básicos","Soporte por email"]', 0, 1),
  ('Pro', '79 €', '/mes', 'El plan completo para clínicas que quieren crecer.',
   '["Todo lo de Essential","WhatsApp integrado","Marketing automatizado","Firma digital de consentimientos","Control de stock","Soporte prioritario"]', 1, 2),
  ('Enterprise', 'A medida', '', 'Para grupos y cadenas con necesidades avanzadas.',
   '["Todo lo de Pro","Multi-sede y permisos avanzados","IA médica y automatizaciones","Integraciones a medida","Gestor de cuenta dedicado"]', 0, 3);

INSERT INTO faqs (question, answer, sort_order) VALUES
  ('¿Cuánto tarda la implantación?', 'La mayoría de centros están operativos en menos de 48 horas. Importamos tus datos y configuramos la agenda contigo.', 1),
  ('¿Los datos están protegidos?', 'Sí. Trabajamos con cifrado en tránsito y en reposo, copias de seguridad diarias y un modelo de tratamiento alineado con el RGPD.', 2),
  ('¿Puedo gestionar varias sedes?', 'Sí. Los planes superiores permiten gestionar varias sedes, equipos y permisos desde un único panel.', 3),
  ('¿Funciona en móvil y tablet?', 'Sí. La plataforma es responsive y dispone de app para iOS y Android.', 4),
  ('¿Hay permanencia?', 'No. Los planes mensuales no tienen permanencia; puedes cancelar cuando quieras.', 5),
  ('¿Ofrecéis formación?', 'Sí. Incluimos onboarding guiado, sesiones de formación y una base de conocimiento siempre disponible.', 6);
