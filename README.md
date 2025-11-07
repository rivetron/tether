<p align="center">
  <img src="logo.svg" alt="Helix Logo">
</p>

# 🎪 Helix

![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go&logoColor=white)
![Pre-commit](https://img.shields.io/badge/Pre--commit-hooks-FAB040?logo=precommit&logoColor=white)
![Commitlint](https://img.shields.io/badge/Commitlint-enforced-000000?logo=commitlint&logoColor=white)
![Testing](https://img.shields.io/badge/Testing-enabled-success?logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-yellow?logo=open-source-initiative&logoColor=white)

**Helix** is a structured Go starter template designed for building modular, maintainable, and scalable applications. It follows Go best practices and industry standards to help you bootstrap production-ready projects with minimal setup.

---

## ⚡ Features

- **Clean Architecture**: Domain-driven design with clear separation of concerns
- **Pre-configured Tooling**:
  - **Pre-commit**: Hooks to enforce consistency before commits.
  - **Testing**: Comprehensive testing setup with mocks and fixtures
- **Modular Structure**: Organized packages for scalability and maintainability
- **CI/CD**: GitHub Actions workflows for testing and deployment
- **Security**: Built-in security middleware and best practices

---

## 🚀 Getting Started

### Prerequisites

Ensure you have the following installed:

- **Go**: v1.24 or later
- **Make**: For running build commands (optional)

---

### ⚙️ Installation

1. **Clone the repository:**

   ```bash
   git clone https://github.com/Kosha-Nirman/helix.git
   cd helix
   ```

2. **Install dependencies:**

   ```bash
   go mod tidy
   ```

3. **Run the application:**

   ```bash
   go run src/cmd/main.go
   ```

4. **Or use Make commands:**

   ```bash
   make run
   ```

---

## 📜 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE.md) file for details.

---

## 🙌 Acknowledgments

- Inspired by **Go best practices** and **clean architecture principles**
- Built for developers who value **maintainable** and **scalable** code
- Named **Helix** for its spiral structure that represents growth and evolution
- Made with ❤️ to accelerate Go development and reduce boilerplate
