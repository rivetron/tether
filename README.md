<p align="center">
  <img src="logo.svg" alt="Tether Logo">
</p>

# 🔗 Tether

![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go&logoColor=white)
![Pre-commit](https://img.shields.io/badge/Pre--commit-hooks-FAB040?logo=precommit&logoColor=white)
![Commitlint](https://img.shields.io/badge/Commitlint-enforced-000000?logo=commitlint&logoColor=white)
![Testing](https://img.shields.io/badge/Testing-enabled-success?logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-yellow?logo=open-source-initiative&logoColor=white)

**Tether** is a production-ready short URL service built on the Helix framework, demonstrating clean architecture and industry best practices.

---

## ⚡ Features

- **🔗 Short URL Generation**: Create short URLs with custom codes or auto-generated ones
- **⏰ TTL Support**: Set expiration times with default 1-month TTL (720 hours)
- **📊 Comprehensive Analytics**: Device, browser, geolocation, and referrer tracking
- **🔒 Security**: CORS, rate limiting, domain blocking, input validation
- **⚡ Performance**: Redis caching, MongoDB TTL indexes, optimized queries
- **🧹 Auto Cleanup**: Background cleanup of expired links with configurable retention
- **🐳 Production Ready**: Docker setup, health checks, graceful shutdown

---

## 🚀 Getting Started

### Prerequisites

Ensure you have the following installed:

- **Go**: v1.25 or later
- **Docker & Docker Compose**: For database services
- **Make**: For development commands

---

### 🐳 Quick Start with Docker

1. **Clone and setup:**

   ```bash
   git clone https://github.com/Kosha-Nirman/tether.git
   cd tether
   cp .env.example .env
   ```

2. **Start all services:**

   ```bash
   docker-compose up -d
   ```

3. **Run Tether:**

   ```bash
   make dev  # Development with hot reload
   # or
   make run  # Production build
   ```

4. **Verify installation:**

   ```bash
   curl http://localhost:8080/health
   ```

5. **Access API Documentation:**

   ```bash
   # Swagger UI
   open http://localhost:8080/docs/index.html

   # JSON API spec
   curl http://localhost:8080/docs/swagger.json
   ```

### 📖 API Usage

**Create a short link:**

```bash
curl -X POST http://localhost:8080/api/links \
  -H "Content-Type: application/json" \
  -d '{
    "original_url": "https://example.com",
    "custom_code": "my-link",
    "ttl_hours": 720
  }'
```

**Access short link:**

```bash
curl -L http://localhost:8080/abc123
```

**View analytics:**

```bash
curl http://localhost:8080/api/links/abc123/stats
```

### 📚 API Documentation

Tether includes comprehensive Swagger/OpenAPI documentation:

- **Interactive UI**: Visit `http://localhost:8080/docs/` for the Swagger UI
- **JSON Spec**: Available at `http://localhost:8080/docs/swagger.json`
- **YAML Spec**: Available at `http://localhost:8080/docs/swagger.yaml`

**Generate updated docs:**

```bash
make swagger  # Regenerate Swagger docs from code annotations
```

---

## 📜 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE.md) file for details.

---

## 🙌 Acknowledgments

- Inspired by **Go best practices** and **clean architecture principles**
- Built for developers who value **maintainable** and **scalable** code
- Named **Tether** for its spiral structure that represents growth and evolution
- Made with ❤️ to accelerate Go development and reduce boilerplate
