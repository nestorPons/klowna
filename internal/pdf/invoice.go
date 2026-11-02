package pdf

import (
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/signintech/gopdf"

	"klowna/internal/model"
)

//go:embed fonts/Carlito-Regular.ttf
var carlitoRegular []byte

//go:embed fonts/Carlito-Bold.ttf
var carlitoBold []byte

const (
	pageWidth  = 595.28
	pageHeight = 841.89
	margin     = 48.0
)

func Generate(billing model.BillingSettings, inv model.Invoice) ([]byte, error) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	if err := pdf.AddTTFFontDataWithOption("Carlito", carlitoRegular, gopdf.TtfOption{Style: gopdf.Regular}); err != nil {
		return nil, fmt.Errorf("add regular font: %w", err)
	}
	if err := pdf.AddTTFFontDataWithOption("Carlito", carlitoBold, gopdf.TtfOption{Style: gopdf.Bold}); err != nil {
		return nil, fmt.Errorf("add bold font: %w", err)
	}
	pdf.AddPage()

	const (
		accent  = uint8(117)
		accentG = uint8(22)
		accentB = uint8(45)
		gray    = uint8(115)
	)

	ink := func() { pdf.SetTextColor(0x2e, 0x1b, 0x1f) }
	ink()
	muted := func() { pdf.SetTextColor(gray, gray, gray) }

	// --- Cabecera: emisor (izquierda) y bloque FACTURA (derecha) ---
	x := margin
	y := margin

	if billing.CompanyName != "" {
		pdf.SetFont("Carlito", "B", 16)
		pdf.SetX(x)
		pdf.SetY(y)
		pdf.Cell(nil, billing.CompanyName)
		y += 22
	}

	pdf.SetFont("Carlito", "", 9)
	for _, line := range []string{billing.TaxID, billing.Address, billing.Phone, billing.Email} {
		if line == "" {
			continue
		}
		pdf.SetX(x)
		pdf.SetY(y)
		pdf.Cell(nil, line)
		y += 13
	}

	// Bloque FACTURA a la derecha
	rightX := pageWidth - margin
	pdf.SetFont("Carlito", "B", 20)
	label := "FACTURA"
	w, _ := pdf.MeasureTextWidth(label)
	pdf.SetX(rightX - w)
	pdf.SetY(margin)
	pdf.Cell(nil, label)

	pdf.SetFont("Carlito", "", 10)
	rightLines := []string{"Nº " + inv.Number, fecha(inv.IssuedAt)}
	ry := margin + 24
	for _, line := range rightLines {
		w, _ := pdf.MeasureTextWidth(line)
		pdf.SetX(rightX - w)
		pdf.SetY(ry)
		pdf.Cell(nil, line)
		ry += 14
	}

	// Línea separadora
	sepY := margin + 78
	pdf.SetLineWidth(0.8)
	pdf.SetStrokeColor(accent, accentG, accentB)
	pdf.Line(margin, sepY, pageWidth-margin, sepY)

	// --- Cliente ---
	cy := sepY + 24
	muted()
	pdf.SetFont("Carlito", "", 9)
	pdf.SetX(margin)
	pdf.SetY(cy)
	pdf.Cell(nil, "CLIENTE")
	cy += 14
	ink()
	pdf.SetFont("Carlito", "B", 12)
	pdf.SetX(margin)
	pdf.SetY(cy)
	pdf.Cell(nil, inv.PatientName)
	cy += 18

	// --- Tabla de líneas ---
	colX := []float64{margin, margin + 268, margin + 348, margin + 420, margin + 480}

	drawRow := func(rowY, rowH float64, cells []string, bold bool, fill bool) {
		if fill {
			pdf.SetFillColor(0xf4, 0xee, 0xe6)
			pdf.RectFromUpperLeftWithStyle(margin, rowY, pageWidth-2*margin, rowH, "F")
		}
		if bold {
			pdf.SetFont("Carlito", "B", 9)
		} else {
			pdf.SetFont("Carlito", "", 9)
		}
		for i, text := range cells {
			pdf.SetX(colX[i] + 4)
			pdf.SetY(rowY + 2)
			pdf.Cell(nil, text)
		}
	}

	headY := cy
	drawRow(headY, 18, []string{"Descripción", "Cant.", "Precio", "", "Importe"}, true, true)

	rowY := headY + 18
	items := inv.Items
	if len(items) == 0 {
		items = []model.InvoiceItem{{Description: inv.Concept, Quantity: 1, UnitPrice: inv.Subtotal}}
	}
	for _, it := range items {
		qty := it.Quantity
		if qty <= 0 {
			qty = 1
		}
		cells := []string{
			it.Description,
			fmt.Sprintf("%d", qty),
			money(it.UnitPrice),
			"",
			money(float64(qty) * it.UnitPrice),
		}
		drawRow(rowY, 16, cells, false, false)
		rowY += 16
	}

	// Línea de cierre de la tabla
	pdf.SetLineWidth(0.4)
	pdf.SetStrokeColor(0xd0, 0xc4, 0xb8)
	pdf.Line(margin, rowY, pageWidth-margin, rowY)
	rowY += 10

	// --- Totales (derecha) ---
	totalX := margin + 320
	drawTotal := func(label string, value string, bold bool) {
		if bold {
			pdf.SetFont("Carlito", "B", 11)
		} else {
			pdf.SetFont("Carlito", "", 10)
		}
		pdf.SetX(totalX)
		pdf.SetY(rowY)
		pdf.Cell(nil, label)
		w, _ := pdf.MeasureTextWidth(value)
		pdf.SetX(pageWidth - margin - w)
		pdf.Cell(nil, value)
		rowY += 16
	}
	drawTotal("Base imponible", money(inv.Subtotal), false)
	drawTotal(fmt.Sprintf("IVA (%.0f%%)", inv.VATRate), money(inv.VATAmount), false)
	pdf.SetLineWidth(0.4)
	pdf.Line(totalX, rowY-3, pageWidth-margin, rowY-3)
	drawTotal("Total", money(inv.Amount), true)

	// --- Estado ---
	if inv.PaidAt != nil {
		muted()
		pdf.SetFont("Carlito", "", 9)
		pdf.SetX(margin)
		pdf.SetY(rowY + 10)
		pdf.Cell(nil, "Pagada el "+fecha(*inv.PaidAt))
	}

	return pdf.GetBytesPdfReturnErr()
}

func fecha(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("02/01/2006")
}

func money(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	return strings.Replace(s, ".", ",", 1) + " €"
}
