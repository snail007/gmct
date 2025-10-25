<div align="center">

# GMCT

<img src="https://mirrors.host900.com/https://raw.githubusercontent.com/snail007/gmc/master/doc/images/logo2.png" width="200" alt="GMC Logo"/>

### 🚀 [GMC](https://github.com/snail007/gmc) 框架的官方工具链

一个强大的命令行工具集，为 GMC Web 框架提供项目脚手架、热编译、资源打包等功能

[![Go Report](https://goreportcard.com/badge/github.com/snail007/gmct)](https://goreportcard.com/report/github.com/snail007/gmct)
[![LICENSE](https://img.shields.io/github/license/snail007/gmct)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/snail007/gmct)](go.mod)

[English](README_EN.md) | 简体中文

[📖 完整文档](https://snail007.github.io/gmc/zh/#/?id=gmct-%e5%b7%a5%e5%85%b7%e9%93%be) | [🎯 快速开始](#-快速开始) | [💡 核心功能](#-核心功能) | [🔧 命令列表](#-命令列表)

</div>

---

## 📋 目录

- [简介](#-简介)
- [核心功能](#-核心功能)
- [安装](#-安装)
- [快速开始](#-快速开始)
- [命令列表](#-命令列表)
- [项目生成](#-项目生成)
- [热编译开发](#-热编译开发)
- [资源打包](#-资源打包)
- [代码生成](#-代码生成)
- [实用工具](#-实用工具)
- [使用示例](#-使用示例)
- [配置文件](#-配置文件)
- [贡献指南](#-贡献指南)
- [许可证](#-许可证)

---

## 🎯 简介

**GMCT**（GMC Toolchain）是 [GMC](https://github.com/snail007/gmc) Web 框架的官方命令行工具链，提供了从项目创建到部署的完整开发工具集。

### 为什么选择 GMCT？

- 🚀 **提升开发效率** - 热编译省去手动重启的麻烦，专注编码
- 📦 **简化部署流程** - 资源打包后可生成单一二进制文件
- 🎨 **标准化项目结构** - 统一的项目脚手架，便于团队协作
- 🛠️ **丰富的工具集** - 30+ 实用命令，覆盖开发全流程
- ⚡ **开箱即用** - 无需复杂配置，一键开始开发

---

## ✨ 核心功能

### 🏗️ 项目管理
- **项目脚手架** - 一键生成标准化的 Web/API 项目结构
- **热编译** - 开发时自动监听文件变化，自动编译重启
- **资源打包** - 将静态文件、模板、i18n 打包进二进制文件

### 🔧 代码生成
- **控制器生成** - 自动生成控制器代码和 CRUD 方法
- **模型生成** - 基于数据库表自动生成模型代码
- **视图生成** - 自动生成列表、详情、编辑视图文件

### 🛠️ 开发工具
- **测试覆盖率** - 智能测试覆盖率统计和报告
- **SSH 工具** - 便捷的远程文件复制和命令执行
- **Docker 集成** - 在容器中编译和运行程序
- **性能分析** - 集成 pprof 性能分析工具

### 📦 实用工具
- **文件管理** - 文件查重、下载等实用功能
- **加密工具** - MD5 哈希计算
- **随机生成** - 随机字符串、随机数生成
- **URL 编解码** - URL 编码解码工具
- **TLS 工具** - 证书信息查看和保存

---

## 📦 安装

### 环境要求

- Go 1.20 或更高版本

### 方式一：使用 go install（推荐）

```bash
go install github.com/snail007/gmct/cmd/gmct@latest

# 验证安装
gmct version
```

### 方式二：使用安装脚本

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

请从 [Releases](https://github.com/snail007/gmct/releases) 页面下载对应平台的预编译二进制文件：
- macOS: `gmct-mac-amd64.tar.gz`
- Windows: `gmct-windows-amd64.tar.gz`

### 验证安装

```bash
gmct version
gmct --help
```

---

## 🚀 快速开始

### 创建新项目

```bash
# 创建项目目录
mkdir myapp && cd myapp

# 生成 Web 项目（完整的 MVC 结构）
gmct new web

# 项目结构
# myapp/
# ├── conf/          # 配置文件
# ├── controller/    # 控制器
# ├── initialize/    # 初始化代码
# ├── router/        # 路由配置
# ├── static/        # 静态文件
# ├── views/         # 视图模板
# ├── grun.toml     # 运行配置
# └── main.go       # 入口文件
```

### 启动开发服务器

```bash
# 热编译模式运行（推荐）
gmct run

# 访问 http://localhost:7080
```

修改代码后，GMCT 会自动检测变化并重新编译启动服务器。

### 生成代码

```bash
# 基于数据库表生成模型
gmct model -n users

# 生成控制器（包含 CRUD 方法）
gmct controller -n User -t users

# 生成视图文件
gmct view -n user -t users
```

### 打包部署

```bash
# 打包静态资源
gmct static --dir ./static

# 打包模板文件
gmct tpl --dir ./views

# 打包国际化文件
gmct i18n --dir ./i18n

# 编译生成单一二进制文件
go build -ldflags "-s -w" -o myapp

# 部署
./myapp
```

---

## 🔧 命令列表

### 项目管理

| 命令 | 说明 | 示例 |
|------|------|------|
| `gmct new web` | 创建 Web 项目 | `gmct new web` |
| `gmct run` | 热编译运行项目 | `gmct run` |
| `gmct update` | 更新 GMCT | `gmct update` |
| `gmct uninstall` | 卸载 GMCT | `gmct uninstall` |

### 资源打包

| 命令 | 说明 | 示例 |
|------|------|------|
| `gmct tpl` | 打包/清理模板文件 | `gmct tpl --dir ./views` |
| `gmct static` | 打包/清理静态文件 | `gmct static --dir ./static` |
| `gmct i18n` | 打包/清理 i18n 文件 | `gmct i18n --dir ./i18n` |

### 代码生成

| 命令 | 说明 | 示例 |
|------|------|------|
| `gmct model` | 生成数据模型 | `gmct model -n users` |
| `gmct controller` | 生成控制器 | `gmct controller -n User -t users` |
| `gmct view` | 生成视图文件 | `gmct view -n user -t users` |

### 开发工具

| 命令 | 说明 | 示例 |
|------|------|------|
| `gmct cover` | 测试覆盖率分析 | `gmct cover` |
| `gmct pprof` | 性能分析工具 | `gmct pprof` |
| `gmct go` | Go 相关工具 | `gmct go api fmt.Println` |
| `gmct docker` | Docker 编译工具 | `gmct docker -g -- go build` |

### 实用工具

| 命令 | 说明 | 示例 |
|------|------|------|
| `gmct ssh` | SSH 文件传输/命令执行 | `gmct ssh -u user@host -c local:remote` |
| `gmct download` | 文件下载工具 | `gmct download url` |
| `gmct file` | 文件管理工具 | `gmct file` |
| `gmct md5` | MD5 哈希计算 | `gmct md5 "text"` |
| `gmct rand` | 随机数生成 | `gmct rand -l 16` |
| `gmct url` | URL 编解码 | `gmct url encode "text"` |
| `gmct tls` | TLS 证书工具 | `gmct tls info example.com:443` |
| `gmct ip` | IP 地址工具 | `gmct ip` |
| `gmct sleep` | 延时工具 | `gmct sleep 5s` |
| `gmct exec` | 命令执行工具 | `gmct exec` |
| `gmct gtag` | Git 标签工具 | `gmct gtag` |
| `gmct install` | 安装工具 | `gmct install go1.20` |

---

## 🏗️ 项目生成

### 生成 Web 项目

```bash
mkdir mywebapp && cd mywebapp
gmct new web

# 生成的项目结构
tree -L 2
```

生成的完整项目结构：

```
mywebapp/
├── conf/
│   └── app.toml           # 配置文件
├── controller/
│   └── demo.go            # 示例控制器
├── initialize/
│   ├── db.go              # 数据库初始化
│   ├── i18n.go            # 国际化初始化
│   ├── initialize.go      # 初始化入口
│   ├── log.go             # 日志初始化
│   ├── session.go         # Session 初始化
│   └── view.go            # 视图初始化
├── router/
│   └── router.go          # 路由配置
├── static/
│   ├── css/               # 样式文件
│   ├── js/                # JS 文件
│   └── images/            # 图片文件
├── views/
│   ├── layout/            # 布局模板
│   │   └── default.html
│   └── demo/              # 页面模板
│       └── index.html
├── grun.toml             # GMCT 运行配置
└── main.go               # 程序入口
```

---

## 🔥 热编译开发

### 基本使用

```bash
# 在项目根目录执行
gmct run
```

GMCT 会：
1. 监听项目文件变化
2. 自动编译项目
3. 自动重启服务器
4. 输出编译和运行日志

### 配置文件 grun.toml

```toml
[run]
# 监听的文件扩展名
watch_ext = [".go", ".toml", ".html", ".js", ".css"]

# 排除的目录
exclude_dir = [
    "vendor",
    ".git",
    ".idea",
    "tmp",
    "bin",
]

# 编译前执行的命令
before_build = []

# 编译命令
build_cmd = "go build -o tmp/app"

# 运行命令
run_cmd = "./tmp/app"

# 运行后执行的命令
after_run = []

# 延迟重启时间（毫秒）
restart_delay = 1000
```

### 自定义配置

```bash
# 只监听 .go 文件
gmct run --watch-ext .go

# 排除特定目录
gmct run --exclude vendor,.git

# 自定义编译命令
gmct run --build "go build -tags prod -o app"
```

---

## 📦 资源打包

将静态资源、模板、国际化文件打包进 Go 代码，实现单文件部署。

### 打包模板文件

```bash
# 打包模板目录
gmct tpl --dir ./views --ext .html

# 生成文件：当前目录下的 gmc_templates_bindata_随机数.go
```

使用打包后的模板：

```go
import _ "yourproject"  // 导入打包文件所在包

// GMC 框架会自动使用打包的模板
```

### 打包静态文件

```bash
# 打包静态资源目录
gmct static --dir ./static

# 生成文件：当前目录下的 gmc_static_bindata_随机数.go
```

### 打包国际化文件

```bash
# 打包 i18n 目录
gmct i18n --dir ./i18n

# 生成文件：当前目录下的 gmc_i18n_bindata_随机数.go
```

### 清理打包文件

```bash
# 清理模板打包文件（删除当前目录所有 gmc_templates_bindata_*.go）
gmct tpl --clean

# 清理静态文件打包文件（删除当前目录所有 gmc_static_bindata_*.go）
gmct static --clean

# 清理 i18n 打包文件（删除当前目录所有 gmc_i18n_bindata_*.go）
gmct i18n --clean
```

### 完整打包流程

```bash
# 1. 开发阶段：使用热编译
gmct run

# 2. 准备发布：打包所有资源
gmct static --dir ./static
gmct tpl --dir ./views
gmct i18n --dir ./i18n

# 3. 编译发布版本
go build -ldflags "-s -w" -o myapp

# 4. 部署（单一二进制文件）
./myapp
```

---

## 🎨 代码生成

### 生成模型

基于数据库表自动生成模型代码：

```bash
# 生成 users 表的模型
gmct model -n users

# 强制覆盖已存在的文件
gmct model -n users -f
```

生成的模型文件包含：
- 表结构定义
- CRUD 基础方法
- 查询辅助方法

### 生成控制器

```bash
# 生成 User 控制器（关联 users 表）
gmct controller -n User -t users

# 强制覆盖
gmct controller -n User -t users -f
```

生成的控制器包含：
- Index - 首页方法
- List - 列表页
- Detail - 详情页
- Create - 创建（GET显示表单，POST处理提交）
- Edit - 编辑（GET显示表单，POST处理提交）
- Delete - 删除方法

### 生成视图

```bash
# 生成视图文件（关联 users 表）
gmct view -n user -t users

# 强制覆盖
gmct view -n user -t users -f
```

生成的视图文件（在当前目录）：
- list.html - 列表页
- form.html - 表单页（添加/编辑共用）
- detail.html - 详情页

### 批量生成

**批量生成模型：**

创建 `tables.toml` 配置文件：

```toml
tables = ["users", "articles", "categories", "comments"]
```

然后执行：

```bash
# 自动批量生成所有表的模型
gmct model
```

**批量生成控制器：**

创建 `controllers.toml` 配置文件：

```toml
# 格式：["ControllerName:table_name"]
# 如果不指定表名，默认使用控制器名的小写形式
controllers = ["User:users", "Article:articles", "Category:categories", "Comment"]
```

然后执行：

```bash
# 自动批量生成所有控制器
gmct controller
```

---

## 🛠️ 实用工具

### 测试覆盖率

```bash
# 运行测试并生成覆盖率报告
gmct cover

# 详细输出
gmct cover -v

# 启用竞态检测
gmct cover -r

# 设置超时时间
gmct cover --timeout 30m

# 只测试当前目录
gmct cover --only

# 指定额外的覆盖包
gmct cover --coverpkg=pkg1,pkg2
```

### SSH 工具

```bash
# 复制文件到远程服务器
gmct ssh -u user@192.168.1.100:22:password -c ./local.txt:/tmp/remote.txt

# 执行远程命令
gmct ssh -u user@host:22:pass -e "ls -la"

# 执行脚本文件
gmct ssh -u user@host:22:pass -e @script.sh

# 使用代理
HTTP_PROXY=http://proxy:8080 gmct ssh -u user@host -c local:remote
```

### Docker 工具

在 Docker 容器中编译 Go 程序：

```bash
# 基础编译
gmct docker -g -- go build

# 设置环境变量
gmct docker -g -e GO111MODULE=off -- go build

# 编译 C 存档
gmct docker -g -- go build -buildmode=c-archive -o lib.a

# 自定义镜像
gmct docker --img golang:1.20 -g -- go build

# 指定容器名
gmct docker -n mybuilder -g -- go build
```

### 下载工具

```bash
# 下载文件
gmct download https://example.com/file.zip

# 使用 axel 多线程下载
gmct download --axel https://example.com/large-file.iso

# 带认证的下载
gmct download -u user:pass https://example.com/file.zip
```

### MD5 工具

```bash
# 计算字符串 MD5
gmct md5 "hello world"

# 计算文件 MD5
gmct md5 file.txt
```

### 随机工具

```bash
# 生成随机字符串（默认 32 字符）
gmct rand

# 指定长度
gmct rand -l 16

# 生成随机数（范围 1-100）
gmct rand -min 1 -max 100
```

### URL 工具

```bash
# URL 编码
gmct url encode "hello world"

# URL 解码
gmct url decode "hello%20world"
```

### TLS 工具

```bash
# 查看证书信息
gmct tls info example.com:443

# 保存证书
gmct tls save example.com:443 -o cert.pem

# 查看文件证书信息
gmct tls info cert.pem
```

### Go 工具

```bash
# 查询 Go API 版本
gmct go api fmt.Println

# 查询某个方法是在哪个 Go 版本加入的
gmct go api net/http.Client.Timeout
```

### 安装工具

```bash
# 安装指定版本的 Go
gmct install go1.20

# 安装 chgo（Go 版本管理）
gmct install chgo

# 配置 Go 环境
gmct install goenv
```

---

## 💡 使用示例

### 完整开发流程

```bash
# 1. 创建项目
mkdir blog && cd blog
gmct new web

# 2. 配置数据库（编辑 conf/app.toml）
# 配置 MySQL 连接信息

# 3. 生成代码
gmct model -n articles
gmct controller -n Article -t articles
gmct view -n article -t articles

# 4. 热编译开发
gmct run
# 访问 http://localhost:7080

# 5. 开发完成，准备发布
gmct static --dir ./static
gmct tpl --dir ./views

# 6. 编译
go build -ldflags "-s -w" -o blog

# 7. 部署
./blog
```

### 跨平台编译

```bash
# 使用 Docker 编译 Linux 版本
gmct docker -g -- go build -o app-linux

# 本地编译多平台版本
GOOS=linux GOARCH=amd64 go build -o app-linux
GOOS=windows GOARCH=amd64 go build -o app-windows.exe
GOOS=darwin GOARCH=amd64 go build -o app-macos
```

### 自动化部署

```bash
# 编译
gmct docker -g -- go build -ldflags "-s -w" -o myapp

# 部署到服务器
gmct ssh -u user@server:22:pass -c ./myapp:/opt/app/myapp
gmct ssh -u user@server:22:pass -e "systemctl restart myapp"
```

---

## ⚙️ 配置文件

### grun.toml 配置详解

```toml
[run]
# 监听的文件扩展名列表
# 当这些类型的文件变化时，会触发重新编译
watch_ext = [".go", ".toml", ".html", ".js", ".css"]

# 排除的目录列表
# 这些目录中的文件变化不会触发重新编译
exclude_dir = [
    "vendor",      # 依赖包目录
    ".git",        # Git 目录
    ".idea",       # IDE 配置
    "tmp",         # 临时文件
    "bin",         # 编译输出
    "node_modules" # 前端依赖
]

# 编译前执行的命令列表
# 可以用于代码生成、资源处理等
before_build = [
    # "go generate",
    # "npm run build"
]

# 编译命令
# 使用 Go 编译命令，输出到临时目录
build_cmd = "go build -o tmp/app"

# 运行命令
# 执行编译后的程序
run_cmd = "./tmp/app"

# 运行后执行的命令列表
# 可以用于打开浏览器、发送通知等
after_run = [
    # "open http://localhost:7080"
]

# 延迟重启时间（毫秒）
# 文件变化后等待一段时间再编译，避免频繁编译
restart_delay = 1000

# 是否打印详细日志
verbose = false

# 自定义环境变量
[run.env]
# APP_ENV = "development"
# DEBUG = "true"
```

### web.toml 配置详解

批量生成配置文件：

```toml
# 数据库配置
[database]
# 数据库类型
type = "mysql"
# 连接字符串
dsn = "root:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4"
# 表前缀
prefix = ""
# 表后缀
suffix = ""

# 生成配置
[generate]
# 是否覆盖已存在的文件
force = false
# 模型输出目录
model_dir = "./model"
# 控制器输出目录
controller_dir = "./controller"
# 视图输出目录
view_dir = "./views"

# 模块列表
[[module]]
# 控制器名称（首字母大写）
name = "User"
# 对应的数据库表名
table = "users"

[[module]]
name = "Article"
table = "articles"

[[module]]
name = "Category"
table = "categories"

[[module]]
name = "Comment"
table = "comments"
```

---

## 🎬 典型场景

### 场景 1：快速开始新项目

```bash
# 创建并初始化项目
mkdir myproject && cd myproject
gmct new web

# 立即开始开发
gmct run
```

### 场景 2：已有项目添加新模块

```bash
# 生成新模块的全部代码
gmct model -n products
gmct controller -n Product -t products
gmct view -n product -t products

# 代码会自动重新编译
```

### 场景 3：单文件部署

```bash
# 打包所有资源
gmct static --dir ./static
gmct tpl --dir ./views
gmct i18n --dir ./i18n

# 编译优化版本
go build -ldflags "-s -w" -trimpath -o app

# 只需要一个文件即可部署
ls -lh app
```

### 场景 4：持续集成

```bash
#!/bin/bash
# ci-build.sh

set -e

# 安装依赖
go mod download

# 运行测试
gmct cover -v

# 打包资源
gmct static --dir ./static
gmct tpl --dir ./views

# 编译多平台版本
GOOS=linux GOARCH=amd64 go build -o dist/app-linux
GOOS=windows GOARCH=amd64 go build -o dist/app-windows.exe
```

---

## 🔗 相关链接

- **GMC 框架**: [https://github.com/snail007/gmc](https://github.com/snail007/gmc)
- **完整文档**: [https://snail007.github.io/gmc/zh/](https://snail007.github.io/gmc/zh/)
- **GMCT 文档**: [https://snail007.github.io/gmc/zh/#/?id=gmct-%e5%b7%a5%e5%85%b7%e9%93%be](https://snail007.github.io/gmc/zh/#/?id=gmct-%e5%b7%a5%e5%85%b7%e9%93%be)
- **API 文档**: [https://pkg.go.dev/github.com/snail007/gmct](https://pkg.go.dev/github.com/snail007/gmct)

---

## 🤝 贡献指南

我们欢迎所有形式的贡献！

### 如何贡献

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

### 代码规范

- 遵循 Go 语言官方代码规范
- 添加必要的注释和文档
- 提交前运行 `go fmt` 和 `go vet`
- 为新功能添加测试用例

---

## 📜 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

---

## 💬 联系我们

- **Issues**: [https://github.com/snail007/gmct/issues](https://github.com/snail007/gmct/issues)
- **Discussions**: [https://github.com/snail007/gmct/discussions](https://github.com/snail007/gmct/discussions)

---

## ⭐ Star 历史

如果这个项目对你有帮助，请给我们一个 Star ⭐

---

<div align="center">

**[⬆ 回到顶部](#gmct)**

Made with ❤️ by the GMC Team

</div>
