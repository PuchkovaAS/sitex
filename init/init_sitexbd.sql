-- Подключаемся к postgres для создания базы (если нужно)
\c postgres

-- Создаём базу, если не существует
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'sitex_db') THEN
        CREATE DATABASE sitex_db;
    END IF;
END $$;

-- Переключаемся на новую базу
\c sitex_db

-- Создаём таблицы и связи

-- Таблица departments
CREATE TABLE IF NOT EXISTS departments (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_department_name ON departments(name);

-- Таблица employees
CREATE TABLE IF NOT EXISTS employees (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    position VARCHAR(255),
    department_id BIGINT REFERENCES departments(id) ON DELETE SET NULL,
    is_active BOOLEAN DEFAULT true,
    is_admin BOOLEAN DEFAULT false,
    who_added_id BIGINT -- будет ссылаться на employees.id, добавим позже после вставки
);

-- Индексы для employees
CREATE INDEX IF NOT EXISTS idx_employees_email ON employees(email);
CREATE INDEX IF NOT EXISTS idx_employees_active ON employees(is_active);
CREATE INDEX IF NOT EXISTS idx_employees_admin ON employees(is_admin);
CREATE INDEX IF NOT EXISTS idx_employees_department_id ON employees(department_id);
CREATE INDEX IF NOT EXISTS idx_employees_who_added_id ON employees(who_added_id);

-- Таблица status_types
CREATE TABLE IF NOT EXISTS status_types (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(255) NOT NULL UNIQUE
);

-- Таблица status_periods
CREATE TABLE IF NOT EXISTS status_periods (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    employee_id BIGINT NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    status_id BIGINT NOT NULL REFERENCES status_types(id) ON DELETE CASCADE,
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    comment TEXT,
    who_added_id BIGINT NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    one_time_event BOOLEAN NOT NULL DEFAULT false
);

-- Индексы для status_periods
CREATE INDEX IF NOT EXISTS idx_status_periods_employee_id ON status_periods(employee_id);
CREATE INDEX IF NOT EXISTS idx_status_periods_status_id ON status_periods(status_id);
CREATE INDEX IF NOT EXISTS idx_status_periods_who_added_id ON status_periods(who_added_id);

-- Теперь добавляем внешний ключ who_added_id → employees(id)
-- (отложенное добавление, чтобы избежать циклической зависимости при вставке первого пользователя)
ALTER TABLE employees
ADD CONSTRAINT fk_employees_who_added
FOREIGN KEY (who_added_id) REFERENCES employees(id) ON DELETE SET NULL;

-- Заполняем данными

-- Статусы
INSERT INTO status_types (name, code, created_at, updated_at)
VALUES
    ('В офисе', 'work_office', NOW(), NOW()),
    ('Удаленная работа', 'work_remote', NOW(), NOW()),
    ('Командировка', 'business_trip', NOW(), NOW()),
    ('Отпуск', 'vacation', NOW(), NOW()),
    ('Больничный', 'sick_leave', NOW(), NOW()),
    ('Выходной', 'weekend', NOW(), NOW()),
    ('Отгул', 'day_off', NOW(), NOW()),
    ('Работа в выходной день', 'weekend_work', NOW(), NOW())
ON CONFLICT (code) DO NOTHING;

-- Отдел УКП
INSERT INTO departments (name, created_at, updated_at)
VALUES ('УКП', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- Пользователь Анастасия Пучкова (без who_added_id сначала)
INSERT INTO employees (
    first_name, last_name, email, password_hash, position,
    department_id, is_active, is_admin,
    created_at, updated_at
)
VALUES (
    'Анастасия', 'Пучкова', 'a@a.ru',
    '$2a$10$ibKRQ.daiCT1q8SQylmlGuTXyOhKa.189e1dQ5FjAtd6c8cPVXsTO',
    'Ведущий инженер',
    (SELECT id FROM departments WHERE name = 'УКП' LIMIT 1),
    true, true,
    '2025-08-21 13:59:42.280 +0300'::timestamptz,
    '2025-09-19 16:23:10.134 +0300'::timestamptz
)
ON CONFLICT (email) DO NOTHING;

-- Обновляем who_added_id — указываем на самого себя
UPDATE employees
SET who_added_id = id
WHERE email = 'a@a.ru' AND who_added_id IS NULL;

-- Готово!
SELECT '✅ База sitex_db успешно инициализирована!' AS result;
