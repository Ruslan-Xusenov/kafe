#!/bin/bash

# Configuration
# Skript o'z papkasini avtomatik aniqlaydigan qilindi
PROJECT_DIR="$(pwd)"
GITHUB_TOKEN="YOUR_GITHUB_TOKEN_HERE"
REPO_URL="https://Ruslan-Xusenov:${GITHUB_TOKEN}@github.com/Ruslan-Xusenov/kafe.git"
LOG_FILE="${PROJECT_DIR}/deploy.log"

# Function to log messages
log_info() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - [INFO] - $1" >> "$LOG_FILE"
}

log_error() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - [ERROR] - $1" >> "$LOG_FILE"
}

# Start Deployment
log_info "Deployment started."

# Navigate to project directory
cd "$PROJECT_DIR" || { log_error "Failed to navigate to project directory."; exit 1; }

# Update from GitHub
log_info "Syncing with GitHub..."
git remote set-url origin "$REPO_URL" >> "$LOG_FILE" 2>&1
if git pull origin main >> "$LOG_FILE" 2>&1; then
    log_info "GitHub sync successful."
else
    log_error "GitHub sync failed. Check your token or network."
    exit 1
fi

# Build and Restart Containers
log_info "Building and restarting Docker containers..."
if docker-compose up -d --build >> "$LOG_FILE" 2>&1; then
    log_info "Docker deployment successful."
else
    log_error "Docker deployment failed."
    exit 1
fi

# Cleanup
log_info "Cleaning up unused Docker images..."
docker image prune -f >> "$LOG_FILE" 2>&1

log_info "Deployment completed successfully. 🚀"
echo "Deployment successful. Logs can be found at ${LOG_FILE}"
