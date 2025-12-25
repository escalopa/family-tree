#!/bin/bash

# VPS Initial Setup Script
# Run this script on your VPS to install Docker and prepare the environment

set -e

echo "🚀 Setting up VPS for Family Tree application..."

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root or with sudo"
    exit 1
fi

# Update system packages
echo "📦 Updating system packages..."
apt-get update
apt-get upgrade -y

# Install prerequisites
echo "📦 Installing prerequisites..."
apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg \
    lsb-release \
    git \
    ufw

# Install Docker
echo "🐳 Installing Docker..."
if ! command -v docker &> /dev/null; then
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    rm get-docker.sh
    echo "✅ Docker installed successfully"
else
    echo "✅ Docker is already installed"
fi

# Install Docker Compose
echo "🐳 Installing Docker Compose..."
DOCKER_COMPOSE_VERSION=$(curl -s https://api.github.com/repos/docker/compose/releases/latest | grep 'tag_name' | cut -d\" -f4)
curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
echo "✅ Docker Compose installed successfully"

# Configure firewall
echo "🔥 Configuring firewall..."
ufw --force enable
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow 80/tcp
ufw allow 443/tcp
echo "✅ Firewall configured"

# Create application user (optional but recommended)
APP_USER="familytree"
if ! id "$APP_USER" &>/dev/null; then
    echo "👤 Creating application user: $APP_USER"
    useradd -m -s /bin/bash $APP_USER
    usermod -aG docker $APP_USER
    echo "✅ User created and added to docker group"
else
    echo "✅ User $APP_USER already exists"
fi

# Create application directory
APP_DIR="/opt/family-tree"
echo "📁 Creating application directory: $APP_DIR"
mkdir -p $APP_DIR
chown -R $APP_USER:$APP_USER $APP_DIR

# Enable Docker to start on boot
echo "🔧 Enabling Docker service..."
systemctl enable docker
systemctl start docker

# Display versions
echo ""
echo "✅ Setup completed successfully!"
echo ""
echo "Installed versions:"
docker --version
docker-compose --version
echo ""
echo "Next steps:"
echo "1. Switch to application user: su - $APP_USER"
echo "2. Clone your repository to $APP_DIR"
echo "3. Configure .env file"
echo "4. Run the deployment script"
echo ""
echo "For the application user, you may want to:"
echo "  cd $APP_DIR"
echo "  git clone <your-repo-url> ."
echo ""
