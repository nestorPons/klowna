package model

import "testing"

func TestSummarizeCash(t *testing.T) {
	movements := []CashMovement{
		{MovementType: "ingreso", PaymentMethod: "efectivo", Amount: 100},
		{MovementType: "ingreso", PaymentMethod: "tarjeta", Amount: 50},
		{MovementType: "gasto", PaymentMethod: "efectivo", Amount: 20},
		{MovementType: "gasto", PaymentMethod: "transferencia", Amount: 5},
	}
	s := SummarizeCash(movements, 200, 280)

	if s.CashIncome != 100 || s.OtherIncome != 50 {
		t.Fatalf("income = %v/%v, want 100/50", s.CashIncome, s.OtherIncome)
	}
	if s.CashExpenses != 20 || s.OtherExpenses != 5 {
		t.Fatalf("expenses = %v/%v, want 20/5", s.CashExpenses, s.OtherExpenses)
	}
	if s.ExpectedCash != 280 {
		t.Fatalf("expected cash = %v, want 280", s.ExpectedCash)
	}
	if s.Difference != 0 {
		t.Fatalf("difference = %v, want 0", s.Difference)
	}
	if s.TotalIncome != 150 || s.TotalExpenses != 25 {
		t.Fatalf("totals = %v/%v, want 150/25", s.TotalIncome, s.TotalExpenses)
	}
}

func TestCashTotal(t *testing.T) {
	denoms := []CashDenomination{
		{Key: "b50", Value: 50, Count: 2},
		{Key: "m2", Value: 2, Count: 3},
		{Key: "m050", Value: 0.50, Count: 2},
	}
	if got := CashTotal(denoms); got != 107 {
		t.Fatalf("CashTotal = %v, want 107", got)
	}
}

func TestAppointmentStatusLabel(t *testing.T) {
	if got := AppointmentStatusLabel("no_asistida"); got != "No asistida" {
		t.Fatalf("status label = %q", got)
	}
	if got := PaymentMethodLabel("transferencia"); got != "Transferencia bancaria" {
		t.Fatalf("payment label = %q", got)
	}
}

func TestAppointmentDeposit(t *testing.T) {
	a := Appointment{DepositDue: 20, DepositPaid: 5}
	if !a.DepositPending() {
		t.Fatal("la señal debería estar pendiente")
	}
	if got := a.DepositRemaining(); got != 15 {
		t.Fatalf("remaining = %v, want 15", got)
	}

	a.DepositPaid = 20
	if a.DepositPending() {
		t.Fatal("la señal no debería estar pendiente")
	}
	if got := a.DepositRemaining(); got != 0 {
		t.Fatalf("remaining = %v, want 0", got)
	}

	if (Appointment{}).DepositPending() {
		t.Fatal("sin señal exigida no hay pendiente")
	}
}
