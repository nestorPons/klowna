package model

import (
	"math"
	"strconv"
	"strings"
)

// Theme contiene los colores personalizables del panel. Los valores son
// colores hexadecimales de 6 dígitos (#RRGGBB).
type Theme struct {
	Preset        string `json:"preset"`
	Bg            string `json:"bg"`
	Surface       string `json:"surface"`
	Surface2      string `json:"surface2"`
	Text          string `json:"text"`
	Muted         string `json:"muted"`
	Accent        string `json:"accent"`
	AccentPressed string `json:"accentPressed"`
	AccentSoft    string `json:"accentSoft"`
	Gold          string `json:"gold"`
	Sage          string `json:"sage"`
}

type ThemeVar struct {
	Name  string
	Value string
}

type ThemePreset struct {
	Slug        string
	Name        string
	Description string
	Theme       Theme
}

func themePresets() []ThemePreset {
	presets := []ThemePreset{
		{
			Slug:        "bordeaux",
			Name:        "Burdeos · Crema",
			Description: "Crema cálida con acento burdeos y detalles dorados.",
			Theme: Theme{
				Bg: "#faf6f0", Surface: "#fffdf9", Surface2: "#f2e8dd",
				Text: "#2e1b1f", Muted: "#73605c",
				Accent: "#75162d", AccentPressed: "#560b18", AccentSoft: "#f3e4e4",
				Gold: "#b8860b", Sage: "#bab8a2",
			},
		},
		{
			Slug:        "champagne",
			Name:        "Champagne · Oro",
			Description: "Beige champán, oro suave y marrón cálido.",
			Theme: Theme{
				Bg: "#fbf4e4", Surface: "#fffcf5", Surface2: "#f4e7c9",
				Text: "#3a2a1e", Muted: "#7a6650",
				Accent: "#8a6414", AccentPressed: "#6b4b0c", AccentSoft: "#f6ebcb",
				Gold: "#d9b45e", Sage: "#bab8a2",
			},
		},
		{
			Slug:        "latte",
			Name:        "Latte · Moca",
			Description: "Nudes suaves, espresso y piedra natural.",
			Theme: Theme{
				Bg: "#f7f2ed", Surface: "#fffdfb", Surface2: "#eee4da",
				Text: "#33261f", Muted: "#77645a",
				Accent: "#5c4336", AccentPressed: "#433026", AccentSoft: "#e8dbd1",
				Gold: "#b7a49c", Sage: "#b8ab9c",
			},
		},
		{
			Slug:        "rosewood",
			Name:        "Rosewood · Oliva",
			Description: "Rosa empolvado con burdeos profundo y oliva.",
			Theme: Theme{
				Bg: "#f9f4f0", Surface: "#fffdfb", Surface2: "#efe4de",
				Text: "#33262a", Muted: "#7a6565",
				Accent: "#7d2030", AccentPressed: "#580d1a", AccentSoft: "#f0dfdf",
				Gold: "#c4a69b", Sage: "#bab8a2",
			},
		},
		{
			Slug:        "noir",
			Name:        "Bordeaux Noir",
			Description: "Fondo oscuro con rosa empolvado y oro.",
			Theme: Theme{
				Bg: "#1a1114", Surface: "#24181c", Surface2: "#2f2025",
				Text: "#f4eae2", Muted: "#c0aea8",
				Accent: "#c4a69b", AccentPressed: "#a9877b", AccentSoft: "#3a272d",
				Gold: "#d8b36a", Sage: "#bab8a2",
			},
		},
	}
	for i := range presets {
		presets[i].Theme.Preset = presets[i].Slug
	}
	return presets
}

// ThemePresets devuelve las paletas predefinidas del panel.
func ThemePresets() []ThemePreset { return themePresets() }

// DefaultTheme devuelve la paleta base de Klowna.
func DefaultTheme() Theme { return themePresets()[0].Theme }

// PresetBySlug busca una paleta predefinida.
func PresetBySlug(slug string) (Theme, bool) {
	for _, p := range themePresets() {
		if p.Slug == slug {
			return p.Theme, true
		}
	}
	return Theme{}, false
}

// Normalize completa los valores vacíos o inválidos con la paleta base.
func (t Theme) Normalize() Theme {
	def := DefaultTheme()
	if t.Preset == "" {
		t.Preset = "custom"
	}
	t.Bg = validHex(t.Bg, def.Bg)
	t.Surface = validHex(t.Surface, def.Surface)
	t.Surface2 = validHex(t.Surface2, def.Surface2)
	t.Text = validHex(t.Text, def.Text)
	t.Muted = validHex(t.Muted, def.Muted)
	t.Accent = validHex(t.Accent, def.Accent)
	t.AccentPressed = validHex(t.AccentPressed, def.AccentPressed)
	t.AccentSoft = validHex(t.AccentSoft, def.AccentSoft)
	t.Gold = validHex(t.Gold, def.Gold)
	t.Sage = validHex(t.Sage, def.Sage)
	return t
}

// Vars devuelve las variables CSS que se inyectan en el panel.
func (t Theme) Vars() []ThemeVar {
	t = t.Normalize()
	return []ThemeVar{
		{"bg", t.Bg},
		{"surface", t.Surface},
		{"surface-2", t.Surface2},
		{"text", t.Text},
		{"muted", t.Muted},
		{"accent", t.Accent},
		{"accent-pressed", t.AccentPressed},
		{"accent-soft", t.AccentSoft},
		{"gold", t.Gold},
		{"sage", t.Sage},
		{"on-accent", contrastText(t.Accent)},
		{"scheme", schemeFor(t.Bg)},
	}
}

// VarsWithPrefix devuelve las variables listas para un bloque :root.
func (t Theme) VarsWithPrefix() []ThemeVar {
	vars := t.Vars()
	for i := range vars {
		vars[i].Name = "--" + vars[i].Name
	}
	return vars
}

func (t Theme) VarsMap() map[string]string {
	out := make(map[string]string, 12)
	for _, v := range t.Vars() {
		out[v.Name] = v.Value
	}
	return out
}

func validHex(v, fallback string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if len(v) != 7 || v[0] != '#' {
		return fallback
	}
	for _, r := range v[1:] {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return fallback
		}
	}
	return v
}

func schemeFor(hex string) string {
	if luminance(hex) < 0.24 {
		return "dark"
	}
	return "light"
}

func contrastText(hex string) string {
	if luminance(hex) > 0.24 {
		return "#2e1b1f"
	}
	return "#fffdf8"
}

func luminance(hex string) float64 {
	hex = strings.TrimPrefix(validHex(hex, "#000000"), "#")
	values, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return 0
	}
	r := float64((values>>16)&0xff) / 255
	g := float64((values>>8)&0xff) / 255
	b := float64(values&0xff) / 255
	channel := func(c float64) float64 {
		if c <= 0.03928 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(r) + 0.7152*channel(g) + 0.0722*channel(b)
}
