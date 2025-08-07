-- Rollback initial schema

DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_channels_type;
DROP INDEX IF EXISTS idx_messages_thread_created;
DROP INDEX IF EXISTS idx_messages_channel_created;
DROP INDEX IF EXISTS idx_channel_members_unique;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;

DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS channel_members;
DROP TABLE IF EXISTS channels;
DROP TABLE IF EXISTS users;