-- Users
INSERT INTO users (id, name, email, created_at, updated_at) VALUES
    ('11111111-1111-1111-1111-111111111111', 'John Doe', 'john@example.com', NOW(), NOW()),
    ('22222222-2222-2222-2222-222222222222', 'Jane Smith', 'jane@example.com', NOW(), NOW());

-- Repositories
INSERT INTO repositories (id, user_id, name, url, ai_enabled, created_at, updated_at) VALUES
    ('aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111', 'Go Project', 'https://github.com/john/go', true, NOW(), NOW()),
    ('bbbbbbb2-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '22222222-2222-2222-2222-222222222222', 'React App', 'https://github.com/jane/react', false, NOW(), NOW());
