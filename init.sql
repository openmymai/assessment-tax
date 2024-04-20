CREATE TYPE allowance_type AS ENUM ('Personal', 'Kreceipt');

CREATE TABLE IF NOT EXISTS allowances (
  id SERIAL PRIMARY KEY,
  allowance_type VARCHAR(255) NOT NULL,
  amount DECIMAL NOT NULL
);

INSERT INTO allowances (allowance_type, amount) VALUES
('Personal', 60000.0),
('Kreceipt', 0.0);