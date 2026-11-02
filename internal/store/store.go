package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"klowna/internal/model"
)

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (st *Store) Settings(ctx context.Context) (model.Settings, error) {
	rows, err := st.db.QueryContext(ctx, `SELECT setting_key, setting_value FROM site_settings`)
	if err != nil {
		return model.Settings{}, fmt.Errorf("settings: %w", err)
	}
	defer rows.Close()

	values := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return model.Settings{}, fmt.Errorf("scan setting: %w", err)
		}
		values[key] = value
	}
	if err := rows.Err(); err != nil {
		return model.Settings{}, err
	}
	return model.Settings{
		SiteName:  values["site_name"],
		Tagline:   values["tagline"],
		Phone:     values["phone"],
		Email:     values["email"],
		WhatsApp:  values["whatsapp"],
		Address:   values["address"],
		LegalNote: values["legal_note"],
	}, nil
}

func (st *Store) Sectors(ctx context.Context) ([]model.Sector, error) {
	rows, err := st.db.QueryContext(ctx,
		`SELECT id, slug, name, icon, tagline, description, sort_order
		 FROM sectors ORDER BY sort_order, id`)
	if err != nil {
		return nil, fmt.Errorf("sectors: %w", err)
	}
	defer rows.Close()

	sectors := make([]model.Sector, 0, 3)
	for rows.Next() {
		var s model.Sector
		if err := rows.Scan(&s.ID, &s.Slug, &s.Name, &s.Icon, &s.Tagline, &s.Description, &s.SortOrder); err != nil {
			return nil, fmt.Errorf("scan sector: %w", err)
		}
		sectors = append(sectors, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	specs, err := st.Specialties(ctx)
	if err != nil {
		return nil, err
	}
	index := make(map[int64]int, len(sectors))
	for i := range sectors {
		index[sectors[i].ID] = i
	}
	for _, spec := range specs {
		if i, ok := index[spec.SectorID]; ok {
			sectors[i].Specialties = append(sectors[i].Specialties, spec)
		}
	}
	return sectors, nil
}

func (st *Store) SectorBySlug(ctx context.Context, slug string) (model.Sector, error) {
	var s model.Sector
	err := st.db.QueryRowContext(ctx,
		`SELECT id, slug, name, icon, tagline, description, sort_order
		 FROM sectors WHERE slug = ?`, slug).
		Scan(&s.ID, &s.Slug, &s.Name, &s.Icon, &s.Tagline, &s.Description, &s.SortOrder)
	if err != nil {
		return model.Sector{}, err
	}
	specs, err := st.Specialties(ctx)
	if err != nil {
		return model.Sector{}, err
	}
	for _, spec := range specs {
		if spec.SectorID == s.ID {
			s.Specialties = append(s.Specialties, spec)
		}
	}
	return s, nil
}

func (st *Store) Specialties(ctx context.Context) ([]model.Specialty, error) {
	rows, err := st.db.QueryContext(ctx,
		`SELECT sp.id, sp.sector_id, se.slug, sp.slug, sp.name, sp.description, sp.sort_order
		 FROM specialties sp
		 JOIN sectors se ON se.id = sp.sector_id
		 ORDER BY se.sort_order, sp.sort_order, sp.id`)
	if err != nil {
		return nil, fmt.Errorf("specialties: %w", err)
	}
	defer rows.Close()

	out := make([]model.Specialty, 0, 20)
	for rows.Next() {
		var s model.Specialty
		if err := rows.Scan(&s.ID, &s.SectorID, &s.SectorSlug, &s.Slug, &s.Name, &s.Description, &s.SortOrder); err != nil {
			return nil, fmt.Errorf("scan specialty: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (st *Store) SpecialtyBySlug(ctx context.Context, sectorSlug, slug string) (model.Specialty, error) {
	var s model.Specialty
	err := st.db.QueryRowContext(ctx,
		`SELECT sp.id, sp.sector_id, se.slug, sp.slug, sp.name, sp.description, sp.sort_order
		 FROM specialties sp
		 JOIN sectors se ON se.id = sp.sector_id
		 WHERE se.slug = ? AND sp.slug = ?`, sectorSlug, slug).
		Scan(&s.ID, &s.SectorID, &s.SectorSlug, &s.Slug, &s.Name, &s.Description, &s.SortOrder)
	return s, err
}

func (st *Store) Solutions(ctx context.Context) ([]model.Solution, error) {
	rows, err := st.db.QueryContext(ctx,
		`SELECT id, slug, name, description, icon, sort_order
		 FROM solutions ORDER BY sort_order, id`)
	if err != nil {
		return nil, fmt.Errorf("solutions: %w", err)
	}
	defer rows.Close()

	out := make([]model.Solution, 0, 9)
	for rows.Next() {
		var s model.Solution
		if err := rows.Scan(&s.ID, &s.Slug, &s.Name, &s.Description, &s.Icon, &s.SortOrder); err != nil {
			return nil, fmt.Errorf("scan solution: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (st *Store) Challenges(ctx context.Context) ([]model.Challenge, error) {
	rows, err := st.db.QueryContext(ctx,
		`SELECT id, title, description, icon, sort_order
		 FROM challenges ORDER BY sort_order, id`)
	if err != nil {
		return nil, fmt.Errorf("challenges: %w", err)
	}
	defer rows.Close()

	out := make([]model.Challenge, 0, 6)
	for rows.Next() {
		var c model.Challenge
		if err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.Icon, &c.SortOrder); err != nil {
			return nil, fmt.Errorf("scan challenge: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (st *Store) Stats(ctx context.Context) ([]model.Stat, error) {
	rows, err := st.db.QueryContext(ctx,
		`SELECT id, value, label, sort_order FROM stats ORDER BY sort_order, id`)
	if err != nil {
		return nil, fmt.Errorf("stats: %w", err)
	}
	defer rows.Close()

	out := make([]model.Stat, 0, 3)
	for rows.Next() {
		var s model.Stat
		if err := rows.Scan(&s.ID, &s.Value, &s.Label, &s.SortOrder); err != nil {
			return nil, fmt.Errorf("scan stat: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (st *Store) Testimonials(ctx context.Context) ([]model.Testimonial, error) {
	rows, err := st.db.QueryContext(ctx,
		`SELECT id, sector, company, quote, url, sort_order
		 FROM testimonials ORDER BY sort_order, id`)
	if err != nil {
		return nil, fmt.Errorf("testimonials: %w", err)
	}
	defer rows.Close()

	out := make([]model.Testimonial, 0, 3)
	for rows.Next() {
		var t model.Testimonial
		if err := rows.Scan(&t.ID, &t.Sector, &t.Company, &t.Quote, &t.URL, &t.SortOrder); err != nil {
			return nil, fmt.Errorf("scan testimonial: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (st *Store) Posts(ctx context.Context, limit int) ([]model.Post, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := st.db.QueryContext(ctx,
		`SELECT id, slug, title, excerpt, category, image_url, published_at
		 FROM posts ORDER BY published_at DESC, id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("posts: %w", err)
	}
	defer rows.Close()

	out := make([]model.Post, 0, limit)
	for rows.Next() {
		var p model.Post
		if err := rows.Scan(&p.ID, &p.Slug, &p.Title, &p.Excerpt, &p.Category, &p.ImageURL, &p.PublishedAt); err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (st *Store) PostBySlug(ctx context.Context, slug string) (model.Post, error) {
	var p model.Post
	err := st.db.QueryRowContext(ctx,
		`SELECT id, slug, title, excerpt, category, body, image_url, published_at
		 FROM posts WHERE slug = ?`, slug).
		Scan(&p.ID, &p.Slug, &p.Title, &p.Excerpt, &p.Category, &p.Body, &p.ImageURL, &p.PublishedAt)
	return p, err
}

func (st *Store) Plans(ctx context.Context) ([]model.Plan, error) {
	rows, err := st.db.QueryContext(ctx,
		`SELECT id, name, price, period, description, features_json, highlighted, sort_order
		 FROM plans ORDER BY sort_order, id`)
	if err != nil {
		return nil, fmt.Errorf("plans: %w", err)
	}
	defer rows.Close()

	out := make([]model.Plan, 0, 3)
	for rows.Next() {
		var p model.Plan
		var features string
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Period, &p.Description, &features, &p.Highlighted, &p.SortOrder); err != nil {
			return nil, fmt.Errorf("scan plan: %w", err)
		}
		if err := json.Unmarshal([]byte(features), &p.Features); err != nil {
			return nil, fmt.Errorf("decode features for plan %s: %w", p.Name, err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (st *Store) FAQs(ctx context.Context) ([]model.FAQ, error) {
	rows, err := st.db.QueryContext(ctx,
		`SELECT id, question, answer, sort_order FROM faqs ORDER BY sort_order, id`)
	if err != nil {
		return nil, fmt.Errorf("faqs: %w", err)
	}
	defer rows.Close()

	out := make([]model.FAQ, 0, 6)
	for rows.Next() {
		var f model.FAQ
		if err := rows.Scan(&f.ID, &f.Question, &f.Answer, &f.SortOrder); err != nil {
			return nil, fmt.Errorf("scan faq: %w", err)
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (st *Store) CreateLead(ctx context.Context, l model.Lead) (int64, error) {
	res, err := st.db.ExecContext(ctx,
		`INSERT INTO leads (name, email, phone, clinic, sector, message, consent, source, ip, user_agent, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		l.Name, l.Email, l.Phone, l.Clinic, l.Sector, l.Message, l.Consent, l.Source, l.IP, l.UserAgent, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("create lead: %w", err)
	}
	return res.LastInsertId()
}

func (st *Store) Leads(ctx context.Context, limit int) ([]model.Lead, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := st.db.QueryContext(ctx,
		`SELECT id, name, email, phone, clinic, sector, message, consent, source, created_at
		 FROM leads ORDER BY created_at DESC, id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("leads: %w", err)
	}
	defer rows.Close()

	out := make([]model.Lead, 0, limit)
	for rows.Next() {
		var l model.Lead
		if err := rows.Scan(&l.ID, &l.Name, &l.Email, &l.Phone, &l.Clinic, &l.Sector, &l.Message, &l.Consent, &l.Source, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan lead: %w", err)
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func (st *Store) Ping(ctx context.Context) error {
	return st.db.PingContext(ctx)
}
