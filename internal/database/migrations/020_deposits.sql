-- 020_deposits.sql — Señal de reserva por servicio y su registro en caja.

-- 1. Cada servicio puede exigir una señal de reserva.
ALTER TABLE services
  ADD COLUMN deposit_required TINYINT(1) NOT NULL DEFAULT 0 AFTER price,
  ADD COLUMN deposit_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00 AFTER deposit_required;

-- 2. La cita congela la señal exigida y acumula lo abonado como señal.
--    deposit_paid_at y deposit_payment_method guardan el último cobro de señal.
ALTER TABLE appointments
  ADD COLUMN deposit_due DECIMAL(10,2) NOT NULL DEFAULT 0.00 AFTER notes,
  ADD COLUMN deposit_paid DECIMAL(10,2) NOT NULL DEFAULT 0.00 AFTER deposit_due,
  ADD COLUMN deposit_payment_method VARCHAR(24) NOT NULL DEFAULT '' AFTER deposit_paid,
  ADD COLUMN deposit_paid_at DATETIME NULL AFTER deposit_payment_method,
  ADD KEY idx_appointments_deposit (user_id, deposit_due);
