package pdf

import (
	"bytes"
	"testing"
	"time"

	"klowna/internal/model"
)

func TestGenerate(t *testing.T) {
	now := time.Now()
	paid := now.Add(-24 * time.Hour)
	inv := model.Invoice{
		Number:     "F-2026-0001",
		PatientName: "Laura García",
		Concept:    "Consulta láser",
		Subtotal:   100,
		VATRate:    21,
		VATAmount:  21,
		Amount:     121,
		Status:     "pagada",
		IssuedAt:   now,
		PaidAt:     &paid,
		Items: []model.InvoiceItem{
			{Description: "Consulta láser", Quantity: 1, UnitPrice: 100, VATRate: 21},
			{Description: "Revisión capilar", Quantity: 2, UnitPrice: 60, VATRate: 21},
		},
	}
	billing := model.BillingSettings{
		CompanyName: "Clínica Demo S.L.",
		TaxID:       "B12345678",
		Address:     "Calle Mayor 10, 28013 Madrid",
		Phone:       "+34 910 000 000",
		Email:       "demo@klowna.com",
		VATRate:     21,
	}
	data, err := Generate(billing, inv)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty PDF")
	}
	if !bytes.HasPrefix(data, []byte("%PDF")) {
		t.Fatalf("no PDF header: %q", data[:8])
	}
}
