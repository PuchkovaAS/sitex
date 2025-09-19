-- Добавляем UNIQUE constraints
ALTER TABLE employees ADD CONSTRAINT uni_employees_email UNIQUE (email);
ALTER TABLE status_types ADD CONSTRAINT uni_status_types_name UNIQUE (name);
ALTER TABLE status_types ADD CONSTRAINT uni_status_types_code UNIQUE (code);
ALTER TABLE official_holidays ADD CONSTRAINT uni_official_holidays_date UNIQUE (date);

-- Добавляем FOREIGN KEY constraints
ALTER TABLE status_periods 
    ADD CONSTRAINT fk_status_periods_employee 
    FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE;

ALTER TABLE status_periods 
    ADD CONSTRAINT fk_status_periods_status 
    FOREIGN KEY (status_id) REFERENCES status_types(id) ON DELETE CASCADE;
