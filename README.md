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

## 📈 Профилирование и нагрузочные бенчмарки

Каталог [bench/](bench/) содержит HTTP-бенчмарки на `testing.B`, поднимающие
chi-роутер с реальными `handler`/`service` и in-memory моками для
`OrderRepository` и `Cache`. Это даёт стабильные цифры без зависимости от
Postgres/Redis/Kafka и подходит для `pprof`/`trace`/`benchstat`.

### Запуск

```bash
# Все бенчмарки, 5 повторов для benchstat
go test -bench=. -benchmem -run=^$ -benchtime=3s -count=5 ./bench/ \
    > profiles/new.txt

# Сравнить с предыдущим прогоном
benchstat profiles/baseline.txt profiles/new.txt
```

### Сбор профилей и трейса

```bash
go test -bench=BenchmarkGenerateOrders -benchmem -run=^$ -benchtime=2s \
    -cpuprofile=profiles/cpu.out \
    -memprofile=profiles/mem.out \
    -trace=profiles/trace.out \
    ./bench/

go tool pprof -top -nodecount=20 profiles/cpu.out
go tool pprof -top -nodecount=20 -alloc_space profiles/mem.out
go tool trace profiles/trace.out
```

При работе с живым сервером в `cmd/main.go` подключается `net/http/pprof`
на том же порту — профили доступны на `/debug/pprof/`.

### Baseline (до оптимизаций)

Машина: `AMD Ryzen 5 3500U`, Go 1.26.1, Linux/amd64.

| Бенчмарк                          | ns/op   | B/op    | allocs/op |
|-----------------------------------|---------|---------|-----------|
| `BenchmarkGenerateOrders`         | 1296201 | 130125  |      1885 |
| `BenchmarkGetOrderByID_CacheHit`  |  255045 |   7213  |        74 |
| `BenchmarkGetOrderByID_DBPath`    |  263881 |   7669  |        75 |
| `BenchmarkGenerateRandomOrder`    |    3537 |    392  |        17 |

Полный вывод — `profiles/baseline.txt`.

## 🗂 Changelog оптимизаций

Каждая строка — отдельный коммит. Числа — `benchstat` дельта против предыдущего
прогона (полные `*.txt` лежат в `profiles/`).

| #  | Что               | Бенчмарк | ns/op | B/op | allocs/op | Как найдено |
|----|-------------------|----------|-------|------|-----------|-------------|
| 00 | bench-инфраструктура + baseline | — | — | — | — | — |

