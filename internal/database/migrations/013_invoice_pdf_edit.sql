-- 013_invoice_pdf_edit.sql — Precio congelado en la cita, factura única por cita y edición.

-- 1. El precio se congela en cada línea de la cita (para que la factura use el
--    precio acordado en el momento de la cita, no el precio actual del servicio).
ALTER TABLE appointment_items
  ADD COLUMN price DECIMAL(10,2) NOT NULL DEFAULT 0.00 AFTER duration_minutes;

UPDATE appointment_items ai
JOIN services s ON s.id = ai.service_id
SET ai.price = COALESCE(s.price, 0)
WHERE ai.service_id IS NOT NULL AND ai.price = 0;

-- 2. Una única factura por cita (evita duplicados si se pulsa dos veces).
ALTER TABLE invoices
  ADD UNIQUE KEY uq_invoices_appointment (appointment_id);
