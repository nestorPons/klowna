-- 009_rebrand.sql — Renombrado de clowww a AiTurnia.

UPDATE site_settings SET setting_value = REPLACE(setting_value, 'clowww', 'AiTurnia');

UPDATE solutions
SET name = REPLACE(name, 'clowww', 'AiTurnia'),
    description = REPLACE(description, 'clowww', 'AiTurnia');

UPDATE testimonials SET quote = REPLACE(quote, 'clowww', 'AiTurnia');

UPDATE posts
SET slug = REPLACE(slug, 'clowww', 'aiturnia'),
    title = REPLACE(title, 'clowww', 'AiTurnia'),
    excerpt = REPLACE(excerpt, 'clowww', 'AiTurnia'),
    body = REPLACE(body, 'clowww', 'AiTurnia');

UPDATE plans SET description = REPLACE(description, 'clowww', 'AiTurnia');

UPDATE faqs
SET question = REPLACE(question, 'clowww', 'AiTurnia'),
    answer = REPLACE(answer, 'clowww', 'AiTurnia');

UPDATE users SET email = REPLACE(email, 'clowww.es', 'aiturnia.es')
WHERE email LIKE '%@clowww.es';

UPDATE users
SET password_hash = '$2a$10$zIi7MZ/zp2oF0Qb.vAth.uy0BwCjxMx4klmeKAFcjmjJRs41ctw6q'
WHERE email IN (
  'demo@aiturnia.es',
  'ana@aiturnia.es',
  'carlos@aiturnia.es',
  'lucia@aiturnia.es'
);
