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
| 01 | генератор: `math/rand/v2` + buffer-based `randomString` + `strconv` вместо `fmt.Sprintf` | `GenerateRandomOrder` | −30.5% | −8.2% | −17.6% (17→14) | pprof: `rand.Int31n` 9% CPU + `prefix+string(result)` 90ms + `fmt.Sprintf` 9MB allocs |
| 01 | то же                                                                                     | `GenerateOrders` (HTTP) | −12.2% | −4.6% | −15.9% (1885→1585) | то же |
| 02 | `middleware.RequestLogger`: буферим body только при `status >= 400` (+ подключили middleware в бенчмарках, как в проде) | `GenerateOrders` | −7.4% | **−45.5%** (228KB → 124KB) | ~ | code-review после iter1: `lrw.body.Write` копировал каждое тело без условия |
| 02 | то же | `GetOrderByID_CacheHit` | ~ | −10.9% | −1 alloc | то же |
| 02 | то же | `GetOrderByID_DBPath` | −19.4% | −10.4% | −1 alloc | то же |
| 03 | `sync.Pool` для `*loggingResponseWriter` (lrw уходит на кучу из-за interface-cast при `next.ServeHTTP`) | `GetOrderByID_CacheHit` | ~ | −0.99% | −1 alloc (80→79) | pprof: `middleware.RequestLogger.func1` 1.5MB flat / 21.5MB cum в alloc_space |
| 03 | то же | `GetOrderByID_DBPath` | ~ | −0.86% | −1 alloc (81→80) | то же |
| 04 | один backing array на все random-строки заказа: 11 отдельных `make([]byte,…)` (TrackNumber, Name, Phone, Zip, City, Address, Region, Email, Rid, ProductName, Brand) → 1 общий buf + `unsafe.String` со сдвигом-длиной | `GenerateRandomOrder` | **−11.4%** | −2.2% | **−71.4%** (14 → 4) | pprof: `randomString` 19.5MB flat, `GenerateRandomOrder` cum 58MB |
| 04 | то же | `GenerateOrders` (HTTP) | ~ | ~ | **−62.9%** (1591 → 590) | то же |

### Итоги (baseline → iter4)

`benchstat profiles/baseline.txt profiles/iter4.txt` (внимание: bench-сетап с iter2 включает middleware, в baseline его не было — поэтому численная асимметрия в `GetOrderByID_*` B/op/allocs не регрессия, а корректный учёт middleware-овер­хеда):

| Бенчмарк | sec/op | B/op | allocs/op |
|---|---:|---:|---:|
| `GenerateOrders` | **−9.6%** | −4.2% | **−68.7%** (1885 → 590) |
| `GenerateRandomOrder` | **−37.1%** | −10.2% | **−76.5%** (17 → 4) |
| `GetOrderByID_CacheHit` | ~ | +0.9%¹ | +5 allocs¹ |
| `GetOrderByID_DBPath` | ~ | +0.9%¹ | +5 allocs¹ |
| geomean | **−10.8%** | −3.3% | **−46.2%** |

¹ Реальный middleware был добавлен в bench только в iter2 — до этого мы не учитывали его аллокации. После iter2+iter3 чистый middleware-вклад уже минимизирован (см. ряд 03).

Главное, что осталось в pprof:
- `bytes.growSlice` 76 MB (stdlib `encoding/json` encoder pool — наш пул не помогает) — побеждается только сменой JSON-библиотеки (sonic/easyjson).
- `make([]domain.Order, count)` 95 MB — `sync.Pool` не сработал (под нагрузкой бенча `pool.New` с большим cap сам становился top-allocator).
- `uuid.UUID.String` (36 байт/заказ), `time.Time.MarshalJSON` — неустранимо без слома API.

