# GEMINI.md — WrenPanel

Context file cho AI agent làm việc trên repo này. Đọc trước khi generate hoặc sửa code.

## Mục tiêu dự án

WrenPanel: control panel quản lý VPS/hosting tự build, tương tự cPanel/DirectAdmin/aaPanel nhưng gọn nhẹ hơn. Ưu tiên: ổn định, dễ cấu hình, ít dependency, bảo mật ở tầng root operation.

## Nguyên tắc kiến trúc — KHÔNG được vi phạm

1. **Một binary Go duy nhất.** Frontend Vue 3 build tĩnh, nhúng bằng `embed.FS`. Không tách service ngoài trừ process bị quản lý (PHP-FPM, MySQL, Node app...).
2. **Tách privilege bắt buộc.** HTTP API chạy user thường, không có quyền root. Mọi thao tác đụng OS (tạo user, chown, iptables/nftables, systemctl, mount) đi qua "root worker" — process riêng giao tiếp qua Unix socket/gRPC, chỉ nhận lệnh nằm trong allowlist. Không dùng `os/exec` với string nối trực tiếp từ input người dùng.
3. **Không tự chạy process nền.** Mọi app (PHP, Python, Node, Docker) chạy qua `systemd` unit do panel sinh ra, để systemd tự restart khi crash.
4. **SSL qua `lego`** (ACME client Go), không shell-out certbot.
5. **Không đụng package manager hệ thống** cho PHP/Node/Python — mỗi version cài side-by-side, không ghi đè symlink mặc định OS (tránh gãy tool hệ thống phụ thuộc `python3`/`php` mặc định).

## Cấu trúc thư mục hosting — chuẩn cPanel

```
/home/<username>/
├── public_html/          # webroot domain chính, addon domain nằm subfolder ở đây
├── mail/<domain>/<user>/ # Maildir, UID/GID riêng, KHÔNG chung UID với web owner
├── etc/<domain>/         # config mail domain phụ
├── logs/                 # access/error log riêng theo domain
├── ssl/{certs,keys}/
├── tmp/                  # session PHP riêng theo user, KHÔNG dùng /tmp hệ thống
└── .cpanel/
```

## Cấu trúc thư mục cài đặt WrenPanel trên server (khác với /home/<username>)

Đây là nơi panel lưu chính nó — binary, config, DB, log, runtime version — tách bạch hoàn toàn khỏi thư mục hosting của account. Theo chuẩn FHS Linux, không tự chế path tuỳ tiện:

```
/opt/wrenpanel/
├── bin/wrenpanel                 # binary chính, chạy user thường (systemd: wrenpanel.service)
├── bin/wrenpanel-worker          # binary root worker (systemd: wrenpanel-worker.service)
├── node/<version>/               # runtime Node cài riêng, KHÔNG phải account nào sở hữu
├── python/<version>/             # runtime Python cài riêng
└── VERSION

/etc/wrenpanel/
├── config.yaml                   # config panel: port API, đường dẫn DB, log level
├── nginx/
│   ├── snippets/                 # template chung: gzip, security header...
│   └── vhosts/<fqdn>.conf        # 1 file/site, panel generate — nguồn xoá/sửa thật sự
├── apache/vhosts/<fqdn>.conf
└── ssl/<fqdn>/{cert.pem,privkey.pem}   # bản panel giữ, symlink hoặc copy vào path Nginx đọc

/var/lib/wrenpanel/
├── wrenpanel.db                  # SQLite — toàn bộ schema đã định nghĩa ở trên
├── db.wal / db.shm               # WAL mode, bật để tránh khoá ghi khi nhiều request
└── backups/staging/              # nơi build file tar.zst TRƯỚC khi đẩy lên đích (S3/FTP/rsync)

/var/log/wrenpanel/
├── panel.log                     # log HTTP API
├── worker.log                    # log root worker — MỌI lệnh allowlist đã chạy
└── audit.log                     # ghi trùng với bảng audit_log trong DB, dạng append-only file để đối chiếu khi DB có sự cố

/run/wrenpanel/
└── worker.sock                   # Unix socket, API layer nói chuyện với root worker qua đây

/etc/systemd/system/
├── wrenpanel.service
├── wrenpanel-worker.service
└── wrenpanel-app-<name>-<hash>.service   # panel tự sinh, 1 file/app Node/Python/Docker
```

**Quy tắc bắt buộc:**
- `/etc/wrenpanel/nginx/vhosts/*.conf` là **nguồn sự thật duy nhất** cho vhost — panel không sửa trực tiếp `/etc/nginx/sites-enabled/`, chỉ symlink từ đó trỏ vào file panel generate. Xoá vhost = xoá file này + reload Nginx, không đi tìm sửa config hệ thống nằm rải rác.
- `wrenpanel.db` chỉ 1 process API được ghi (root worker không đụng DB trực tiếp, chỉ nhận lệnh qua socket rồi API tự ghi kết quả vào DB) — tránh race condition ghi đồng thời.
- `worker.sock` set quyền `0660`, chỉ user chạy `wrenpanel.service` được connect — không public socket ra ngoài group khác.
- Toàn bộ thư mục trên **không nằm trong `/home/`** — tách biệt hoàn toàn dữ liệu quản trị panel khỏi dữ liệu account, để backup full server và backup panel tự thân là 2 việc độc lập.

## Web server — Nginx đứng trước, Apache lùi nội bộ

- Nginx giữ cổng 80/443 duy nhất, SSL termination, đọc SNI + Host header route site.
- Apache (nếu site cần `.htaccess`/mod_php) chỉ bind `127.0.0.1:<port nội bộ>`, không bao giờ `0.0.0.0`. Nginx `proxy_pass` sang Apache cho PHP request, tự trả static file trực tiếp.
- Site không cần Apache: Nginx nói thẳng PHP-FPM qua unix socket, chọn "Web engine" theo từng site.
- Add domain: sinh vhost → reload Nginx (không restart) → verify DNS A record đúng IP trước khi xin SSL qua HTTP-01.

## PHP Manager

- Repo side-by-side (`ondrej/php` PPA / Remi repo), mỗi version = gói `phpX.Y-fpm` độc lập, systemd unit riêng, pool riêng theo site (`pool.d/site1.conf`, socket riêng).
- Extension load ở cấp PHP-FPM master, dùng chung mọi pool cùng version — không thể khác extension set nếu cùng version, muốn khác thật sự phải đổi version.
- Per-site override qua `php_admin_value` (khoá cứng, không override runtime được — dùng cho `open_basedir`, `disable_functions`) và `php_value` (site tự `ini_set()` được — dùng cho giá trị không nhạy cảm).
- `open_basedir` bắt buộc set theo từng site, cô lập dù chung PHP-FPM master.
- Gỡ version: chặn nếu còn site đang gán version đó, chỉ cho gỡ khi 0 tham chiếu.

## Node.js Manager

- Không dùng gói `nodejs` từ apt. Tải binary chính thức vào `/opt/wrenpanel/node/<version>/`, không cài vào `/usr/bin`.
- Mỗi app 1 systemd unit, `ExecStart` trỏ thẳng binary đúng version. Không dùng PM2.
- Gỡ version: quét toàn bộ unit file panel quản lý, chặn nếu còn `ExecStart` tham chiếu.

## Python Manager

- `deadsnakes` PPA hoặc build từ source vào `/opt/wrenpanel/python/<version>/`, không đụng `python3` mặc định hệ thống (OS tool phụ thuộc đúng version này).
- Mỗi app 1 venv riêng, tạo bằng đúng interpreter cần — dependency cô lập hoàn toàn giữa các app.
- Gỡ version: cảnh báo nếu còn venv tham chiếu, không tự động gỡ.

## Docker Manager

- Docker SDK Go chính thức qua `/var/run/docker.sock`, không parse CLI output.
- Mỗi app 1 `compose.yml` do panel sinh từ form.

## Quy tắc chung tránh xung đột port/tên

- Mỗi systemd unit panel tạo: tên duy nhất `<loại>-<tên-app>-<hash-ngắn>.service`.
- Bảng `port_allocations` trong SQLite cấp port nội bộ tự động, thử bind trước khi ghi vào unit — không đoán port cố định.
- Trước khi apt install, check `apt list --installed` tránh cài đè setup thủ công có sẵn của user.

## Domain / Subdomain / Service prefix — phân biệt dứt điểm

| Loại | Ví dụ | Bản chất |
|---|---|---|
| Domain chính | domain.com | 1 website, docroot `public_html/` |
| Subdomain | shop.domain.com | Website khác, docroot mặc định `public_html/shop/` |
| Addon domain | otherbrand.com | Domain khác trỏ vào subfolder site chính |
| Service prefix | mail.domain.com | KHÔNG phải website — vhost hệ thống, panel tự sinh |

- **Reserved prefix list** (không cho user tạo subdomain trùng): `mail`, `webmail`, `wrenpanel`, `autoconfig`, `autodiscover`, `ftp`, `ns1`, `ns2`.
- **MX record** độc lập với vhost `mail.domain.com` — MX dùng cho SMTP nhận thư (Postfix/Dovecot), không qua HTTP/Nginx. Vhost `mail.` chỉ phục vụ webmail UI (Roundcube/SnappyMail). Hai hệ thống tách biệt, không gộp logic.
- Domain mới → sinh sẵn DNS template (A, MX, TXT SPF) cho user, hoặc tự ghi zone file nếu panel tự chạy BIND/PowerDNS.

## Vhost — model docroot_owner (mấu chốt xoá an toàn)

- Mỗi vhost record lưu `docroot_owner` = id vhost đã tạo folder đó lần đầu.
- **Xoá vhost:** query xem còn vhost active nào khác trỏ chung `docroot` không.
  - Có → chỉ xoá config, không đụng file.
  - Không, và vhost đang xoá chính là `docroot_owner` → hỏi xác nhận xoá kèm file hay giữ lại.
  - Không, nhưng không phải `docroot_owner` (share folder từ domain gốc) → chỉ xoá config, giữ nguyên file tuyệt đối.
- **Add subdomain** UI luôn hỏi rõ 2 lựa chọn: tạo thư mục mới (mặc định) hoặc dùng chung thư mục có sẵn (dropdown chọn path tồn tại) — set `docroot_owner` tương ứng ngay lúc tạo, không đoán ngầm.
- **Xoá domain chính còn subdomain/addon con:** chặn cứng, bắt xoá hết con trước, không tự cascade (nguy hiểm nếu con đang share folder với domain khác).

## Database services

Độc lập, không gắn theo domain — 1 instance MySQL/PostgreSQL/MongoDB dùng chung cả server. Panel chỉ tạo user + database + gán quyền theo site. phpMyAdmin/Adminer đặt sau path/subdomain nội bộ có xác thực panel, không public mặc định.

## Backup/Restore

3 cấp: site / DB riêng lẻ / full server. Format `tar.zst` + `manifest.json` (version panel, PHP/Node version, danh sách DB). Restore giải nén vào thư mục tạm, verify manifest xong mới swap. Đích: local, S3-compatible, FTP/SFTP, rsync.

## SQLite Schema

```sql
-- ==== Tài khoản hosting (1 user Linux = 1 account) ====
CREATE TABLE accounts (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    username        TEXT UNIQUE NOT NULL,       -- trùng tên user Linux
    home_dir        TEXT NOT NULL,               -- /home/<username>
    email           TEXT,
    disk_quota_mb   INTEGER DEFAULT 0,           -- 0 = unlimited
    status          TEXT DEFAULT 'active',       -- active | suspended
    created_at      TEXT DEFAULT CURRENT_TIMESTAMP
);

-- ==== Vhost — domain, subdomain, addon, service prefix ====
CREATE TABLE vhosts (
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
CREATE INDEX idx_vhosts_docroot ON vhosts(docroot);
CREATE INDEX idx_vhosts_account ON vhosts(account_id);

-- ==== PHP version cài trên server ====
CREATE TABLE php_versions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    version         TEXT UNIQUE NOT NULL,         -- "8.2"
    fpm_service     TEXT NOT NULL,                -- "php8.2-fpm"
    binary_path     TEXT NOT NULL,
    installed_at    TEXT DEFAULT CURRENT_TIMESTAMP
);

-- ==== Node.js version cài trên server ====
CREATE TABLE node_versions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    version         TEXT UNIQUE NOT NULL,         -- "20.11.0"
    install_path    TEXT NOT NULL,                -- /opt/wrenpanel/node/20.11.0
    installed_at    TEXT DEFAULT CURRENT_TIMESTAMP
);

-- ==== Python version cài trên server ====
CREATE TABLE python_versions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    version         TEXT UNIQUE NOT NULL,         -- "3.11"
    interpreter_path TEXT NOT NULL,
    installed_at    TEXT DEFAULT CURRENT_TIMESTAMP
);

-- ==== App chạy nền: Node/Python/Docker, mỗi app 1 systemd unit ====
CREATE TABLE apps (
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
CREATE TABLE port_allocations (
    port            INTEGER PRIMARY KEY,
    app_id          INTEGER REFERENCES apps(id) ON DELETE CASCADE,
    reserved_at     TEXT DEFAULT CURRENT_TIMESTAMP
);

-- ==== SSL certificate ====
CREATE TABLE certs (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    fqdn            TEXT NOT NULL,
    cert_path       TEXT NOT NULL,
    key_path        TEXT NOT NULL,
    issuer          TEXT DEFAULT 'lets-encrypt',   -- lets-encrypt | uploaded
    issued_at       TEXT,
    expires_at      TEXT,
    auto_renew      INTEGER DEFAULT 1
);

-- ==== Database (MySQL/Postgres/Mongo) cấp cho từng site ====
CREATE TABLE databases (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id      INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    vhost_id        INTEGER REFERENCES vhosts(id) ON DELETE SET NULL,
    engine          TEXT NOT NULL,                 -- mysql | postgresql | mongodb
    db_name         TEXT NOT NULL,
    db_user         TEXT NOT NULL,
    created_at      TEXT DEFAULT CURRENT_TIMESTAMP
);

-- ==== Backup ====
CREATE TABLE backups (
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
CREATE TABLE cron_jobs (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id      INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    schedule        TEXT NOT NULL,                 -- cron expression
    command         TEXT NOT NULL,
    status          TEXT DEFAULT 'active'
);

-- ==== Mail account ====
CREATE TABLE mail_accounts (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    vhost_id        INTEGER NOT NULL REFERENCES vhosts(id) ON DELETE CASCADE,
    address         TEXT UNIQUE NOT NULL,
    maildir_path    TEXT NOT NULL,
    quota_mb        INTEGER DEFAULT 0,
    status          TEXT DEFAULT 'active'
);

-- ==== Audit log — mọi thao tác root worker ====
CREATE TABLE audit_log (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    actor           TEXT NOT NULL,                 -- admin username
    action          TEXT NOT NULL,                 -- vhost.create, php.remove_version...
    target          TEXT,
    payload_json    TEXT,
    result          TEXT,                          -- success | failed
    created_at      TEXT DEFAULT CURRENT_TIMESTAMP
);
```

## Cấu trúc repo — Backend (Go)

```
wrenpanel/
├── cmd/
│   ├── wrenpanel/          # main binary — HTTP API, không có quyền root
│   └── wrenpanel-worker/   # root worker — nhận lệnh qua gRPC, allowlist
├── internal/
│   ├── api/                # HTTP handlers, theo domain: vhost, php, node, ssl, backup...
│   ├── worker/              # implement các lệnh allowlist, gọi thẳng syscall/exec
│   ├── store/                # SQLite queries, mỗi bảng 1 file
│   ├── systemdgen/           # sinh unit file
│   ├── vhostgen/              # sinh vhost Nginx/Apache
│   └── acme/                   # wrap lego
├── web/                        # Vue 3 source, build ra dist/ rồi embed
├── migrations/                  # SQLite schema versioned
└── GEMINI.md
```

## Cấu trúc repo — Frontend (Vue 3)

```
web/src/
├── views/
│   ├── Vhosts/          # list, create, edit, add subdomain
│   ├── PhpManager/
│   ├── NodeManager/
│   ├── PythonManager/
│   ├── Docker/
│   ├── Databases/
│   ├── SSL/
│   └── Backups/
├── components/          # form, table, modal xác nhận xoá dùng chung
├── api/                  # 1 file gọi API tương ứng mỗi domain backend
└── stores/                # Pinia, theo domain
```

## Giao diện tham khảo

- **Layout tổng thể:** tham khảo aaPanel/1Panel — sidebar trái cố định theo nhóm chức năng (Website, App Manager, Database, SSL, Backup, Firewall), khu vực chính dạng bảng + action inline, không dùng nhiều màn hình wizard nhiều bước cho tác vụ đơn giản.
- **Table-first, không card-heavy:** danh sách vhost/database/app hiển thị dạng bảng có cột trạng thái, hành động (edit/delete/restart) nằm cuối hàng — giống DirectAdmin, tra cứu nhanh hơn dạng card lớn.
- **Modal xác nhận xoá bắt buộc có chi tiết ảnh hưởng:** không chỉ "Bạn có chắc?" — liệt kê rõ những gì bị xoá theo (config/file/DB/cert), đúng theo cây quyết định `docroot_owner` đã thiết kế ở trên.
- **Không dùng hiệu ứng trang trí, gradient, animation thừa** — tuân theo quy ước code/UI đã đặt ra: tối giản, tập trung vào dữ liệu và thao tác.
