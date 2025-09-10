# Order Processing Service

Микросервис для обработки заказов с использованием Go, PostgreSQL, Kafka и Redis.

## 🛠️ Зависимости

Перед запуском убедитесь, что установлены:

- **Docker** (версия 20.10+)
- **Docker Compose** (версия 2.0+)
- **Make** (обычно предустановлен на macOS/Linux)

### Установка зависимостей

#### macOS:
```bash
# Установка Docker Desktop (включает Docker Compose)
brew install --cask docker

# Или через Homebrew:
brew install docker docker-compose make

# Запустите Docker Desktop из Applications
```

#### Ubuntu/Debian:
```bash
# Установка Docker
sudo apt update
sudo apt install docker.io docker-compose make

# Добавьте пользователя в группу docker
sudo usermod -aG docker $USER
newgrp docker
```

#### Windows:
Установите [Docker Desktop](https://www.docker.com/products/docker-desktop) и используйте WSL2.

## 🚀 Быстрый старт

1. **Клонируйте репозиторий**:
   ```bash
   git clone <your-repository-url>
   cd <project-directory>
   ```

2. **Запустите сервисы**:
   ```bash
   make docker-up
   ```

3. **Проверьте работу**:
   Откройте в браузере: http://localhost:8081


### Просмотр логов:
```bash
# Логи основного сервера
docker logs go-server

# Логи в реальном времени
docker logs -f go-server
```
