-- Seed initial data for Turnate

-- Create default admin user (password: admin123)
INSERT OR IGNORE INTO users (
    id, username, email, password, display_name, role, is_active, created_at, updated_at
) VALUES (
    '01234567-89ab-7def-8901-234567890123',
    'admin',
    'admin@turnate.com',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- admin123
    'Administrator',
    'admin',
    true,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Create default general channel
INSERT OR IGNORE INTO channels (
    id, name, description, type, created_by, created_at, updated_at
) VALUES (
    '01234567-89ab-7def-8901-234567890124',
    'general',
    'General discussion channel',
    'public',
    '01234567-89ab-7def-8901-234567890123',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Add admin to general channel
INSERT OR IGNORE INTO channel_members (
    channel_id, user_id, created_at, updated_at
) VALUES (
    '01234567-89ab-7def-8901-234567890124',
    '01234567-89ab-7def-8901-234567890123',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Welcome message
INSERT OR IGNORE INTO messages (
    id, content, user_id, channel_id, created_at, updated_at
) VALUES (
    '01234567-89ab-7def-8901-234567890125',
    'Welcome to Turnate! 🎉 This is your general discussion channel.',
    '01234567-89ab-7def-8901-234567890123',
    '01234567-89ab-7def-8901-234567890124',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);