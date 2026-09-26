# WrenPanel

<p align="center">
  <strong>Control panel quản lý VPS & Hosting nhẹ, bảo mật cấp kernel Linux, tương tự cPanel/DirectAdmin nhưng tối giản và độc lập.</strong>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go" alt="Go Version" />
  <img src="https://img.shields.io/badge/Vue-3.4+-4FC08D?style=flat&logo=vue.js" alt="Vue 3" />
  <img src="https://img.shields.io/badge/TailwindCSS-3.4+-38B2AC?style=flat&logo=tailwind-css" alt="TailwindCSS" />
  <img src="https://img.shields.io/badge/Target-Linux%20Ubuntu%20%2F%20Debian-E95420?style=flat&logo=ubuntu" alt="Ubuntu" />
  <img src="https://img.shields.io/badge/Architecture-FHS%20Standard-blue" alt="FHS Standard" />
</p>

---

## 🌟 Các đặc điểm kiến trúc cốt lõi (Tuân thủ [.agent/GEMINI.md](.agent/GEMINI.md))

1. **Một binary Go duy nhất (`wrenpanel`):**
   - Frontend Vue 3 build tĩnh, nhúng trực tiếp qua `embed.FS`.
   - API HTTP chạy với quyền user thường (`wrenpanel.service`), **tuyệt đối không chạy root**.
2. **Tách biệt đặc quyền bắt buộc (`wrenpanel-worker`):**
   - Mọi thao tác đụng OS (tạo user, chown, systemctl, reload nginx) chạy qua **root worker**.
   - Giao tiếp qua Unix domain socket nội bộ `/run/wrenpanel/worker.sock` (set cứng quyền `0660`).
   - Chỉ nhận lệnh trong danh sách **Allowlist**, cấm tuyệt đối nối chuỗi shell từ input người dùng.
3. **Mô hình Vhost & `docroot_owner` (Xóa an toàn):**
   - Ghi nhận vhost nào đã tạo thư mục webroot lần đầu.
   - Khi xóa vhost: nếu có website khác dùng chung thì giữ nguyên file; nếu là chủ sở hữu và không ai dùng chung mới cho phép tùy chọn xóa; chặn xóa domain chính nếu còn subdomain con.
4. **Quản lý Ứng dụng qua Systemd độc lập:**
   - Mỗi app Node.js, Python, Docker chạy qua 1 systemd unit riêng do panel sinh ra (`wrenpanel-app-<name>-<hash>.service`). Systemd tự động auto-restart khi crash.
   - Bảng `port_allocations` cấp port nội bộ tự động và probe socket trước khi ghi vào unit file.
5. **Cấp SSL tự động qua Lego (Go ACME):**
   - ACME client Go tích hợp sẵn, xác thực HTTP-01 trực tiếp, không cần certbot.
   - Tự động kiểm tra bản ghi DNS A trước khi gửi yêu cầu chứng chỉ.
6. **phpMyAdmin / Web Database Manager an toàn:**
   - Đặt sau proxy nội bộ và cơ chế Single Sign-On (SSO) có xác thực của panel, không public lộ đường dẫn ra ngoài internet.
7. **FTP Manager với Chroot Jail:**
   - Cấp tài khoản FTP ảo bị giới hạn chặt chẽ trong thư mục website, không cho phép leo thang thư mục hệ thống.

---

## 🏠 Hướng dẫn Triển khai Homelab (Ubuntu / Debian / Proxmox / Mini PC)

WrenPanel cực kỳ phù hợp để tự build một máy chủ Web Hosting ngay tại nhà trên:
* Máy chủ vật lý / PC cũ / Mini PC (Intel NUC, HP Elitedesk, ThinkCentre...)
* Proxmox VE (chạy trên LXC Container hoặc VM Ubuntu 22.04/24.04 LTS)
* Raspberry Pi 4/5 (Ubuntu Server 64-bit)

### 1. Chuẩn bị máy chủ Homelab
Cài đặt hệ điều hành **Ubuntu 22.04 LTS**, **Ubuntu 24.04 LTS** hoặc **Debian 12**.

Cài đặt các gói phụ thuộc cơ bản (Nginx & Git):
```bash
sudo apt update
sudo apt install -y git curl wget nginx
```

### 2. Cài đặt WrenPanel chỉ với 1 bước

Clone repo và chạy script cài đặt tự động:
```bash
git clone https://github.com/nguyenquocanhz/wrenpanel.git
cd wrenpanel
sudo bash scripts/install-ubuntu.sh
```

Script sẽ tự động:
1. Tạo user hệ thống `wrenpanel`.
2. Tạo toàn bộ cấu trúc thư mục FHS Linux (`/opt/wrenpanel/`, `/etc/wrenpanel/`, `/var/lib/wrenpanel/`, `/var/log/wrenpanel/`).
3. Cài đặt các binary vào `/opt/wrenpanel/bin/`.
4. Cài đặt và kích hoạt 2 dịch vụ systemd:
   * `wrenpanel-worker.service` (Root worker)
   * `wrenpanel.service` (Web API panel)

### 3. Truy cập Control Panel
* Mở trình duyệt và truy cập:
  ```
  http://<IP-LAN-HOMELAB>:8080
  ```
  *(Ví dụ: `http://192.168.1.100:8080`)*

---

## 🌐 Cấu hình Mạng & Tên miền cho Homelab

### Trường hợp 1: Có IP Tĩnh / Mở Port Router (Port Forwarding)
Cấu hình Port Forwarding trên Router nhà mạng về IP cục bộ của máy chủ Homelab:
* Port `80` (HTTP) -> `IP-LAN:80`
* Port `443` (HTTPS) -> `IP-LAN:443`
* Port `8080` (WrenPanel Dashboard) -> `IP-LAN:8080` (khuyên dùng VPN Wireguard/Tailscale để truy cập panel an toàn thay vì mở thẳng ra ngoài).

### Trường hợp 2: Không có IP tĩnh / Bị chặn Port (CGNAT) — Khuyên dùng Cloudflare Tunnel
Nếu mạng gia đình bị dính CGNAT (không mở được port 80/443), bạn có thể dùng **Cloudflare Tunnel (`cloudflared`)**:
```bash
# Cài cloudflared
curl -L --output cloudflared.deb https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
sudo dpkg -i cloudflared.deb

# Kết nối tunnel trỏ domain về cổng Nginx Homelab (cổng 80)
cloudflared tunnel run <TÊN-TUNNEL>
```
Toàn bộ traffic web và chứng chỉ SSL sẽ tự động được Cloudflare định tuyến thẳng về máy chủ Homelab của bạn mà không cần mở bất kỳ cổng nào trên Router!

---

## 📂 Cấu trúc thư mục FHS trên Server

```
/opt/wrenpanel/
├── bin/wrenpanel                 # Binary chính (user thường)
├── bin/wrenpanel-worker          # Binary root worker (chạy root)
├── node/<version>/               # Runtime Node.js cài biệt lập
└── python/<version>/             # Runtime Python cài biệt lập

/etc/wrenpanel/
├── nginx/vhosts/<fqdn>.conf      # Nguồn sự thật duy nhất cho vhost
├── apache/vhosts/<fqdn>.conf     # Vhost nội bộ cho Apache (nếu dùng)
└── ssl/<fqdn>/                   # Chứng chỉ SSL Let's Encrypt

/var/lib/wrenpanel/
├── wrenpanel.db                  # SQLite database (chế độ WAL)
└── backups/staging/              # Thư mục tạm nén tar.zst trước khi đẩy

/var/log/wrenpanel/
├── panel.log                     # Log HTTP API
├── worker.log                    # Log kiểm toán toàn bộ lệnh Root Worker
└── audit.log                     # Log thao tác admin (append-only)

/run/wrenpanel/
└── worker.sock                   # Unix socket giao tiếp (quyền 0660)

/home/<username>/                 # Cấu trúc hosting account chuẩn cPanel
├── public_html/                  # Webroot website
├── logs/                         # Log riêng theo domain
└── tmp/                          # PHP session riêng theo user
```

---

## 💻 Môi trường Phát triển (Development)

### Biên dịch Binary từ mã nguồn:
```bash
# 1. Build frontend Vue 3
cd web
npm install
npm run build
cd ..

# 2. Biên dịch binary Go cho Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/linux-amd64/wrenpanel ./cmd/wrenpanel
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/linux-amd64/wrenpanel-worker ./cmd/wrenpanel-worker
```

### Chạy kiểm thử tự động (Unit & Integration tests):
```bash
go test -v ./...
```

---

## 📜 Giấy phép
Mã nguồn phát hành theo giấy phép MIT License.
