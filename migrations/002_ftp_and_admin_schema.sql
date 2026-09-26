-- Migration 002: FTP Manager, Admins, and phpMyAdmin / Adminer support

-- ==== Admin accounts ====
CREATE TABLE IF NOT EXISTS admins (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    username        TEXT UNIQUE NOT NULL,
    password_hash   TEXT NOT NULL,
    role            TEXT DEFAULT 'superadmin',
    totp_secret     TEXT,
    totp_enabled    INTEGER DEFAULT 0,
    status          TEXT DEFAULT 'active',
    last_login_at   TEXT,
    created_at      TEXT DEFAULT CURRENT_TIMESTAMP
);

-- ==== FTP Accounts ====
CREATE TABLE IF NOT EXISTS ftp_accounts (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id  INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    username    TEXT UNIQUE NOT NULL,
    root_dir    TEXT NOT NULL,
    status      TEXT DEFAULT 'active', -- active | disabled
    created_at  TEXT DEFAULT CURRENT_TIMESTAMP
);

-- ==== phpMyAdmin / Database Web Manager ====
CREATE TABLE IF NOT EXISTS db_managers (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    engine      TEXT UNIQUE NOT NULL, -- phpmyadmin | adminer
    internal_port INTEGER DEFAULT 8089,
    secret_path TEXT NOT NULL,
    status      TEXT DEFAULT 'stopped' -- running | stopped
);
