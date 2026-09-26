#!/usr/bin/env bash
# ==============================================================================
# WrenPanel Installer for Ubuntu / Debian Linux
# Adheres strictly to FHS architecture and specifications in .agent/GEMINI.md
# ==============================================================================

set -euo pipefail

if [[ $EUID -ne 0 ]]; then
   echo "[ERROR] Script này phải được chạy với quyền root (sudo ./install-ubuntu.sh)"
   exit 1
fi

echo "=========================================================="
echo "    Bắt đầu cài đặt WrenPanel trên Linux Ubuntu/Debian    "
echo "=========================================================="

# 1. Tạo user hệ thống không đặc quyền 'wrenpanel'
if ! id -u wrenpanel >/dev/null 2>&1; then
    echo "[1/7] Tạo system user 'wrenpanel'..."
    useradd -r -s /usr/sbin/nologin -d /opt/wrenpanel -M wrenpanel
else
    echo "[1/7] System user 'wrenpanel' đã tồn tại."
fi

# 2. Khởi tạo các thư mục FHS Linux theo GEMINI.md
echo "[2/7] Khởi tạo cấu trúc thư mục FHS..."
mkdir -p /opt/wrenpanel/bin
mkdir -p /opt/wrenpanel/node
mkdir -p /opt/wrenpanel/python

mkdir -p /etc/wrenpanel/nginx/snippets
mkdir -p /etc/wrenpanel/nginx/vhosts
mkdir -p /etc/wrenpanel/apache/vhosts
mkdir -p /etc/wrenpanel/ssl

mkdir -p /var/lib/wrenpanel/backups/staging
mkdir -p /var/log/wrenpanel
mkdir -p /run/wrenpanel

# 3. Phân quyền thư mục
echo "[3/7] Cấu hình quyền truy cập an toàn..."
chown -R root:root /opt/wrenpanel
chown -R root:root /etc/wrenpanel
chown -R wrenpanel:wrenpanel /var/lib/wrenpanel
chown -R wrenpanel:wrenpanel /var/log/wrenpanel
chown -R root:wrenpanel /run/wrenpanel
chmod 775 /run/wrenpanel

# 4. Sao chép các binary Linux
echo "[4/7] Cài đặt binary vào /opt/wrenpanel/bin/..."
if [[ -f "bin/linux-amd64/wrenpanel" && -f "bin/linux-amd64/wrenpanel-worker" ]]; then
    cp bin/linux-amd64/wrenpanel /opt/wrenpanel/bin/wrenpanel
    cp bin/linux-amd64/wrenpanel-worker /opt/wrenpanel/bin/wrenpanel-worker
elif [[ -f "wrenpanel" && -f "wrenpanel-worker" ]]; then
    cp wrenpanel /opt/wrenpanel/bin/wrenpanel
    cp wrenpanel-worker /opt/wrenpanel/bin/wrenpanel-worker
else
    echo "[BUILD] Đang build trực tiếp binary Go trên máy chủ Linux..."
    go build -o /opt/wrenpanel/bin/wrenpanel ./cmd/wrenpanel
    go build -o /opt/wrenpanel/bin/wrenpanel-worker ./cmd/wrenpanel-worker
fi

chmod 755 /opt/wrenpanel/bin/wrenpanel
chmod 755 /opt/wrenpanel/bin/wrenpanel-worker

# 5. Cài đặt Nginx tích hợp (symlink nguồn sự thật)
echo "[5/7] Cấu hình Nginx..."
if command -v nginx >/dev/null 2>&1; then
    mkdir -p /etc/nginx/sites-enabled
fi

# 6. Cài đặt Systemd Services
echo "[6/7] Cài đặt Systemd Services..."
cp scripts/systemd/wrenpanel-worker.service /etc/systemd/system/wrenpanel-worker.service
cp scripts/systemd/wrenpanel.service /etc/systemd/system/wrenpanel.service

systemctl daemon-reload
systemctl enable wrenpanel-worker.service
systemctl enable wrenpanel.service

systemctl restart wrenpanel-worker.service
systemctl restart wrenpanel.service

echo "[7/7] Kiểm tra trạng thái dịch vụ..."
systemctl --no-pager status wrenpanel-worker.service || true
systemctl --no-pager status wrenpanel.service || true

echo "=========================================================="
echo "    WrenPanel đã cài đặt và khởi động thành công!         "
echo "    Địa chỉ truy cập: http://<IP-MÁY-CHỦ>:8080            "
echo "    API Endpoint:     http://<IP-MÁY-CHỦ>:8080/api/system/status"
echo "    Worker Log:       /var/log/wrenpanel/worker.log       "
echo "    Panel Log:        /var/log/wrenpanel/panel.log        "
echo "=========================================================="
