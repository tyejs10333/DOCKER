
BEGIN;
INSERT INTO users (id, max_user_id, first_name, last_name, role, teacher_verified)
VALUES
  (1, -900001, 'Тестовый', 'Ученик 1', 'student', FALSE),
  (2, -900002, 'Тестовый', 'Ученик 2', 'student', FALSE)
ON CONFLICT (id) DO NOTHING;
SELECT setval(pg_get_serial_sequence('users', 'id'),
              GREATEST((SELECT COALESCE(MAX(id), 1) FROM users), 1), TRUE);
COMMIT;
