package render

// ServiceIconNames lista los iconos disponibles para el catálogo de servicios.
var ServiceIconNames = []string{
	"num-0", "num-1", "num-2", "num-3", "num-4", "num-5", "num-6", "num-7", "num-8", "num-9",
	"sparkles", "star", "heart", "leaf", "flame", "sun", "droplet",
	"eye", "smile", "scissors", "zap", "stethoscope", "clock",
}

// numberIcons mapea los iconos numéricos a su dígito.
var numberIcons = map[string]string{
	"num-0": "0", "num-1": "1", "num-2": "2", "num-3": "3", "num-4": "4", "num-5": "5",
	"num-6": "6", "num-7": "7", "num-8": "8", "num-9": "9",
}

// NumberIconDigit devuelve el dígito de un icono numérico (num-0..num-9).
func NumberIconDigit(name string) (string, bool) {
	d, ok := numberIcons[name]
	return d, ok
}

var iconPaths = map[string]string{
	"chevron":        `<path d="m6 9 6 6 6-6"/>`,
	"check":          `<path d="m5 13 4 4L19 7"/>`,
	"arrow-right":    `<path d="M5 12h14"/><path d="m13 6 6 6-6 6"/>`,
	"arrow-left":     `<path d="M19 12H5"/><path d="m11 6-6 6 6 6"/>`,
	"sparkles":       `<path d="M12 3l1.8 4.7L18.5 9.5l-4.7 1.8L12 16l-1.8-4.7L5.5 9.5l4.7-1.8Z"/><path d="M19 15l.9 2.1L22 18l-2.1.9L19 21l-.9-2.1L16 18l2.1-.9Z"/>`,
	"leaf":           `<path d="M20 4c-9 0-15 4-15 11a5 5 0 0 0 5 5c7 0 10-6 10-16Z"/><path d="M4 21c2-6 6-9 12-12"/>`,
	"heart":          `<path d="M12 20s-7-4.3-7-9.5A4 4 0 0 1 12 7a4 4 0 0 1 7 3.5C19 15.7 12 20 12 20Z"/>`,
	"cloud":          `<path d="M7 18a4 4 0 0 1-.6-8 6 6 0 0 1 11.5 1.5A3.5 3.5 0 0 1 17.5 18Z"/>`,
	"smartphone":     `<rect x="7" y="2" width="10" height="20" rx="2"/><path d="M11 18h2"/>`,
	"message-circle": `<path d="M21 12a8 8 0 0 1-11.7 7.1L4 21l1.9-5.3A8 8 0 1 1 21 12Z"/>`,
	"stethoscope":    `<path d="M6 3v5a4 4 0 0 0 8 0V3"/><path d="M10 12v3a5 5 0 0 0 10 0v-2"/><circle cx="20" cy="11" r="2"/>`,
	"globe":          `<circle cx="12" cy="12" r="9"/><path d="M3 12h18"/><path d="M12 3a14 14 0 0 1 0 18 14 14 0 0 1 0-18Z"/>`,
	"users":          `<circle cx="9" cy="8" r="3"/><path d="M3 20a6 6 0 0 1 12 0"/><path d="M16 5a3 3 0 0 1 0 6"/><path d="M18 20a6 6 0 0 0-3-5.2"/>`,
	"box":            `<path d="M12 3 3 7.5v9L12 21l9-4.5v-9Z"/><path d="M3 7.5 12 12l9-4.5"/><path d="M12 12v9"/>`,
	"calendar":       `<rect x="3" y="5" width="18" height="16" rx="2"/><path d="M8 3v4M16 3v4M3 10h18"/>`,
	"share":          `<circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><path d="m8.6 10.5 6.8-4M8.6 13.5l6.8 4"/>`,
	"file":           `<path d="M14 3H6a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9Z"/><path d="M14 3v6h6M8 13h8M8 17h5"/>`,
	"note":           `<path d="M4 4h16v11l-5 5H4Z"/><path d="M20 15h-5v5M8 8h8M8 12h5"/>`,
	"megaphone":      `<path d="M3 11v2a2 2 0 0 0 2 2h2l7 4V5L7 9H5a2 2 0 0 0-2 2Z"/><path d="M18 9a4 4 0 0 1 0 6"/>`,
	"card":           `<rect x="2" y="5" width="20" height="14" rx="2"/><path d="M2 10h20M6 15h4"/>`,
	"chart":          `<path d="M4 20V10M10 20V4M16 20v-7M22 20H2"/>`,
	"quote":          `<path d="M7 7h4v4c0 3-2 5-5 5M15 7h4v4c0 3-2 5-5 5"/>`,
	"phone":          `<path d="M5 3h3l2 5-2.5 1.5a11 11 0 0 0 6 6L15 13l5 2v3a2 2 0 0 1-2 2A16 16 0 0 1 3 5a2 2 0 0 1 2-2Z"/>`,
	"mail":           `<rect x="3" y="5" width="18" height="14" rx="2"/><path d="m3 7 9 6 9-6"/>`,
	"clock":          `<circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/>`,
	"whatsapp":       `<path d="M21 12a8 8 0 0 1-11.7 7.1L4 21l1.9-5.3A8 8 0 1 1 21 12Z"/><path d="M9 9c0 3 3 6 6 6l1-2-2-1-1 1c-1-.5-2-1.5-2.5-2.5l1-1-1-2Z"/>`,
	"map-pin":        `<path d="M12 21s7-6 7-11a7 7 0 1 0-14 0c0 5 7 11 7 11Z"/><circle cx="12" cy="10" r="2.5"/>`,
	"grid":           `<rect x="3" y="3" width="7" height="7" rx="1.5"/><rect x="14" y="3" width="7" height="7" rx="1.5"/><rect x="3" y="14" width="7" height="7" rx="1.5"/><rect x="14" y="14" width="7" height="7" rx="1.5"/>`,
	"plus":           `<path d="M12 5v14M5 12h14"/>`,
	"search":         `<circle cx="11" cy="11" r="7"/><path d="m21 21-4.3-4.3"/>`,
	"logout":         `<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><path d="m16 17 5-5-5-5"/><path d="M21 12H9"/>`,
	"menu":           `<path d="M4 6h16M4 12h16M4 18h16"/>`,
	"euro":           `<path d="M18 6a7 7 0 1 0 0 12"/><path d="M4 10h9M4 14h9"/>`,
	"inbox":          `<path d="M4 4h16l2 5v9a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V9Z"/><path d="M2 9h5l2 3h6l2-3h5"/>`,
	"user":           `<circle cx="12" cy="8" r="4"/><path d="M4 21a8 8 0 0 1 16 0"/>`,
	"x":              `<path d="M18 6 6 18M6 6l12 12"/>`,
	"trash":          `<path d="M4 7h16"/><path d="M9 7V5a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/><path d="M6 7l1 13a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2l1-13"/><path d="M10 11v6M14 11v6"/>`,
	"settings":       `<circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.7 1.7 0 0 0 .34 1.87l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.7 1.7 0 0 0-1.87-.34 1.7 1.7 0 0 0-1 1.55V21a2 2 0 1 1-4 0v-.09a1.7 1.7 0 0 0-1.11-1.55 1.7 1.7 0 0 0-1.87.34l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.7 1.7 0 0 0 .34-1.87 1.7 1.7 0 0 0-1.55-1H3a2 2 0 1 1 0-4h.09A1.7 1.7 0 0 0 4.64 8a1.7 1.7 0 0 0-.34-1.87l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06A1.7 1.7 0 0 0 9 3.64h.08A1.7 1.7 0 0 0 10 2.09V2a2 2 0 1 1 4 0v.09a1.7 1.7 0 0 0 1 1.55 1.7 1.7 0 0 0 1.87-.34l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.7 1.7 0 0 0-.34 1.87V8a1.7 1.7 0 0 0 1.55 1H21a2 2 0 1 1 0 4h-.09a1.7 1.7 0 0 0-1.51 1Z"/>`,
	"palette":        `<path d="M12 3a9 9 0 1 0 0 18h1.5a2 2 0 0 0 0-4H12a2 2 0 0 1 0-4h4.5A4.5 4.5 0 0 0 21 8.5 5.5 5.5 0 0 0 15.5 3Z"/><circle cx="7.5" cy="10.5" r="1"/><circle cx="9.5" cy="6.5" r="1"/><circle cx="14" cy="6.5" r="1"/><circle cx="17" cy="10" r="1"/>`,
	"star":           `<path d="M12 3l2.7 5.6 6.3.9-4.5 4.4 1 6.3-5.5-2.9-5.5 2.9 1-6.3L3 9.5l6.3-.9Z"/>`,
	"scissors":       `<circle cx="6" cy="6" r="3"/><circle cx="6" cy="18" r="3"/><path d="M20 4 8.5 15.5"/><path d="M14.5 14.5 20 20"/><path d="M8.5 8.5 12 12"/>`,
	"droplet":        `<path d="M12 3s6 6.5 6 11a6 6 0 0 1-12 0c0-4.5 6-11 6-11Z"/>`,
	"eye":            `<path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7-10-7-10-7Z"/><circle cx="12" cy="12" r="3"/>`,
	"smile":          `<circle cx="12" cy="12" r="9"/><path d="M8 14s1.5 2 4 2 4-2 4-2"/><path d="M9 9h.01M15 9h.01"/>`,
	"sun":            `<circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4"/>`,
	"zap":            `<path d="M13 2 4 14h6l-1 8 9-12h-6l1-8Z"/>`,
	"flame":          `<path d="M12 2c1 3 4 4.5 4 8a4 4 0 0 1-8 0c0-1.5.5-2.5 1-3.5 0 1 .5 1.5 1 1.5C9 6 10 4 12 2Z"/>`,
}

// solidIconPaths contiene las versiones rellenas de los iconos de servicios y equipo.
var solidIconPaths = map[string]string{
	"sparkles":    `<path d="M12 2.5l2 5.2 5.2 2-5.2 2-2 5.2-2-5.2-5.2-2 5.2-2Z"/><path d="M19 14l1 2.6 2.6 1-2.6 1-1 2.6-1-2.6-2.6-1 2.6-1Z"/>`,
	"star":        `<path d="M12 3l2.7 5.6 6.3.9-4.5 4.4 1 6.3-5.5-2.9-5.5 2.9 1-6.3L3 9.5l6.3-.9Z"/>`,
	"heart":       `<path d="M12 20s-7-4.3-7-9.5A4 4 0 0 1 12 7a4 4 0 0 1 7 3.5C19 15.7 12 20 12 20Z"/>`,
	"leaf":        `<path d="M20 4c-9 0-15 4-15 11a5 5 0 0 0 5 5c7 0 10-6 10-16Z"/>`,
	"flame":       `<path d="M12 2c1 3 4 4.5 4 8a4 4 0 0 1-8 0c0-1.5.5-2.5 1-3.5 0 1 .5 1.5 1 1.5C9 6 10 4 12 2Z"/>`,
	"sun":         `<circle cx="12" cy="12" r="4.2"/><path d="M12 1.5v3M12 19.5v3M1.5 12h3M19.5 12h3M4.2 4.2l2.1 2.1M17.7 17.7l2.1 2.1M19.8 4.2l-2.1 2.1M6.3 17.7l-2.1 2.1" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" fill="none"/>`,
	"droplet":     `<path d="M12 3s6 6.5 6 11a6 6 0 0 1-12 0c0-4.5 6-11 6-11Z"/>`,
	"eye":         `<path fill-rule="evenodd" d="M12 4.5C5.5 4.5 2 12 2 12s3.5 7.5 10 7.5S22 12 22 12 18.5 4.5 12 4.5Zm0 3a4.5 4.5 0 1 0 0 9 4.5 4.5 0 0 0 0-9Z"/>`,
	"smile":       `<path fill-rule="evenodd" d="M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20ZM9 9a1.2 1.2 0 1 1 0 2.4A1.2 1.2 0 0 1 9 9Zm6 0a1.2 1.2 0 1 1 0 2.4A1.2 1.2 0 0 1 15 9Zm-6.7 4.4h7.4c.5 0 .9.4.8.9-.5 2.4-2.4 4.2-4.5 4.2s-4-1.8-4.5-4.2c-.1-.5.3-.9.8-.9Z"/>`,
	"scissors":    `<circle cx="6" cy="6" r="3"/><circle cx="6" cy="18" r="3"/><path d="M8.5 8.5L20 20M20 4 8.5 15.5" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" fill="none"/>`,
	"zap":         `<path d="M13 2 4 14h6l-1 8 9-12h-6l1-8Z"/>`,
	"stethoscope": `<path d="M6 3v5a4 4 0 0 0 8 0V3M10 12v3a5 5 0 0 0 10 0v-2" stroke="currentColor" stroke-width="2.6" stroke-linecap="round" fill="none"/><circle cx="20" cy="11" r="2.6"/>`,
	"clock":       `<path fill-rule="evenodd" d="M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20Zm-1 5a1 1 0 0 1 2 0v4.6l2.7 1.6a1 1 0 1 1-1 1.7l-3.2-1.9a1 1 0 0 1-.5-.9Z"/>`,
	"user":        `<circle cx="12" cy="8" r="4"/><path d="M4 21a8 8 0 0 1 16 0Z"/>`,
}
