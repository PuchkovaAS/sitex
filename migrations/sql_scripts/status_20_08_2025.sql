-- Вставляем типы статусов
INSERT INTO status_types (name, code, created_at, updated_at) VALUES
('В офисе', 'work_office', NOW(), NOW()),
('Удаленная работа', 'work_remote', NOW(), NOW()),
('Командировка', 'business_trip', NOW(), NOW()),
('Отпуск', 'vacation', NOW(), NOW()),
('Больничный', 'sick_leave', NOW(), NOW()),
('Выходной', 'weekend', NOW(), NOW()),
('Отгул', 'day_off', NOW(), NOW())
ON CONFLICT (code) DO NOTHING;

-- Проверяем что добавилось
SELECT * FROM status_types ORDER BY id;
