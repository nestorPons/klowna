-- 021_user_avatar.sql — Avatar editable desde el panel de conectados.

ALTER TABLE users
  ADD COLUMN avatar MEDIUMTEXT NOT NULL AFTER icon;
