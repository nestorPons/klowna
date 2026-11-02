-- dev-seed.sql — Datos de prueba para desarrollo.
-- Amplía el contenido demo del usuario demo@klowna.com y añade leads de ejemplo.
-- Idempotente: borra los datos demo existentes antes de volver a insertarlos.
-- Requiere las migraciones 005_agenda.sql, 006_services.sql y 020_deposits.sql.
-- Aplicar con: scripts/seed-dev.sh

SET @demo := (SELECT id FROM users WHERE email = 'demo@klowna.com');
SET @ana   := (SELECT id FROM users WHERE email = 'ana@klowna.com');
SET @carlos:= (SELECT id FROM users WHERE email = 'carlos@klowna.com');
SET @lucia := (SELECT id FROM users WHERE email = 'lucia@klowna.com');

-- --- Reinicio de datos demo -------------------------------------------------
DELETE FROM cash_movements WHERE user_id = @demo;
DELETE FROM cash_sessions WHERE user_id = @demo;
DELETE FROM appointments WHERE user_id = @demo;
DELETE FROM services WHERE user_id = @demo;
DELETE FROM invoices WHERE user_id = @demo;
DELETE FROM patient_history WHERE user_id = @demo;
DELETE FROM patients WHERE user_id = @demo;

-- --- Catálogo de servicios -------------------------------------------------
-- Cada servicio tiene precio, color e icono. Los que exigen señal la indican.
INSERT INTO services (user_id, name, duration_minutes, sort_order, price, deposit_required, deposit_amount, color, icon, created_at) VALUES
  (@demo, 'Consulta láser',          30,  0,  60.00, 1, 15.00, '#e05d5d', 'zap',        NOW()),
  (@demo, 'Revisión capilar',        30,  1,  45.00, 0,  0.00, '#8e6fd8', 'eye',        NOW()),
  (@demo, 'Primera visita',          30,  2,  50.00, 1, 10.00, '#f0a63a', 'star',       NOW()),
  (@demo, 'Documentación',           15,  3,   0.00, 0,  0.00, '#5b8def', 'clock',      NOW()),
  (@demo, 'Toxina botulínica',       30,  4, 180.00, 1, 30.00, '#43b8a0', 'droplet',    NOW()),
  (@demo, 'Sesión fisioterapia',     45,  5,  40.00, 1, 10.00, '#ef6a9b', 'heart',      NOW()),
  (@demo, 'Control nutricional',     30,  6,  55.00, 0,  0.00, '#6aa84f', 'leaf',       NOW()),
  (@demo, 'Revisión ortodoncia',     30,  7,  60.00, 0,  0.00, '#3aa7d8', 'smile',      NOW()),
  (@demo, 'Sesión psicología',       45,  8,  55.00, 0,  0.00, '#9b59b6', 'stethoscope',NOW()),
  (@demo, 'Fisioterapia deportiva',  45,  9,  45.00, 1, 10.00, '#e2703a', 'flame',      NOW()),
  (@demo, 'Depilación láser',        30, 10,  90.00, 1, 20.00, '#f2c14e', 'sun',        NOW()),
  (@demo, 'Diagnóstico capilar',     15, 11,  35.00, 0,  0.00, '#c07ac4', 'sparkles',   NOW()),
  (@demo, 'Fotografías',             15, 12,  20.00, 0,  0.00, '#d95f9b', 'num-0',      NOW()),
  (@demo, 'Plan nutricional',        30, 13,  70.00, 0,  0.00, '#4f9d5b', 'num-1',      NOW()),
  (@demo, 'Estudio podológico',      30, 14,  60.00, 0,  0.00, '#b5793f', 'scissors',   NOW()),
  (@demo, 'Ácido hialurónico',       45, 15, 260.00, 1, 50.00, '#2f9bc2', 'num-2',      NOW()),
  (@demo, 'Fisioterapia lumbalgia',  45, 16,  45.00, 1, 10.00, '#d04a6b', 'num-3',      NOW()),
  (@demo, 'Valoración inicial',      30, 17,  40.00, 0,  0.00, '#d98e2b', 'num-4',      NOW()),
  (@demo, 'Primera consulta nutrición', 30, 18, 60.00, 0, 0.00, '#7fb34f', 'num-5',     NOW());

-- --- Trabajadores (icono y color) -----------------------------------------
-- Colores propios del equipo, distintos a los del catálogo de servicios.
UPDATE users SET color = '#2563eb', icon = 'num-1' WHERE email = 'ana@klowna.com';
UPDATE users SET color = '#059669', icon = 'num-2' WHERE email = 'carlos@klowna.com';
UPDATE users SET color = '#db2777', icon = 'num-3' WHERE email = 'lucia@klowna.com';

-- --- Horario del centro y festivos ----------------------------------------
-- Domingo cerrado (no aparece en la agenda de semana) y sábado con horario corto.
INSERT INTO clinic_schedule (user_id, schedule_json, updated_at) VALUES
  (@demo, '[{"day":0,"open":"","close":""},{"day":1,"open":"08:00","close":"20:00"},{"day":2,"open":"08:00","close":"20:00"},{"day":3,"open":"08:00","close":"20:00"},{"day":4,"open":"08:00","close":"20:00"},{"day":5,"open":"08:00","close":"20:00"},{"day":6,"open":"09:00","close":"14:00"}]', NOW())
ON DUPLICATE KEY UPDATE schedule_json = VALUES(schedule_json), updated_at = VALUES(updated_at);

-- Descanso para comer de cada trabajador (lunes a viernes).
DELETE FROM schedule_blocks WHERE user_id = @demo;
INSERT INTO schedule_blocks (series_id, user_id, staff_id, weekday, start_time, end_time, reason) VALUES
  (1001, @demo, @ana,    1, '14:00:00', '15:00:00', 'Comida'),
  (1001, @demo, @ana,    2, '14:00:00', '15:00:00', 'Comida'),
  (1001, @demo, @ana,    3, '14:00:00', '15:00:00', 'Comida'),
  (1001, @demo, @ana,    4, '14:00:00', '15:00:00', 'Comida'),
  (1001, @demo, @ana,    5, '14:00:00', '15:00:00', 'Comida'),
  (1002, @demo, @carlos, 1, '14:30:00', '15:30:00', 'Comida'),
  (1002, @demo, @carlos, 2, '14:30:00', '15:30:00', 'Comida'),
  (1002, @demo, @carlos, 3, '14:30:00', '15:30:00', 'Comida'),
  (1002, @demo, @carlos, 4, '14:30:00', '15:30:00', 'Comida'),
  (1002, @demo, @carlos, 5, '14:30:00', '15:30:00', 'Comida'),
  (1003, @demo, @lucia,  1, '15:00:00', '16:00:00', 'Comida'),
  (1003, @demo, @lucia,  2, '15:00:00', '16:00:00', 'Comida'),
  (1003, @demo, @lucia,  3, '15:00:00', '16:00:00', 'Comida'),
  (1003, @demo, @lucia,  4, '15:00:00', '16:00:00', 'Comida'),
  (1003, @demo, @lucia,  5, '15:00:00', '16:00:00', 'Comida');

-- --- Festivos --------------------------------------------------------------
-- El domingo ya está cerrado en el horario; además se marca como festivo.
DELETE FROM holidays WHERE user_id = @demo;
INSERT INTO holidays (user_id, holiday_date, name) VALUES
  (@demo, DATE_ADD(CURDATE(), INTERVAL 10 DAY), 'Festivo de ejemplo')
ON DUPLICATE KEY UPDATE name = VALUES(name);
SET @next_sunday := DATE_ADD(CURDATE(), INTERVAL MOD(8 - DAYOFWEEK(CURDATE()), 7) DAY);
INSERT INTO holidays (user_id, holiday_date, name)
SELECT @demo, DATE_ADD(@next_sunday, INTERVAL n WEEK), 'Domingo (festivo)'
FROM (SELECT 0 AS n UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3
      UNION ALL SELECT 4 UNION ALL SELECT 5 UNION ALL SELECT 6 UNION ALL SELECT 7) AS weeks
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- --- Pacientes -------------------------------------------------------------
INSERT INTO patients (user_id, name, email, phone, notes, created_at) VALUES
  (@demo, 'Laura García',      'laura@ejemplo.com',           '+34 600 111 222', 'Medicina estética. Piel sensible. Firma digital pendiente de activar.', NOW()),
  (@demo, 'Marta Ruiz',        'marta@ejemplo.com',           '+34 600 333 444', 'Seguimiento capilar mensual. Fotografías de evolución.', NOW()),
  (@demo, 'Javier Sanz',       'javier@ejemplo.com',          '+34 600 555 666', 'Prefiere citas a primera hora.', NOW()),
  (@demo, 'Carmen Ortega',     'carmen.ortega@correo.com',    '+34 611 202 303', 'Toxina botulínica. Revisión a 30 días.', NOW()),
  (@demo, 'Pablo Navarro',     'pablo.navarro@correo.com',    '+34 622 404 505', 'Rehabilitación de rodilla. 8 sesiones de fisioterapia.', NOW()),
  (@demo, 'Lucía Molina',      'lucia.molina@correo.com',     '+34 633 606 707', 'Plan de nutrición. Control mensual de medidas.', NOW()),
  (@demo, 'Andrés Herrera',    'andres.herrera@correo.com',   '+34 644 808 909', 'Dermatología. Revisión de lesión pigmentada.', NOW()),
  (@demo, 'Elena Vidal',       'elena.vidal@correo.com',      '+34 655 101 202', 'Odontología. Presupuesto de ortodoncia.', NOW()),
  (@demo, 'Raquel Jiménez',    'raquel.jimenez@correo.com',   '+34 666 303 404', 'Psicología. Sesiones quincenales.', NOW()),
  (@demo, 'Miguel Ángel Torres','mangel.torres@correo.com',   '+34 677 505 606', 'Fisioterapia deportiva. Prefiere turno de tarde.', NOW()),
  (@demo, 'Sara Iglesias',     'sara.iglesias@correo.com',    '+34 688 707 808', 'Centro de estética. Depilación láser.', NOW()),
  (@demo, 'David Campos',      'david.campos@correo.com',     '+34 699 909 101', 'Medicina capilar. Protocolo de recuperación.', NOW()),
  (@demo, 'Nuria Serrano',     'nuria.serrano@correo.com',    '+34 610 212 313', 'Nutrición. Intolerancias alimentarias.', NOW()),
  (@demo, 'Álvaro Peña',       'alvaro.pena@correo.com',      '+34 621 414 515', 'Podología. Plantillas a medida.', NOW()),
  (@demo, 'Beatriz Lorenzo',   'beatriz.lorenzo@correo.com',  '+34 632 616 717', 'Medicina estética. Ácido hialurónico.', NOW()),
  (@demo, 'Héctor Román',      'hector.roman@correo.com',     '+34 643 818 919', 'Fisioterapia. Lumbalgia crónica.', NOW());

-- --- Historial de clientes -------------------------------------------------
INSERT INTO patient_history (user_id, patient_id, author_id, entry_type, content, created_at, updated_at) VALUES
  (@demo, (SELECT id FROM patients WHERE user_id=@demo AND name='Laura García'), @demo, 'nota',
   'Piel sensible. Evitar láser en zona peribucal. Señal de 15 € cobrada por Bizum.', NOW(), NOW()),
  (@demo, (SELECT id FROM patients WHERE user_id=@demo AND name='Marta Ruiz'), @demo, 'seguimiento',
   'Seguimiento capilar mensual. Buena tolerancia al tratamiento.', DATE_SUB(NOW(), INTERVAL 7 DAY), DATE_SUB(NOW(), INTERVAL 7 DAY)),
  (@demo, (SELECT id FROM patients WHERE user_id=@demo AND name='Pablo Navarro'), @demo, 'comentario',
   'Rehabilitación de rodilla: 5 de 8 sesiones completadas.', DATE_SUB(NOW(), INTERVAL 3 DAY), DATE_SUB(NOW(), INTERVAL 3 DAY)),
  (@demo, (SELECT id FROM patients WHERE user_id=@demo AND name='Elena Vidal'), @demo, 'incidencia',
   'Pendiente de traer la radiografía panorámica para el presupuesto de ortodoncia.', DATE_SUB(NOW(), INTERVAL 1 DAY), DATE_SUB(NOW(), INTERVAL 1 DAY));

-- --- Citas + servicios -----------------------------------------------------
-- Hoy
INSERT INTO appointments (user_id, patient_id, assignee_id, starts_at, ends_at, status, notes, created_at) VALUES
  (@demo, (SELECT id FROM patients WHERE user_id=@demo AND name='Laura García'),  @ana,    CONCAT(CURDATE(),' 09:30:00'), CONCAT(CURDATE(),' 10:00:00'), 'confirmada', 'Cabina 2', NOW());
SET @a := LAST_INSERT_ID();
INSERT INTO appointment_items (appointment_id, service_id, duration_minutes, sort_order) VALUES (@a, (SELECT id FROM services WHERE user_id=@demo AND name='Consulta láser'), 30, 0);

INSERT INTO appointments (user_id, patient_id, assignee_id, starts_at, ends_at, status, notes, created_at) VALUES
  (@demo, (SELECT id FROM patients WHERE user_id=@demo AND name='Marta Ruiz'),   @carlos, CONCAT(CURDATE(),' 10:30:00'), CONCAT(CURDATE(),' 11:00:00'), 'pendiente',  'Consulta 1', NOW());
SET @a := LAST_INSERT_ID();
INSERT INTO appointment_items (appointment_id, service_id, duration_minutes, sort_order) VALUES (@a, (SELECT id FROM services WHERE user_id=@demo AND name='Revisión capilar'), 30, 0);

INSERT INTO appointments (user_id, patient_id, assignee_id, starts_at, ends_at, status, notes, created_at) VALUES
  (@demo, (SELECT id FROM patients WHERE user_id=@demo AND name='Javier Sanz'),  @lucia,  CONCAT(CURDATE(),' 11:30:00'), CONCAT(CURDATE(),' 12:15:00'), 'confirmada', 'Documentación inicial', NOW());
SET @a := LAST_INSERT_ID();
INSERT INTO appointment_items (appointment_id, service_id, duration_minutes, sort_order) VALUES (@a, (SELECT id FROM services WHERE user_id=@demo AND name='Primera visita'), 30, 0), (@a, (SELECT id FROM services WHERE user_id=@demo AND name='Documentación'), 15, 1);

INSERT INTO appointments (user_id, patient_id, assignee_id, starts_at, ends_at, status, notes, created_at) VALUES
  (@demo, (SELECT id FROM patients WHERE user_id=@demo AND name='Carmen Ortega'), NULL,    CONCAT(CURDATE(),' 12:30:00'), CONCAT(CURDATE(),' 13:00:00'), 'pendiente',  'Cabina 1', NOW());
SET @a := LAST_INSERT_ID();
INSERT INTO appointment_items (appointment_id, service_id, duration_minutes, sort_order) VALUES (@a, (SELECT id FROM services WHERE user_id=@demo AND name='Toxina botulínica'), 30, 0);

INSERT INTO appointments (user_id, patient_id, assignee_id, starts_at, ends_at, status, notes, created_at) VALUES
  (@demo, (SELECT id FROM patients WHERE user_id=@demo AND name='Pablo Navarro'), @ana,    CONCAT(CURDATE(),' 16:00:00'), CONCAT(CURDATE(),' 16:45:00'), 'confirmada', 'Sala 3', NOW());
SET @a := LAST_INSERT_ID();
INSERT INTO appointment_items (appointment_id, service_id, duration_minutes, sort_order) VALUES (@a, (SELECT id FROM services WHERE user_id=@demo AND name='Sesión fisioterapia'), 45, 0);

INSERT INTO appointments (user_id, patient_id, assignee_id, starts_at, ends_at, status, notes, created_at) VALUES
  (@demo, (SELECT id FROM patients WHERE user_id=@demo AND name='Lucía Molina'),  @carlos, CONCAT(CURDATE(),' 17:30:00'), CONCAT(CURDATE(),' 18:00:00'), 'pendiente',  'Consulta 2', NOW());
SET @a := LAST_INSERT_ID();
INSERT INTO appointment_items (appointment_id, service_id, duration_minutes, sort_order) VALUES (@a, (SELECT id FROM services WHERE user_id=@demo AND name='Control nutricional'), 30, 0);

-- Mañana (+1)
INSERT INTO appointments (user_id, patient_id, assignee_id, starts_at, ends_at, status, notes, created_at) VALUES
  (@demo, (SELECT id FROM patients WHERE user_id=@demo AND name='Elena Vidal'),   @lucia,  CONCAT(DATE_ADD(CURDATE(), INTERVAL 1 DAY),' 10:00:00'), CONCAT(DATE_ADD(CURDATE(), INTERVAL 1 DAY),' 10:30:00'), 'pendiente',  'Traer radiografía panorámica', NOW());
SET @a := LAST_INSERT_ID();
INSERT INTO appointment_items (appointment_id, service_id, duration_minutes, sort_order) VALUES (@a, (SELECT id FROM services WHERE user_id=@demo AND name='Revisión ortodoncia'), 30, 0);

INSERT INTO appointments (user_id, patient_id, assignee_id, starts_at, ends_at, status, notes, created_at) VALUES
  (@demo, (SELECT id FROM patients WHERE user_id=@demo AND name='Raquel Jiménez'), @ana,    CONCAT(DATE_ADD(CURDATE(), INTERVAL 1 DAY),' 11:00:00'), CONCAT(DATE_ADD(CURDATE(), INTERVAL 1 DAY),' 11:45:00'), 'confirmada', 'Sesión quincenal · seguimiento', NOW());
SET @a := LAST_INSERT_ID();
INSERT INTO appointment_items (appointment_id, service_id, duration_minutes, sort_order) VALUES (@a, (SELECT id FROM services WHERE user_id=@demo AND name='Sesión psicología'), 45, 0);

INSERT INTO appointments (user_id, patient_id, assignee_id, starts_at, ends_at, status, notes, created_at) VALUES
  (@demo, (SELECT id FROM patients WHERE user_id=@demo AND name='Miguel Ángel Torres'), NULL, CONCAT(DATE_ADD(CURDATE(), INTERVAL 1 DAY),' 17:00:00'), CONCAT(DATE_ADD(CURDATE(), INTERVAL 1 DAY),' 17:45:00'), 'pendiente',  'Traer rodillera y estudio previo', NOW());
SET @a := LAST_INSERT_ID();
INSERT INTO appointment_items (appointment_id, service_id, duration_minutes, sort_order) VALUES (@a, (SELECT id FROM services WHERE user_id=@demo AND name='Fisioterapia deportiva'), 45, 0);

INSERT INTO appointments (user_id, patient_id, assignee_id, starts_at, ends_at, status, notes, created_at) VALUES
  (@demo, (SELECT id FROM patients WHERE user_id=@demo AND name='Sara Iglesias'), @carlos, CONCAT(DATE_ADD(CURDATE(), INTERVAL 1 DAY),' 18:00:00'), CONCAT(DATE_ADD(CURDATE(), INTERVAL 1 DAY),' 18:30:00'), 'confirmada', 'Zona piernas · evitar sol 48 h', NOW());
SET @a := LAST_INSERT_ID();
INSERT INTO appointment_items (appointment_id, service_id, duration_minutes, sort_order) VALUES (@a, (SELECT id FROM services WHERE user_id=@demo AND name='Depilación láser'), 30, 0);

-- +2 días
INSERT INTO appointments (user_id, patient_id, assignee_id, starts_at, ends_at, status, notes, created_at) VALUES
  (@demo, (SELECT id FROM patients WHERE user_id=@demo AND name='David Campos'),  @lucia,  CONCAT(DATE_ADD(CURDATE(), INTERVAL 2 DAY),' 09:30:00'), CONCAT(DATE_ADD(CURDATE(), INTERVAL 2 DAY),' 10:00:00'), 'pendiente',  'Revisión de fotos de evolución', NOW());
SET @a := LAST_INSERT_ID();
INSERT INTO appointment_items (appointment_id, service_id, duration_minutes, sort_order) VALUES (@a, (SELECT id FROM services WHERE user_id=@demo AND name='Diagnóstico capilar'), 15, 0), (@a, (SELECT id FROM services WHERE user_id=@demo AND name='Fotografías'), 15, 1);

INSERT INTO appointments (user_id, patient_id, assignee_id, starts_at, ends_at, status, notes, created_at) VALUES
  (@demo, (SELECT id FROM patients WHERE user_id=@demo AND name='Nuria Serrano'), @ana,    CONCAT(DATE_ADD(CURDATE(), INTERVAL 2 DAY),' 12:00:00'), CONCAT(DATE_ADD(CURDATE(), INTERVAL 2 DAY),' 12:30:00'), 'confirmada', 'Traer analítica reciente', NOW());
SET @a := LAST_INSERT_ID();
INSERT INTO appointment_items (appointment_id, service_id, duration_minutes, sort_order) VALUES (@a, (SELECT id FROM services WHERE user_id=@demo AND name='Plan nutricional'), 30, 0);

INSERT INTO appointments (user_id, patient_id, assignee_id, starts_at, ends_at, status, notes, created_at) VALUES
  (@demo, (SELECT id FROM patients WHERE user_id=@demo AND name='Álvaro Peña'),   NULL,    CONCAT(DATE_ADD(CURDATE(), INTERVAL 2 DAY),' 16:30:00'), CONCAT(DATE_ADD(CURDATE(), INTERVAL 2 DAY),' 17:00:00'), 'pendiente',  'Entrega de plantillas personalizadas', NOW());
SET @a := LAST_INSERT_ID();
INSERT INTO appointment_items (appointment_id, service_id, duration_minutes, sort_order) VALUES (@a, (SELECT id FROM services WHERE user_id=@demo AND name='Estudio podológico'), 30, 0);

-- +3 días
INSERT INTO appointments (user_id, patient_id, assignee_id, starts_at, ends_at, status, notes, created_at) VALUES
  (@demo, (SELECT id FROM patients WHERE user_id=@demo AND name='Beatriz Lorenzo'), @carlos, CONCAT(DATE_ADD(CURDATE(), INTERVAL 3 DAY),' 10:00:00'), CONCAT(DATE_ADD(CURDATE(), INTERVAL 3 DAY),' 10:45:00'), 'confirmada', 'Zona surcos nasogenianos', NOW());
SET @a := LAST_INSERT_ID();
INSERT INTO appointment_items (appointment_id, service_id, duration_minutes, sort_order) VALUES (@a, (SELECT id FROM services WHERE user_id=@demo AND name='Ácido hialurónico'), 45, 0);

INSERT INTO appointments (user_id, patient_id, assignee_id, starts_at, ends_at, status, notes, created_at) VALUES
  (@demo, (SELECT id FROM patients WHERE user_id=@demo AND name='Héctor Román'),  @ana,    CONCAT(DATE_ADD(CURDATE(), INTERVAL 3 DAY),' 11:00:00'), CONCAT(DATE_ADD(CURDATE(), INTERVAL 3 DAY),' 11:45:00'), 'pendiente',  'Lumbalgia · valorar calor local', NOW());
SET @a := LAST_INSERT_ID();
INSERT INTO appointment_items (appointment_id, service_id, duration_minutes, sort_order) VALUES (@a, (SELECT id FROM services WHERE user_id=@demo AND name='Fisioterapia lumbalgia'), 45, 0);

-- Ayer (-1)
INSERT INTO appointments (user_id, patient_id, assignee_id, starts_at, ends_at, status, notes, created_at) VALUES
  (@demo, (SELECT id FROM patients WHERE user_id=@demo AND name='Carmen Ortega'), @ana,    CONCAT(DATE_SUB(CURDATE(), INTERVAL 1 DAY),' 12:00:00'), CONCAT(DATE_SUB(CURDATE(), INTERVAL 1 DAY),' 12:30:00'), 'confirmada', 'Revisión 30 días post-toxina', NOW());
SET @a := LAST_INSERT_ID();
INSERT INTO appointment_items (appointment_id, service_id, duration_minutes, sort_order) VALUES (@a, (SELECT id FROM services WHERE user_id=@demo AND name='Valoración inicial'), 30, 0);

INSERT INTO appointments (user_id, patient_id, assignee_id, starts_at, ends_at, status, notes, created_at) VALUES
  (@demo, (SELECT id FROM patients WHERE user_id=@demo AND name='Pablo Navarro'), NULL,    CONCAT(DATE_SUB(CURDATE(), INTERVAL 1 DAY),' 17:00:00'), CONCAT(DATE_SUB(CURDATE(), INTERVAL 1 DAY),' 17:45:00'), 'confirmada', 'Sesión 5/8 · evolución favorable', NOW());
SET @a := LAST_INSERT_ID();
INSERT INTO appointment_items (appointment_id, service_id, duration_minutes, sort_order) VALUES (@a, (SELECT id FROM services WHERE user_id=@demo AND name='Sesión fisioterapia'), 45, 0);

-- -2 días
INSERT INTO appointments (user_id, patient_id, assignee_id, starts_at, ends_at, status, notes, created_at) VALUES
  (@demo, (SELECT id FROM patients WHERE user_id=@demo AND name='Lucía Molina'),  @carlos, CONCAT(DATE_SUB(CURDATE(), INTERVAL 2 DAY),' 10:30:00'), CONCAT(DATE_SUB(CURDATE(), INTERVAL 2 DAY),' 11:00:00'), 'confirmada', 'Ajuste de plan · intolerancias', NOW());
SET @a := LAST_INSERT_ID();
INSERT INTO appointment_items (appointment_id, service_id, duration_minutes, sort_order) VALUES (@a, (SELECT id FROM services WHERE user_id=@demo AND name='Primera consulta nutrición'), 30, 0);

-- --- Señales de reserva ----------------------------------------------------
-- La señal exigida se calcula sumando las señales de los servicios de cada cita.
UPDATE appointments a
JOIN (
  SELECT ai.appointment_id,
         SUM(CASE WHEN s.deposit_required = 1 THEN s.deposit_amount ELSE 0 END) AS due
  FROM appointment_items ai
  JOIN services s ON s.id = ai.service_id
  GROUP BY ai.appointment_id
) d ON d.appointment_id = a.id
SET a.deposit_due = d.due
WHERE a.user_id = @demo;

-- Citas con la señal ya cobrada (aparecen con el cobro en caja).
SET @appt_laura := (SELECT a.id FROM appointments a JOIN patients p ON p.id = a.patient_id
  WHERE a.user_id = @demo AND p.name = 'Laura García' AND DATE(a.starts_at) = CURDATE() LIMIT 1);
SET @appt_pablo := (SELECT a.id FROM appointments a JOIN patients p ON p.id = a.patient_id
  WHERE a.user_id = @demo AND p.name = 'Pablo Navarro' AND DATE(a.starts_at) = CURDATE() LIMIT 1);

UPDATE appointments SET deposit_paid = deposit_due, deposit_payment_method = 'bizum', deposit_paid_at = NOW()
WHERE id = @appt_laura;
UPDATE appointments SET deposit_paid = deposit_due, deposit_payment_method = 'efectivo', deposit_paid_at = NOW()
WHERE id = @appt_pablo;

-- --- Facturas --------------------------------------------------------------
INSERT INTO invoices (user_id, patient_id, number, concept, amount, status, issued_at, paid_at, created_at) VALUES
  (@demo, (SELECT id FROM patients WHERE user_id = @demo AND name = 'Laura García'),       'F-2026-0001', 'Tratamiento láser (3 sesiones)',      240.00, 'pagada',    CONCAT(DATE_SUB(CURDATE(), INTERVAL 20 DAY), ' 09:00:00'), CONCAT(DATE_SUB(CURDATE(), INTERVAL 20 DAY), ' 09:05:00'), NOW()),
  (@demo, (SELECT id FROM patients WHERE user_id = @demo AND name = 'Marta Ruiz'),         'F-2026-0002', 'Consulta y diagnóstico capilar',        90.00, 'pendiente', CONCAT(DATE_SUB(CURDATE(), INTERVAL 3 DAY),  ' 09:00:00'), NULL, NOW()),
  (@demo, (SELECT id FROM patients WHERE user_id = @demo AND name = 'Javier Sanz'),        'F-2026-0003', 'Bono 5 sesiones de mantenimiento',     320.00, 'pendiente', CONCAT(DATE_SUB(CURDATE(), INTERVAL 2 DAY),  ' 09:00:00'), NULL, NOW()),
  (@demo, (SELECT id FROM patients WHERE user_id = @demo AND name = 'Carmen Ortega'),      'F-2026-0004', 'Toxina botulínica (1 zona)',          180.00, 'pagada',    CONCAT(DATE_SUB(CURDATE(), INTERVAL 15 DAY), ' 09:00:00'), CONCAT(DATE_SUB(CURDATE(), INTERVAL 15 DAY), ' 09:10:00'), NOW()),
  (@demo, (SELECT id FROM patients WHERE user_id = @demo AND name = 'Pablo Navarro'),      'F-2026-0005', 'Pack 8 sesiones de fisioterapia',      400.00, 'pagada',    CONCAT(DATE_SUB(CURDATE(), INTERVAL 30 DAY), ' 09:00:00'), CONCAT(DATE_SUB(CURDATE(), INTERVAL 29 DAY), ' 09:00:00'), NOW()),
  (@demo, (SELECT id FROM patients WHERE user_id = @demo AND name = 'Lucía Molina'),       'F-2026-0006', 'Plan nutricional 3 meses',            210.00, 'pagada',    CONCAT(DATE_SUB(CURDATE(), INTERVAL 10 DAY), ' 09:00:00'), CONCAT(DATE_SUB(CURDATE(), INTERVAL 10 DAY), ' 09:02:00'), NOW()),
  (@demo, (SELECT id FROM patients WHERE user_id = @demo AND name = 'Andrés Herrera'),     'F-2026-0007', 'Dermatoscopia digital',                95.00, 'pendiente', CONCAT(DATE_SUB(CURDATE(), INTERVAL 1 DAY),  ' 09:00:00'), NULL, NOW()),
  (@demo, (SELECT id FROM patients WHERE user_id = @demo AND name = 'Elena Vidal'),        'F-2026-0008', 'Presupuesto de ortodoncia',            60.00, 'pendiente', CONCAT(CURDATE(), ' 09:00:00'), NULL, NOW()),
  (@demo, (SELECT id FROM patients WHERE user_id = @demo AND name = 'Raquel Jiménez'),     'F-2026-0009', 'Sesión de psicología',                 55.00, 'pagada',    CONCAT(DATE_SUB(CURDATE(), INTERVAL 5 DAY),  ' 09:00:00'), CONCAT(DATE_SUB(CURDATE(), INTERVAL 5 DAY),  ' 09:00:00'), NOW()),
  (@demo, (SELECT id FROM patients WHERE user_id = @demo AND name = 'Sara Iglesias'),      'F-2026-0010', 'Bono 6 sesiones de láser',            300.00, 'pendiente', CONCAT(DATE_SUB(CURDATE(), INTERVAL 1 DAY),  ' 09:00:00'), NULL, NOW()),
  (@demo, (SELECT id FROM patients WHERE user_id = @demo AND name = 'Miguel Ángel Torres'),'F-2026-0011', 'Fisioterapia deportiva',              120.00, 'pagada',    CONCAT(DATE_SUB(CURDATE(), INTERVAL 7 DAY),  ' 09:00:00'), CONCAT(DATE_SUB(CURDATE(), INTERVAL 7 DAY),  ' 09:03:00'), NOW()),
  (@demo, (SELECT id FROM patients WHERE user_id = @demo AND name = 'Beatriz Lorenzo'),    'F-2026-0012', 'Ácido hialurónico',                  260.00, 'pendiente', CONCAT(CURDATE(), ' 09:00:00'), NULL, NOW());

-- Forma de pago de las facturas ya cobradas que no la tuvieran.
UPDATE invoices SET payment_method = 'tarjeta'
WHERE user_id = @demo AND status = 'pagada' AND payment_method = '';

-- --- Caja de hoy -----------------------------------------------------------
-- Caja abierta con fondo inicial, dos señales cobradas, un cobro de factura y un gasto.
INSERT INTO cash_sessions (user_id, session_date, opening_amount, counted_amount, expected_cash, difference, denoms_json, notes, status, opened_at, updated_at)
VALUES (@demo, CURDATE(), 100.00, 0, 0, 0, '[]', '', 'abierta', NOW(), NOW());
SET @cash := LAST_INSERT_ID();

UPDATE invoices SET status = 'pagada', paid_at = NOW(), payment_method = 'tarjeta'
WHERE user_id = @demo AND number = 'F-2026-0008';

INSERT INTO cash_movements (user_id, session_id, movement_type, payment_method, amount, concept, invoice_id, appointment_id, created_at) VALUES
  (@demo, @cash, 'ingreso', 'bizum',     15.00, 'Señal cita · Laura García',  NULL, @appt_laura, NOW()),
  (@demo, @cash, 'ingreso', 'efectivo',  10.00, 'Señal cita · Pablo Navarro', NULL, @appt_pablo, NOW()),
  (@demo, @cash, 'ingreso', 'tarjeta',   60.00, 'Cobro factura F-2026-0008',
    (SELECT id FROM invoices WHERE user_id = @demo AND number = 'F-2026-0008'), NULL, NOW()),
  (@demo, @cash, 'gasto',   'efectivo',  12.50, 'Compra de consumibles', NULL, NULL, NOW());

-- --- Leads -----------------------------------------------------------------
DELETE FROM leads WHERE email IN (
  'maria.lopez@clinica.com', 'info@centromedico.es', 'clara@spa.es', 'pedro@dentalpro.es',
  'lucia@estetica.es', 'jorge@clinicasol.com', 'ana@podologia.es', 'marta@psicologia.es',
  'info@fisioterapia.es', 'david@capilar.es', 'nuria@nutricion.es', 'eva@derma.es'
);

INSERT INTO leads (name, email, phone, clinic, sector, message, consent, source, ip, user_agent, created_at) VALUES
  ('María López',      'maria.lopez@clinica.com',     '+34 600 111 000', 'Clínica Bella',      'Medicina estética', 'Quiero digitalizar la historia clínica y las fotos de mis pacientes.', 1, 'web',      '83.45.1.10', 'Mozilla/5.0', DATE_SUB(NOW(), INTERVAL 3 DAY)),
  ('Juan Gómez',       'info@centromedico.es',        '+34 600 222 000', 'Centro Médico Sol',  'Salud',             'Necesito firma digital de consentimientos y cumplir RGPD.', 1, 'contacto', '83.45.1.11', 'Mozilla/5.0', DATE_SUB(NOW(), INTERVAL 2 DAY)),
  ('Clara Fernández',  'clara@spa.es',                '+34 600 333 000', 'Spa Zen',            'Belleza y bienestar','Buscamos reservas online y recordatorios por WhatsApp.', 1, 'web',      '83.45.1.12', 'Mozilla/5.0', DATE_SUB(NOW(), INTERVAL 1 DAY)),
  ('Pedro Sánchez',    'pedro@dentalpro.es',          '+34 600 444 000', 'Dental Pro',         'Salud',             'Presupuestos y planes dentales para mi clínica de 3 gabinetes.', 1, 'web', '83.45.1.13', 'Mozilla/5.0', DATE_SUB(NOW(), INTERVAL 4 DAY)),
  ('Lucía Torres',     'lucia@estetica.es',           '+34 600 555 000', 'Estética Lucía',     'Belleza y bienestar','Quiero probar la demo de 10 días.', 1, 'contacto', '83.45.1.14', 'Mozilla/5.0', DATE_SUB(NOW(), INTERVAL 5 DAY)),
  ('Jorge Martín',     'jorge@clinicasol.com',        '+34 600 666 000', 'Clínica Sol',        'Medicina estética', 'Gestionamos dos sedes, ¿qué licencia necesito?', 0, 'web', '83.45.1.15', 'Mozilla/5.0', DATE_SUB(NOW(), INTERVAL 6 DAY)),
  ('Ana Prieto',       'ana@podologia.es',            '+34 600 777 000', 'Podología Ana',      'Salud',             'Control de plantillas y seguimiento de pacientes.', 1, 'web', '83.45.1.16', 'Mozilla/5.0', DATE_SUB(NOW(), INTERVAL 7 DAY)),
  ('Marta Ibáñez',     'marta@psicologia.es',         '+34 600 888 000', 'Consulta Ibáñez',    'Salud',             'Necesito notas privadas y gestión de citas.', 1, 'web', '83.45.1.17', 'Mozilla/5.0', DATE_SUB(NOW(), INTERVAL 8 DAY)),
  ('Info',             'info@fisioterapia.es',        '+34 600 999 000', 'Fisio Plus',         'Salud',             '¿Tenéis planes para fisioterapeutas con bonos de sesiones?', 1, 'web', '83.45.1.18', 'Mozilla/5.0', DATE_SUB(NOW(), INTERVAL 9 DAY)),
  ('David Roca',       'david@capilar.es',            '+34 601 111 000', 'Capilar Center',     'Medicina estética', 'Quiero campañas de fidelización post-tratamiento.', 1, 'contacto', '83.45.1.19', 'Mozilla/5.0', DATE_SUB(NOW(), INTERVAL 10 DAY)),
  ('Nuria Pastor',     'nuria@nutricion.es',          '+34 601 222 000', 'Nutri Vida',         'Salud',             'Seguimiento de medidas y planes personalizados.', 1, 'web', '83.45.1.20', 'Mozilla/5.0', DATE_SUB(NOW(), INTERVAL 11 DAY)),
  ('Eva Gil',          'eva@derma.es',                '+34 601 333 000', 'Derma Estética',     'Medicina estética', 'Control de stock y trazabilidad de consumibles.', 1, 'web', '83.45.1.21', 'Mozilla/5.0', DATE_SUB(NOW(), INTERVAL 12 DAY));
