<div align="center">

# GMCT

<img src="https://mirrors.host900.com/https://raw.githubusercontent.com/snail007/gmc/master/doc/images/logo2.png" width="200" alt="GMC Logo"/>

### 🚀 Official Toolchain for [GMC](https://github.com/snail007/gmc) Framework

A powerful command-line toolkit providing project scaffolding, hot reloading, and resource bundling for GMC Web Framework

[![Go Report](https://goreportcard.com/badge/github.com/snail007/gmct)](https://goreportcard.com/report/github.com/snail007/gmct)
[![LICENSE](https://img.shields.io/github/license/snail007/gmct)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/snail007/gmct)](go.mod)

English | [简体中文](README.md)

[📖 Documentation](https://snail007.github.io/gmc/#/?id=gmct-tool-chain) | [🎯 Quick Start](#-quick-start) | [💡 Features](#-features) | [🔧 Commands](#-commands)

</div>

---

## 📋 Table of Contents

- [Introduction](#-introduction)
- [Features](#-features)
- [Installation](#-installation)
- [Quick Start](#-quick-start)
- [Commands](#-commands)
- [Project Generation](#-project-generation)
- [Hot Reloading](#-hot-reloading)
- [Resource Bundling](#-resource-bundling)
- [Code Generation](#-code-generation)
- [Utility Tools](#-utility-tools)
- [Examples](#-examples)
- [Configuration](#-configuration)
- [Contributing](#-contributing)
- [License](#-license)

---

## 🎯 Introduction

**GMCT** (GMC Toolchain) is the official command-line toolkit for the [GMC](https://github.com/snail007/gmc) Web Framework, providing a complete development toolset from project creation to deployment.

### Why GMCT?

- 🚀 **Boost Productivity** - Hot reloading eliminates manual restarts, letting you focus on coding
- 📦 **Simplify Deployment** - Bundle resources into a single binary file
- 🎨 **Standardize Projects** - Unified project scaffolding for better team collaboration
- 🛠️ **Rich Toolset** - 30+ practical commands covering the entire development lifecycle
- ⚡ **Ready to Use** - No complex configuration, start developing immediately

---

## ✨ Features

### 🏗️ Project Management
- **Project Scaffolding** - Generate standardized Web/API project structures with one command
- **Hot Reloading** - Auto-compile and restart on file changes during development
- **Resource Bundling** - Bundle static files, templates, and i18n into binaries

### 🔧 Code Generation
- **Controller Generation** - Auto-generate controller code with CRUD methods
- **Model Generation** - Auto-generate model code from database tables
- **View Generation** - Auto-generate list, detail, and edit view files

### 🛠️ Development Tools
- **Test Coverage** - Intelligent test coverage statistics and reports
- **SSH Tool** - Convenient remote file transfer and command execution
- **Docker Integration** - Compile and run programs in containers
- **Performance Analysis** - Integrated pprof performance profiling

### 📦 Utility Tools
- **File Management** - File deduplication, downloading, and more
- **Encryption Tools** - MD5 hash calculation
- **Random Generation** - Generate random strings and numbers
- **URL Encoding** - URL encoding/decoding utilities
- **TLS Tools** - Certificate information viewing and saving

---

## 📦 Installation

### Requirements

- Go 1.20 or higher

### Method 1: Using go install (Recommended)

```bash
go install github.com/snail007/gmct/cmd/gmct@latest

# Verify installation
gmct version
```

### Method 2: Using Installation Scripts

#### Linux AMD64

```bash
bash -c "$(curl -L https://github.com/snail007/gmct/raw/master/install.sh)" @ linux-amd64
```

#### Linux ARM64

```bash
bash -c "$(curl -L https://github.com/snail007/gmct/raw/master/install.sh)" @ linux-arm64
```

#### Linux ARM

```bash
bash -c "$(curl -L https://github.com/snail007/gmct/raw/master/install.sh)" @ linux-arm-v5
```

#### macOS / Windows

Please download pre-compiled binaries from the [Releases](https://github.com/snail007/gmct/releases) page:
- macOS: `gmct-mac-amd64.tar.gz`
- Windows: `gmct-windows-amd64.tar.gz`

### Verify Installation

```bash
gmct version
gmct --help
```

---

## 🚀 Quick Start

### Create a New Project

```bash
# Create project directory
mkdir myapp && cd myapp

# Generate Web project (full MVC structure)
gmct new web

# Project structure:
# myapp/
# ├── conf/          # Configuration files
# ├── controller/    # Controllers
# ├── initialize/    # Initialization code
# ├── router/        # Route configuration
# ├── static/        # Static files
# ├── views/         # View templates
# ├── grun.toml     # Run configuration
# └── main.go       # Entry point
```

### Start Development Server

```bash
# Run with hot reloading (recommended)
gmct run

# Visit http://localhost:7080
```

GMCT will automatically detect file changes and recompile/restart the server.

### Generate Code

```bash
# Generate model from database table
gmct model -n users

# Generate controller with CRUD methods
gmct controller -n User -t users

# Generate view files
gmct view -n user -t users
```

### Build for Production

```bash
# Bundle static resources
gmct static --dir ./static

# Bundle template files
gmct tpl --dir ./views

# Bundle i18n files
gmct i18n --dir ./i18n

# Build single binary
go build -ldflags "-s -w" -o myapp

# Deploy
./myapp
```

---

## 🔧 Commands

### Project Management

| Command | Description | Example |
|---------|-------------|---------|
| `gmct new web` | Create Web project | `gmct new web` |
| `gmct run` | Run with hot reloading | `gmct run` |
| `gmct update` | Update GMCT | `gmct update` |
| `gmct uninstall` | Uninstall GMCT | `gmct uninstall` |

### Resource Bundling

| Command | Description | Example |
|---------|-------------|---------|
| `gmct tpl` | Bundle/clean template files | `gmct tpl --dir ./views` |
| `gmct static` | Bundle/clean static files | `gmct static --dir ./static` |
| `gmct i18n` | Bundle/clean i18n files | `gmct i18n --dir ./i18n` |

### Code Generation

| Command | Description | Example |
|---------|-------------|---------|
| `gmct model` | Generate data model | `gmct model -n users` |
| `gmct controller` | Generate controller | `gmct controller -n User -t users` |
| `gmct view` | Generate view files | `gmct view -n user -t users` |

### Development Tools

| Command | Description | Example |
|---------|-------------|---------|
| `gmct cover` | Test coverage analysis | `gmct cover` |
| `gmct pprof` | Performance profiling | `gmct pprof` |
| `gmct go` | Go-related tools | `gmct go api fmt.Println` |
| `gmct docker` | Docker build tool | `gmct docker -g -- go build` |

### Utility Tools

| Command | Description | Example |
|---------|-------------|---------|
| `gmct ssh` | SSH file transfer/execution | `gmct ssh -u user@host -c local:remote` |
| `gmct download` | File download tool | `gmct download url` |
| `gmct file` | File management tool | `gmct file` |
| `gmct md5` | MD5 hash calculation | `gmct md5 "text"` |
| `gmct rand` | Random number generation | `gmct rand -l 16` |
| `gmct url` | URL encoding/decoding | `gmct url encode "text"` |
| `gmct tls` | TLS certificate tool | `gmct tls info example.com:443` |
| `gmct ip` | IP address tool | `gmct ip` |
| `gmct sleep` | Delay tool | `gmct sleep 5s` |
| `gmct exec` | Command execution | `gmct exec` |
| `gmct gtag` | Git tag tool | `gmct gtag` |
| `gmct install` | Installation tool | `gmct install go1.20` |

---

## 🏗️ Project Generation

### Generate Web Project

```bash
mkdir mywebapp && cd mywebapp
gmct new web
```

Generated project structure:

```
mywebapp/
├── conf/
│   └── app.toml           # Configuration file
├── controller/
│   └── demo.go            # Example controller
├── initialize/
│   ├── db.go              # Database initialization
│   ├── i18n.go            # i18n initialization
│   ├── initialize.go      # Init entry
│   ├── log.go             # Logger initialization
│   ├── session.go         # Session initialization
│   └── view.go            # View initialization
├── router/
│   └── router.go          # Route configuration
├── static/
│   ├── css/               # Stylesheets
│   ├── js/                # JavaScript files
│   └── images/            # Images
├── views/
│   ├── layout/            # Layout templates
│   │   └── default.html
│   └── demo/              # Page templates
│       └── index.html
├── grun.toml             # GMCT run configuration
└── main.go               # Program entry
```

---

## 🔥 Hot Reloading

### Basic Usage

```bash
# Run in project root directory
gmct run
```

GMCT will:
1. Monitor project file changes
2. Auto-compile the project
3. Auto-restart the server
4. Output compilation and runtime logs

### Configuration File: grun.toml

```toml
[run]
# File extensions to watch
watch_ext = [".go", ".toml", ".html", ".js", ".css"]

# Directories to exclude
exclude_dir = [
    "vendor",
    ".git",
    ".idea",
    "tmp",
    "bin",
]

# Commands to run before build
before_build = []

# Build command
build_cmd = "go build -o tmp/app"

# Run command
run_cmd = "./tmp/app"

# Commands to run after start
after_run = []

# Restart delay in milliseconds
restart_delay = 1000
```

---

## 📦 Resource Bundling

Bundle static resources, templates, and i18n files into Go code for single-file deployment.

### Bundle Template Files

```bash
# Bundle template directory
gmct tpl --dir ./views --ext .html

# Generates: gmc_templates_bindata_<random>.go in current directory
```

Using bundled templates:

```go
import _ "yourproject"  // Import package containing bundled file

// GMC framework will automatically use bundled templates
```

### Bundle Static Files

```bash
# Bundle static resource directory
gmct static --dir ./static

# Generates: gmc_static_bindata_<random>.go in current directory
```

### Bundle i18n Files

```bash
# Bundle i18n directory
gmct i18n --dir ./i18n

# Generates: gmc_i18n_bindata_<random>.go in current directory
```

### Clean Bundled Files

```bash
# Clean template bundles (removes all gmc_templates_bindata_*.go in current directory)
gmct tpl --clean

# Clean static bundles (removes all gmc_static_bindata_*.go in current directory)
gmct static --clean

# Clean i18n bundles (removes all gmc_i18n_bindata_*.go in current directory)
gmct i18n --clean
```

### Complete Bundling Workflow

```bash
# 1. Development: use hot reloading
gmct run

# 2. Prepare release: bundle all resources
gmct static --dir ./static
gmct tpl --dir ./views
gmct i18n --dir ./i18n

# 3. Build release version
go build -ldflags "-s -w" -o myapp

# 4. Deploy (single binary file)
./myapp
```

---

## 🎨 Code Generation

### Generate Model

Auto-generate model code from database tables:

```bash
# Generate model for users table
gmct model -n users

# Force overwrite existing files
gmct model -n users -f
```

Generated model includes:
- Table structure definition
- Basic CRUD methods
- Query helper methods

### Generate Controller

```bash
# Generate User controller (associated with users table)
gmct controller -n User -t users

# Force overwrite
gmct controller -n User -t users -f
```

Generated controller includes:
- Index - Home method
- List - List page
- Detail - Detail page
- Create - Create (GET shows form, POST handles submission)
- Edit - Edit (GET shows form, POST handles submission)
- Delete - Delete method

### Generate Views

```bash
# Generate view files (associated with users table)
gmct view -n user -t users

# Force overwrite
gmct view -n user -t users -f
```

Generated view files (in current directory):
- list.html - List page
- form.html - Form page (for add/edit)
- detail.html - Detail page

### Batch Generation

**Batch generate models:**

Create `tables.toml` configuration file:

```toml
tables = ["users", "articles", "categories", "comments"]
```

Then execute:

```bash
# Auto-generate models for all tables
gmct model
```

**Batch generate controllers:**

Create `controllers.toml` configuration file:

```toml
# Format: ["ControllerName:table_name"]
# If table name is not specified, lowercase controller name is used by default
controllers = ["User:users", "Article:articles", "Category:categories", "Comment"]
```

Then execute:

```bash
# Auto-generate all controllers
gmct controller
```

---

## 🛠️ Utility Tools

### Test Coverage

```bash
# Run tests and generate coverage report
gmct cover

# Verbose output
gmct cover -v

# Enable race detection
gmct cover -r

# Set timeout
gmct cover --timeout 30m

# Test current directory only
gmct cover --only

# Specify additional coverage packages
gmct cover --coverpkg=pkg1,pkg2
```

### SSH Tool

```bash
# Copy file to remote server
gmct ssh -u user@192.168.1.100:22:password -c ./local.txt:/tmp/remote.txt

# Execute remote command
gmct ssh -u user@host:22:pass -e "ls -la"

# Execute script file
gmct ssh -u user@host:22:pass -e @script.sh

# Using proxy
HTTP_PROXY=http://proxy:8080 gmct ssh -u user@host -c local:remote
```

### Docker Tool

Compile Go programs in Docker containers:

```bash
# Basic compilation
gmct docker -g -- go build

# Set environment variables
gmct docker -g -e GO111MODULE=off -- go build

# Build C archive
gmct docker -g -- go build -buildmode=c-archive -o lib.a

# Custom image
gmct docker --img golang:1.20 -g -- go build

# Specify container name
gmct docker -n mybuilder -g -- go build
```

### Download Tool

```bash
# Download file
gmct download https://example.com/file.zip

# Use axel for multi-threaded download
gmct download --axel https://example.com/large-file.iso

# Download with authentication
gmct download -u user:pass https://example.com/file.zip
```

### Other Utilities

```bash
# MD5 calculation
gmct md5 "hello world"
gmct md5 file.txt

# Random string generation
gmct rand -l 16

# URL encoding/decoding
gmct url encode "hello world"
gmct url decode "hello%20world"

# TLS certificate info
gmct tls info example.com:443

# Go API version query
gmct go api fmt.Println

# Install Go version
gmct install go1.20
```

---

## 💡 Examples

### Complete Development Workflow

```bash
# 1. Create project
mkdir blog && cd blog
gmct new web

# 2. Configure database (edit conf/app.toml)
# Configure MySQL connection

# 3. Generate code
gmct model -n articles
gmct controller -n Article -t articles
gmct view -n article -t articles

# 4. Hot reloading development
gmct run
# Visit http://localhost:7080

# 5. Prepare release
gmct static --dir ./static
gmct tpl --dir ./views

# 6. Build
go build -ldflags "-s -w" -o blog

# 7. Deploy
./blog
```

### Cross-Platform Build

```bash
# Use Docker to build for Linux
gmct docker -g -- go build -o app-linux

# Local multi-platform build
GOOS=linux GOARCH=amd64 go build -o app-linux
GOOS=windows GOARCH=amd64 go build -o app-windows.exe
GOOS=darwin GOARCH=amd64 go build -o app-macos
```

### Automated Deployment

```bash
# Build
gmct docker -g -- go build -ldflags "-s -w" -o myapp

# Deploy to server
gmct ssh -u user@server:22:pass -c ./myapp:/opt/app/myapp
gmct ssh -u user@server:22:pass -e "systemctl restart myapp"
```

---

## ⚙️ Configuration

### grun.toml Configuration

```toml
[run]
# File extensions to watch
watch_ext = [".go", ".toml", ".html", ".js", ".css"]

# Directories to exclude
exclude_dir = [
    "vendor",
    ".git",
    ".idea",
    "tmp",
    "bin",
    "node_modules"
]

# Commands before build
before_build = []

# Build command
build_cmd = "go build -o tmp/app"

# Run command
run_cmd = "./tmp/app"

# Commands after run
after_run = []

# Restart delay (milliseconds)
restart_delay = 1000

# Verbose logging
verbose = false

# Custom environment variables
[run.env]
# APP_ENV = "development"
# DEBUG = "true"
```

---

## 🔗 Related Links

- **GMC Framework**: [https://github.com/snail007/gmc](https://github.com/snail007/gmc)
- **Documentation**: [https://snail007.github.io/gmc/](https://snail007.github.io/gmc/)
- **API Docs**: [https://pkg.go.dev/github.com/snail007/gmct](https://pkg.go.dev/github.com/snail007/gmct)

---

## 🤝 Contributing

We welcome all forms of contributions!

### How to Contribute

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

### Code Guidelines

- Follow official Go code conventions
- Add necessary comments and documentation
- Run `go fmt` and `go vet` before committing
- Add test cases for new features

---

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 💬 Contact

- **Issues**: [https://github.com/snail007/gmct/issues](https://github.com/snail007/gmct/issues)
- **Discussions**: [https://github.com/snail007/gmct/discussions](https://github.com/snail007/gmct/discussions)

---

## ⭐ Star History

If this project helps you, please give us a Star ⭐

---

<div align="center">

**[⬆ Back to Top](#gmct)**

Made with ❤️ by the GMC Team

</div>
