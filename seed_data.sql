-- Pastikan kamu sudah berada di database yang benar, misal: \c golang_crud

-- Insert dummy users
INSERT INTO users (name, email)
VALUES
  ('Alice Johnson', 'alice@example.com'),
  ('Bob Smith', 'bob@example.com'),
  ('Charlie Brown', 'charlie@example.com'),
  ('Diana Prince', 'diana@example.com'),
  ('Ethan Hunt', 'ethan@example.com');
