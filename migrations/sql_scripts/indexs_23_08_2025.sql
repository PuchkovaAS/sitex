-- Индекс для быстрого поиска по employee_id и дате
CREATE INDEX IF NOT EXISTS idx_status_periods_employee_date 
ON status_periods(employee_id, start_date);

-- Индекс для статусов
CREATE INDEX IF NOT EXISTS idx_status_periods_status 
ON status_periods(status_id);

-- Индекс для email сотрудников
CREATE INDEX IF NOT EXISTS idx_employees_email 
ON employees(email);
