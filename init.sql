CREATE TYPE allowance_type AS ENUM ('Private', 'Donation', 'Kreceipt');

CREATE TABLE IF NOT EXISTS allowances (
  allowance_type VARCHAR(255) NOT NULL,
  amount DECIMAL NOT NULL
);

INSERT INTO allowances (allowance_type, amount) VALUES
('Private', 60000.00),
('Kreceipt', 0.0);