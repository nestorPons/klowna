-- 012_billing.sql — Facturación: precios de servicios, líneas de factura, IVA y datos fiscales.

-- 1. Precio del catálogo de servicios
ALTER TABLE services
  ADD COLUMN price DECIMAL(10,2) NOT NULL DEFAULT 0.00 AFTER duration_minutes;

-- 2. Datos fiscales e IVA por gestor
CREATE TABLE IF NOT EXISTS billing_settings (
  user_id BIGINT UNSIGNED NOT NULL PRIMARY KEY,
  company_name VARCHAR(160) NOT NULL DEFAULT '',
  tax_id VARCHAR(40) NOT NULL DEFAULT '',
  address VARCHAR(255) NOT NULL DEFAULT '',
  phone VARCHAR(40) NOT NULL DEFAULT '',
  email VARCHAR(160) NOT NULL DEFAULT '',
  vat_rate DECIMAL(5,2) NOT NULL DEFAULT 21.00,
  updated_at DATETIME NOT NULL,
  CONSTRAINT fk_billing_settings_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. Vínculo con la cita y desglose de importes de la factura
ALTER TABLE invoices
  ADD COLUMN appointment_id BIGINT UNSIGNED NULL AFTER patient_id,
  ADD COLUMN subtotal DECIMAL(10,2) NOT NULL DEFAULT 0.00 AFTER concept,
  ADD COLUMN vat_rate DECIMAL(5,2) NOT NULL DEFAULT 0.00 AFTER subtotal,
  ADD COLUMN vat_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00 AFTER vat_rate,
  ADD KEY idx_invoices_appointment (appointment_id),
  ADD CONSTRAINT fk_invoices_appointment FOREIGN KEY (appointment_id) REFERENCES appointments(id) ON DELETE SET NULL;

-- 4. Líneas de detalle de cada factura
CREATE TABLE IF NOT EXISTS invoice_items (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  invoice_id BIGINT UNSIGNED NOT NULL,
  service_id BIGINT UNSIGNED NULL,
  description VARCHAR(255) NOT NULL,
  quantity INT NOT NULL DEFAULT 1,
  unit_price DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  vat_rate DECIMAL(5,2) NOT NULL DEFAULT 0.00,
  sort_order INT NOT NULL DEFAULT 0,
  KEY idx_invoice_items_invoice (invoice_id),
  KEY idx_invoice_items_service (service_id),
  CONSTRAINT fk_invoice_items_invoice FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE,
  CONSTRAINT fk_invoice_items_service FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 5. Las facturas existentes se interpretan como importe total con IVA incluido
UPDATE invoices
SET subtotal = ROUND(amount / 1.21, 2),
    vat_rate = 21.00,
    vat_amount = amount - ROUND(amount / 1.21, 2)
WHERE amount > 0;

INSERT INTO invoice_items (invoice_id, description, quantity, unit_price, vat_rate, sort_order)
SELECT id, concept, 1, subtotal, vat_rate, 0
FROM invoices
WHERE concept <> '' AND subtotal > 0;

-- 6. Datos fiscales de ejemplo para el gestor demo
INSERT IGNORE INTO billing_settings
  (user_id, company_name, tax_id, address, phone, email, vat_rate, updated_at)
SELECT u.id, 'Clínica Demo', 'B12345678', 'Calle Mayor 10, 28013 Madrid',
       '+34 910 000 000', u.email, 21.00, NOW()
FROM users u
WHERE u.email = 'demo@klowna.com';

-- 7. Precios orientativos para los servicios del gestor demo
UPDATE services s
JOIN users u ON u.id = s.user_id
SET s.price = CASE
  WHEN s.name LIKE '%láser%' THEN 120.00
  WHEN s.name LIKE '%capilar%' THEN 60.00
  WHEN s.name LIKE '%Primera visita%' THEN 50.00
  ELSE 45.00
END
WHERE u.email = 'demo@klowna.com' AND s.price = 0;
