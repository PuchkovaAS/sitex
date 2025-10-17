-- Подключаемся к postgres
\c postgres

-- Создаём базу ТОЛЬКО если её нет
SELECT 'CREATE DATABASE sitex_db'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'sitex_db')
\gexec

-- Переключаемся на новую базу
\c sitex_db

-- === Создание таблиц БЕЗ внешнего ключа who_added_id в employees ===

CREATE TABLE IF NOT EXISTS departments (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_department_name ON departments(name);

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
    show_time_events BOOLEAN DEFAULT false,  -- ← ДОБАВЛЕНО
    who_added_id BIGINT -- временно без внешнего ключа
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_employees_email ON employees(email);
CREATE INDEX IF NOT EXISTS idx_employees_active ON employees(is_active);
CREATE INDEX IF NOT EXISTS idx_employees_admin ON employees(is_admin);
CREATE INDEX IF NOT EXISTS idx_employees_department_id ON employees(department_id);
CREATE INDEX IF NOT EXISTS idx_employees_who_added_id ON employees(who_added_id);



CREATE TABLE IF NOT EXISTS status_types (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(255) NOT NULL UNIQUE
);

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

-- Индексы
CREATE INDEX IF NOT EXISTS idx_status_periods_employee_id ON status_periods(employee_id);
CREATE INDEX IF NOT EXISTS idx_status_periods_status_id ON status_periods(status_id);
CREATE INDEX IF NOT EXISTS idx_status_periods_who_added_id ON status_periods(who_added_id);



-- === Таблицы для учёта времени (опоздания, ранние уходы) ===

CREATE TABLE IF NOT EXISTS time_event_types (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    code VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_time_event_types_deleted_at ON time_event_types (deleted_at);

CREATE TABLE IF NOT EXISTS time_events (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    employee_id BIGINT NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    who_added_id BIGINT NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    event_type_id BIGINT NOT NULL REFERENCES time_event_types(id) ON DELETE RESTRICT,
    date DATE NOT NULL,
    scheduled_time TIME NOT NULL,
    actual_time TIME NOT NULL,
    description TEXT,
    difference_min INTEGER NOT NULL
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_time_events_employee_id ON time_events (employee_id);
CREATE INDEX IF NOT EXISTS idx_time_events_who_added_id ON time_events (who_added_id);
CREATE INDEX IF NOT EXISTS idx_time_events_event_type_id ON time_events (event_type_id);
CREATE INDEX IF NOT EXISTS idx_time_events_date ON time_events (date);
CREATE INDEX IF NOT EXISTS idx_time_events_deleted_at ON time_events (deleted_at);

-- === Заполнение данными ===

-- Статусы
INSERT INTO status_types (name, code)
VALUES
    ('В офисе', 'work_office'),
    ('Удаленная работа', 'work_remote'),
    ('Командировка', 'business_trip'),
    ('Отпуск', 'vacation'),
    ('Больничный', 'sick_leave'),
    ('Выходной', 'weekend'),
    ('Отгул', 'day_off'),
    ('Работа в выходной день', 'weekend_work'),
    ('Отпуск за свой счёт', 'unpaid_vacation')
ON CONFLICT (code) DO NOTHING;

-- Типы временных событий
INSERT INTO time_event_types (code, name)
VALUES
    ('late', 'Опоздание'),
    ('early_leave', 'Ранний уход')
ON CONFLICT (code) DO NOTHING;

-- Отдел
INSERT INTO departments (name)
VALUES ('УКП')
ON CONFLICT (name) DO NOTHING;

-- Пользователь (who_added_id = NULL на этом этапе — безопасно)
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

-- === ТОЛЬКО СЕЙЧАС добавляем внешний ключ — все данные уже на месте ===
ALTER TABLE employees
ADD CONSTRAINT fk_employees_who_added
FOREIGN KEY (who_added_id) REFERENCES employees(id) ON DELETE SET NULL;

-- Готово!
SELECT '✅ База sitex_db успешно инициализирована!' AS result;
