-- 010_rebrand_klowna.sql — Renombrado de AiTurnia a Klowna.

UPDATE site_settings
SET setting_value = REPLACE(REPLACE(setting_value, 'AiTurnia', 'Klowna'), 'aiturnia', 'klowna');

UPDATE solutions
SET name = REPLACE(REPLACE(name, 'AiTurnia', 'Klowna'), 'aiturnia', 'klowna'),
    description = REPLACE(REPLACE(description, 'AiTurnia', 'Klowna'), 'aiturnia', 'klowna');

UPDATE testimonials
SET quote = REPLACE(REPLACE(quote, 'AiTurnia', 'Klowna'), 'aiturnia', 'klowna');

UPDATE posts
SET slug = REPLACE(slug, 'aiturnia', 'klowna'),
    title = REPLACE(REPLACE(title, 'AiTurnia', 'Klowna'), 'aiturnia', 'klowna'),
    excerpt = REPLACE(REPLACE(excerpt, 'AiTurnia', 'Klowna'), 'aiturnia', 'klowna'),
    body = REPLACE(REPLACE(body, 'AiTurnia', 'Klowna'), 'aiturnia', 'klowna');

UPDATE plans
SET description = REPLACE(REPLACE(description, 'AiTurnia', 'Klowna'), 'aiturnia', 'klowna');

UPDATE faqs
SET question = REPLACE(REPLACE(question, 'AiTurnia', 'Klowna'), 'aiturnia', 'klowna'),
    answer = REPLACE(REPLACE(answer, 'AiTurnia', 'Klowna'), 'aiturnia', 'klowna');

UPDATE users
SET email = REPLACE(email, '@aiturnia.es', '@klowna.com')
WHERE email LIKE '%@aiturnia.es';

UPDATE users
SET password_hash = '$2a$10$jbVZX8CVGijvTnuyVTp99ONz4o0o5FSLnBzK58JOfUj2K/3qwlO.6'
WHERE email IN (
  'demo@klowna.com',
  'prueba@klowna.com',
  'ana@klowna.com',
  'carlos@klowna.com',
  'lucia@klowna.com'
);
