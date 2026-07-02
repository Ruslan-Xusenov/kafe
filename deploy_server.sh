#!/bin/bash
set -e

echo "Updating system and installing dependencies..."
export DEBIAN_FRONTEND=noninteractive
apt-get update -y
apt-get install -y postgresql postgresql-contrib nginx curl git build-essential certbot python3-certbot-nginx rsync

# Install Node.js 20
if ! command -v node &> /dev/null; then
    echo "Installing Node.js 20..."
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
    apt-get install -y nodejs
fi

# Install PM2
if ! command -v pm2 &> /dev/null; then
    echo "Installing PM2..."
    npm install -g pm2
fi

# Install Go 1.22
if ! command -v go &> /dev/null; then
    echo "Installing Go 1.22..."
    wget -q https://go.dev/dl/go1.22.2.linux-amd64.tar.gz
    rm -rf /usr/local/go
    tar -C /usr/local -xzf go1.22.2.linux-amd64.tar.gz
    rm go1.22.2.linux-amd64.tar.gz
    echo "export PATH=\$PATH:/usr/local/go/bin" >> /etc/profile
    export PATH=$PATH:/usr/local/go/bin
fi

export PATH=$PATH:/usr/local/go/bin

echo "Setting up PostgreSQL..."
sudo -u postgres psql -c "CREATE USER kafe_user WITH PASSWORD 'kafe_pass';" || true
sudo -u postgres psql -c "CREATE DATABASE kafe_db OWNER kafe_user;" || true
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE kafe_db TO kafe_user;" || true
# For superuser permissions required by some migrations if any
sudo -u postgres psql -c "ALTER USER kafe_user SUPERUSER;" || true

echo "Dependencies installed successfully!"
