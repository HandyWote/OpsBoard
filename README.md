# OpsBoard

<div align="center">
  <h1>OpsBoard</h1>
  <p>try网运维部门任务认领平台</p>
  
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version">
  <img src="https://img.shields.io/badge/Vue-3.x-4FC08D?style=for-the-badge&logo=vue.js&logoColor=white" alt="Vue Version">
  <img src="https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker">
  <img src="https://img.shields.io/badge/PostgreSQL-336791?style=for-the-badge&logo=postgresql&logoColor=white" alt="PostgreSQL">
</div>

## 📖 项目简介

OpsBoard 是专为 try 网运维部门设计的任务认领与管理系统。该平台旨在提高团队协作效率，简化任务分配流程，让运维工作更加有序和高效。

### 🎯 核心功能

- **🔐 校园网认证登录**：集成校园网账号系统，安全便捷的用户认证
- **📝 任务发布管理**：支持发布经过审核的代码任务，确保任务质量
- **👥 成员接单系统**：团队成员可以自由浏览和接取感兴趣的任务
- **📊 任务状态跟踪**：实时监控任务进度，便于项目管理和协调
- **🔔 消息通知系统**：及时推送任务更新和重要信息

## 🛠️ 技术栈

### 前端技术
- **Vue 3**：采用最新的 Vue 3 框架，提供更优秀的性能和开发体验
- **Vue Router**：官方路由管理器，实现单页应用导航
- **Element Plus**：基于 Vue 3 的组件库，提供丰富的 UI 组件

### 后端技术
- **Go 1.21+**：高性能编程语言，提供出色的并发处理能力
- **Gin**：轻量级 Web 框架，提供快速的路由和中间件支持
- **GORM**：ORM 库，简化数据库操作

### 数据库
- **PostgreSQL**：功能强大的开源关系型数据库，提供稳定可靠的数据存储

### 部署与运维
- **Docker**：容器化部署，确保环境一致性
- **Docker Compose**：多容器应用编排工具
- **Nginx**：高性能 Web 服务器，用于反向代理和负载均衡

## 🚀 启动指南

### 环境变量准备

在启动项目之前，请先为前后端准备所需的环境变量：

- 后端：在 `backend/.env` 中填写数据库和 JWT 等配置，示例：

```env
PORT=9012
DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=opsboard
DB_PASSWORD=admin
DB_NAME=opsboard
DB_SSLMODE=disable
JWT_SECRET=please-change-me
```

- 前端：如有需要自定义 API 地址，可在 `frontend/.env.local` 中设置：

```env
VITE_API_BASE_URL=http://localhost:9012
```

> 注意：`.env.local` 不会被提交到版本库，可在本地安全存储敏感信息。

### 本地直接启动

1. 启动 PostgreSQL（可使用本地服务或 docker 容器）。
2. 启动后端：

   ```bash
   cd backend
   go mod tidy
   go run ./cmd/api
   ```

3. 启动前端：

   ```bash
   cd frontend
   npm install
   npm run dev
   ```

4. 浏览器访问 `http://localhost:5173`（具体端口取决于 Vite 输出）。

### Docker 一键启动

1. 确保已安装 Docker 与 Docker Compose。
2. 在项目根目录直接启动服务：

   ```bash
   docker compose up -d --build
   ```

3. 前端默认通过 `http://localhost:8080` 访问，后端和数据库会自动在容器内启动。
4. 如需停止，执行：

   ```bash
   docker compose down
   ```

## 📄 许可证

本项目采用 [MIT 许可证](LICENSE) - 查看 [LICENSE](LICENSE) 文件了解详情。

## 📞 联系我们

- 项目维护者：[HandyWote](https://github.com/HandyWote)
- 邮箱：handy@handywote.top
- 项目链接：[https://github.com/HandyWote/OpsBoard](https://github.com/HandyWote/OpsBoard)

---

<div align="center">
  <p>由 ❤️ 驱动，为 try 网运维部门打造</p>
</div>
