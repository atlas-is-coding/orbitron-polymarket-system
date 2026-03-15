#!/usr/bin/env bash

# Universal Setup Script for Polytrade Bot (Linux, macOS, Windows via MSYS/Git Bash)
# This script installs dependencies (Go 1.24.4, Node.js), configures the project, and builds it.

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

GO_VERSION="1.24.4"

echo -e "${GREEN}Starting Orbitron Universal Setup...${NC}"

# 1. OS and Arch Detection
OS="$(uname -s)"
ARCH="$(uname -m)"

echo -e "${YELLOW}Detected OS: $OS, Arch: $ARCH${NC}"

# Normalize OS
case "$OS" in
    Linux*)     PLATFORM="linux";;
    Darwin*)    PLATFORM="darwin";;
    CYGWIN*|MINGW*|MSYS*) PLATFORM="windows";;
    *)          PLATFORM="unknown"
esac

# Normalize Arch for Go
case "$ARCH" in
    x86_64)  GOARCH="amd64" ;;
    amd64)   GOARCH="amd64" ;;
    arm64)   GOARCH="arm64" ;;
    aarch64) GOARCH="arm64" ;;
    *)       GOARCH="unknown" ;;
esac

if [ "$PLATFORM" == "unknown" ] || [ "$GOARCH" == "unknown" ]; then
    echo -e "${RED}Unsupported OS or Architecture: $OS / $ARCH${NC}"
    exit 1
fi

# 2. Check and Install Go
check_go() {
    if command -v go >/dev/null 2>&1; then
        INSTALLED_GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        if [ "$INSTALLED_GO_VERSION" == "$GO_VERSION" ]; then
            echo -e "${GREEN}Golang $GO_VERSION is already installed.${NC}"
            return 0
        else
            echo -e "${YELLOW}Found Go $INSTALLED_GO_VERSION, but require $GO_VERSION. Will install locally.${NC}"
            return 1
        fi
    fi
    return 1
}

install_go() {
    echo -e "${YELLOW}Installing Golang $GO_VERSION for $PLATFORM-$GOARCH...${NC}"
    GO_TAR="go$GO_VERSION.$PLATFORM-$GOARCH.tar.gz"
    
    if [ "$PLATFORM" == "windows" ]; then
        GO_ZIP="go$GO_VERSION.windows-$GOARCH.zip"
        echo -e "${YELLOW}Downloading $GO_ZIP...${NC}"
        curl -LO "https://go.dev/dl/$GO_ZIP"
        echo -e "${YELLOW}Extracting to ./local_go...${NC}"
        rm -rf ./local_go
        mkdir -p ./local_go
        unzip -q $GO_ZIP -d ./local_go
        rm $GO_ZIP
        export PATH="$(pwd)/local_go/go/bin:$PATH"
    else
        echo -e "${YELLOW}Downloading $GO_TAR...${NC}"
        curl -LO "https://go.dev/dl/$GO_TAR"
        echo -e "${YELLOW}Extracting to ./local_go...${NC}"
        rm -rf ./local_go
        mkdir -p ./local_go
        tar -C ./local_go -xzf $GO_TAR
        rm $GO_TAR
        export PATH="$(pwd)/local_go/go/bin:$PATH"
    fi
    
    if command -v go >/dev/null 2>&1; then
        echo -e "${GREEN}Golang successfully installed locally: $(go version)${NC}"
    else
        echo -e "${RED}Failed to install Golang. Please install manually.${NC}"
        exit 1
    fi
}

if ! check_go; then
    install_go
fi

# 3. Check and Install Node.js
check_node() {
    if command -v node >/dev/null 2>&1 && command -v npm >/dev/null 2>&1; then
        echo -e "${GREEN}Node.js $(node -v) and npm $(npm -v) are already installed.${NC}"
        return 0
    fi
    return 1
}

install_node() {
    echo -e "${YELLOW}Node.js/npm not found. Attempting to install...${NC}"
    if [ "$PLATFORM" == "darwin" ]; then
        if command -v brew >/dev/null 2>&1; then
            echo -e "${YELLOW}Installing via Homebrew...${NC}"
            brew install node
        else
            echo -e "${RED}Homebrew not found. Please install Node.js manually: https://nodejs.org/${NC}"
            exit 1
        fi
    elif [ "$PLATFORM" == "linux" ]; then
        if command -v apt-get >/dev/null 2>&1; then
            echo -e "${YELLOW}Installing via apt...${NC}"
            curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
            sudo apt-get install -y nodejs
        elif command -v yum >/dev/null 2>&1; then
            echo -e "${YELLOW}Installing via yum...${NC}"
            curl -fsSL https://rpm.nodesource.com/setup_20.x | sudo bash -
            sudo yum install -y nodejs
        else
            echo -e "${RED}Unsupported package manager. Please install Node.js manually: https://nodejs.org/${NC}"
            exit 1
        fi
    elif [ "$PLATFORM" == "windows" ]; then
        echo -e "${RED}Cannot automatically install Node.js on Windows. Please install it manually from https://nodejs.org/ and rerun this script.${NC}"
        exit 1
    fi
}

if ! check_node; then
    install_node
fi

# 4. Project Configuration (config.toml)
CONFIG_FILE="config.toml"
if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${YELLOW}config.toml not found. Creating a minimal template...${NC}"
    read -p "Enter your L1 wallet private key (hex, no 0x prefix, needed for trading): " PRIV_KEY
    
    cat <<EOF > $CONFIG_FILE
[webui]
enabled = true
listen = "127.0.0.1:8080"
jwt_secret = "changeme123"

[ui]
language = "en"

[auth]
private_key = "$PRIV_KEY"

[log]
level = "info"
format = "pretty"
EOF
    echo -e "${GREEN}config.toml created successfully.${NC}"
else
    echo -e "${GREEN}config.toml already exists. Skipping creation.${NC}"
fi

# 5. Build Frontend (Web UI)
echo -e "${YELLOW}Building Vue 3 Frontend...${NC}"
if [ -d "internal/webui/web" ]; then
    cd internal/webui/web
    echo -e "${YELLOW}Running npm install...${NC}"
    npm install --silent
    echo -e "${YELLOW}Running npm run build...${NC}"
    npm run build
    cd ../../../
    echo -e "${GREEN}Frontend built successfully.${NC}"
else
    echo -e "${RED}Directory internal/webui/web not found! Frontend build skipped.${NC}"
fi

# 6. Build Backend (Go)
echo -e "${YELLOW}Building Go Backend...${NC}"
echo -e "${YELLOW}Running go mod tidy...${NC}"
go mod tidy

BIN_NAME="orbitron-polytrade-bot"
if [ "$PLATFORM" == "windows" ]; then
    BIN_NAME="orbitron-polytrade-bot.exe"
fi

echo -e "${YELLOW}Compiling main binary ($BIN_NAME)...${NC}"
go build -o $BIN_NAME -ldflags="-X 'github.com/atlas-is-coding/orbitron-polymarket-system/internal/license.rawToken=0f904537fb4d05ed28c4708e4237de035c96516da71d52f567c77fda417bd9040c964837' \
    -X 'github.com/atlas-is-coding/orbitron-polymarket-system/internal/license.LicenseServerURL=https://getorbitron.net/api/v1/license'" \
    ./cmd/bot/

if [ -f "$BIN_NAME" ]; then
    echo -e "${GREEN}Backend built successfully: $BIN_NAME${NC}"
else
    echo -e "${RED}Failed to build backend!${NC}"
    exit 1
fi

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Setup Complete!${NC}"
echo -e "${GREEN}You can now run the bot using:${NC}"
echo -e "${YELLOW}  ./$BIN_NAME --config config.toml${NC}"
echo -e "${GREEN}========================================${NC}"