-- Создаем пользователя a@a.ru (пароль: 123456)

INSERT INTO employees (first_name, last_name, email, password_hash, role, position, department, is_active, created_at, updated_at) 
VALUES (
    'Test', 
    'User', 
    'a@a.ru', 
    '$2a$10$rL.6zJNWe2b6b8b1b6b1b.b1b6b1b6b1b6b1b6b1b6b1b6b1b6b1b6', -- хэш для "123456"
    'employee', 
    'Тестовая должность', 
    'Тестовый отдел', 
    TRUE,
    NOW(),
    NOW()
)
ON CONFLICT (email) DO NOTHING;

-- Проверяем пользователя
SELECT id, first_name, last_name, email, role, is_active FROM employees WHERE email = 'a@a.ru';

