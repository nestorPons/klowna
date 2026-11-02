package model

import "time"

type Settings struct {
	SiteName  string
	Tagline   string
	Phone     string
	Email     string
	WhatsApp  string
	Address   string
	LegalNote string
}

type Sector struct {
	ID          int64
	Slug        string
	Name        string
	Icon        string
	Tagline     string
	Description string
	SortOrder   int
	Specialties []Specialty
}

type Specialty struct {
	ID          int64
	SectorID    int64
	SectorSlug  string
	Slug        string
	Name        string
	Description string
	SortOrder   int
}

type Solution struct {
	ID          int64
	Slug        string
	Name        string
	Description string
	Icon        string
	SortOrder   int
}

type Challenge struct {
	ID          int64
	Title       string
	Description string
	Icon        string
	SortOrder   int
}

type Stat struct {
	ID        int64
	Value     string
	Label     string
	SortOrder int
}

type Testimonial struct {
	ID        int64
	Sector    string
	Company   string
	Quote     string
	URL       string
	SortOrder int
}

type Post struct {
	ID          int64
	Slug        string
	Title       string
	Excerpt     string
	Category    string
	Body        string
	ImageURL    string
	PublishedAt time.Time
}

type Plan struct {
	ID          int64
	Name        string
	Price       string
	Period      string
	Description string
	Features    []string
	Highlighted bool
	SortOrder   int
}

type FAQ struct {
	ID        int64
	Question  string
	Answer    string
	SortOrder int
}

type Lead struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Clinic  string `json:"clinic"`
	Sector  string `json:"sector"`
	Message string `json:"message"`
	Consent bool   `json:"consent"`
	Website string `json:"website"`
	Source  string `json:"source"`
	// CaptchaToken viaja en "cf-turnstile-response"; no se guarda en la base.
	CaptchaToken string    `json:"cf-turnstile-response"`
	IP           string    `json:"-"`
	UserAgent    string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}
