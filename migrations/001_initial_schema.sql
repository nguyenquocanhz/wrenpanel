-- WrenPanel initial SQLite schema
-- Defined in .agent/GEMINI.md

PRAGMA foreign_keys = ON;

-- ==== Tài khoản hosting (1 user Linux = 1 account) ====
CREATE TABLE IF NOT EXISTS accounts (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    username        TEXT UNIQUE NOT NULL,       -- trùng tên user Linux
    home_dir        TEXT NOT NULL,               -- /home/<username>
    email           TEXT,
    disk_quota_mb   INTEGER DEFAULT 0,           -- 0 = unlimited
    status          TEXT DEFAULT 'active',       -- active | suspended
    created_at      TEXT DEFAULT CURRENT_TIMESTAMP
);

-- ==== PHP version cài trên server ====
CREATE TABLE IF NOT EXISTS php_versions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    version         TEXT UNIQUE NOT NULL,         -- "8.2"
    fpm_service     TEXT NOT NULL,                -- "php8.2-fpm"
    binary_path     TEXT NOT NULL,
    installed_at    TEXT DEFAULT CURRENT_TIMESTAMP
);

-- ==== Node.js version cài trên server ====
CREATE TABLE IF NOT EXISTS node_versions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    version         TEXT UNIQUE NOT NULL,         -- "20.11.0"
    install_path    TEXT NOT NULL,                -- /opt/wrenpanel/node/20.11.0
    installed_at    TEXT DEFAULT CURRENT_TIMESTAMP
);

-- ==== Python version cài trên server ====
CREATE TABLE IF NOT EXISTS python_versions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    version         TEXT UNIQUE NOT NULL,         -- "3.11"
    interpreter_path TEXT NOT NULL,
    installed_at    TEXT DEFAULT CURRENT_TIMESTAMP
);

-- ==== SSL certificate ====
CREATE TABLE IF NOT EXISTS certs (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    fqdn            TEXT NOT NULL,
    cert_path       TEXT NOT NULL,
    key_path        TEXT NOT NULL,
    issuer          TEXT DEFAULT 'lets-encrypt',   -- lets-encrypt | uploaded
    issued_at       TEXT,
    expires_at      TEXT,
    auto_renew      INTEGER DEFAULT 1
);

-- ==== Vhost — domain, subdomain, addon, service prefix ====
CREATE TABLE IF NOT EXISTS vhosts (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id      INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    parent_vhost_id INTEGER REFERENCES vhosts(id) ON DELETE RESTRICT,
    type            TEXT NOT NULL,               -- primary | subdomain | addon | alias | system
    fqdn            TEXT UNIQUE NOT NULL,
    docroot         TEXT NOT NULL,
    docroot_owner_id INTEGER REFERENCES vhosts(id), -- vhost đã tạo folder này lần đầu
    web_engine      TEXT DEFAULT 'nginx',         -- nginx | nginx+apache
    php_version_id  INTEGER REFERENCES php_versions(id),
    ssl_cert_id     INTEGER REFERENCES certs(id),
    status          TEXT DEFAULT 'active',        -- active | disabled
    created_at      TEXT DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_vhosts_docroot ON vhosts(docroot);
CREATE INDEX IF NOT EXISTS idx_vhosts_account ON vhosts(account_id);

-- ==== App chạy nền: Node/Python/Docker, mỗi app 1 systemd unit ====
CREATE TABLE IF NOT EXISTS apps (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id      INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    vhost_id        INTEGER REFERENCES vhosts(id) ON DELETE SET NULL,
    kind            TEXT NOT NULL,                -- node | python | docker
    name            TEXT NOT NULL,
    systemd_unit    TEXT UNIQUE,                  -- tên unit panel sinh
    runtime_version TEXT,                          -- version node/python dùng
    internal_port   INTEGER,
    entrypoint      TEXT,
    status          TEXT DEFAULT 'stopped',
    created_at      TEXT DEFAULT CURRENT_TIMESTAMP
);

-- ==== Cấp phát port nội bộ, tránh trùng ====
CREATE TABLE IF NOT EXISTS port_allocations (
    port            INTEGER PRIMARY KEY,
    app_id          INTEGER REFERENCES apps(id) ON DELETE CASCADE,
    reserved_at     TEXT DEFAULT CURRENT_TIMESTAMP
);

-- ==== Database (MySQL/Postgres/Mongo) cấp cho từng site ====
CREATE TABLE IF NOT EXISTS databases (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id      INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    vhost_id        INTEGER REFERENCES vhosts(id) ON DELETE SET NULL,
    engine          TEXT NOT NULL,                 -- mysql | postgresql | mongodb
    db_name         TEXT NOT NULL,
    db_user         TEXT NOT NULL,
    created_at      TEXT DEFAULT CURRENT_TIMESTAMP
);

-- ==== Backup ====
CREATE TABLE IF NOT EXISTS backups (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id      INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    scope           TEXT NOT NULL,                 -- site | database | full
    vhost_id        INTEGER REFERENCES vhosts(id) ON DELETE SET NULL,
    file_path       TEXT NOT NULL,
    destination     TEXT NOT NULL,                 -- local | s3 | ftp | rsync
    manifest_json   TEXT,
    created_at      TEXT DEFAULT CURRENT_TIMESTAMP
);

-- ==== Cron job của account ====
CREATE TABLE IF NOT EXISTS cron_jobs (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id      INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    schedule        TEXT NOT NULL,                 -- cron expression
    command         TEXT NOT NULL,
    status          TEXT DEFAULT 'active'
);

-- ==== Mail account ====
CREATE TABLE IF NOT EXISTS mail_accounts (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    vhost_id        INTEGER NOT NULL REFERENCES vhosts(id) ON DELETE CASCADE,
    address         TEXT UNIQUE NOT NULL,
    maildir_path    TEXT NOT NULL,
    quota_mb        INTEGER DEFAULT 0,
    status          TEXT DEFAULT 'active'
);

-- ==== Audit log — mọi thao tác root worker ====
CREATE TABLE IF NOT EXISTS audit_log (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    actor           TEXT NOT NULL,                 -- admin username
    action          TEXT NOT NULL,                 -- vhost.create, php.remove_version...
    target          TEXT,
    payload_json    TEXT,
    result          TEXT,                          -- success | failed
    created_at      TEXT DEFAULT CURRENT_TIMESTAMP
);
