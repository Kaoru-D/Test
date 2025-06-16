CREATE DATABASE finanzas;


CREATE table if not exists accounts (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255),
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE if not exists transactions (
  id SERIAL PRIMARY KEY,
  account_id INTEGER REFERENCES accounts(id),
  amount NUMERIC(12,2),
  currency VARCHAR(3),
  type VARCHAR(20),
  description TEXT,
  created_at TIMESTAMP DEFAULT NOW()
);

