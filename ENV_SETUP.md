# Настройка переменных окружения

## Создание .env файла

1. Скопируйте файл `env.example` в `.env`:
```bash
cp env.example .env
```

2. Отредактируйте `.env` файл, заменив значения на свои:

```bash
# Обязательные настройки
JWT_SECRET=your-super-secret-jwt-key-that-is-at-least-32-characters-long
DADATA_API_KEY=your-dadata-api-key
DADATA_SECRET_KEY=your-dadata-secret-key
```

## Настройки Connection Pooling

### User Service (PostgreSQL)

```bash
# Database connection pooling
DB_MAX_OPEN_CONNS=25        # Максимум открытых соединений
DB_MAX_IDLE_CONNS=5         # Максимум неактивных соединений
DB_CONN_MAX_LIFETIME=5m     # Максимальное время жизни соединения
DB_CONN_MAX_IDLE_TIME=5m    # Максимальное время неактивности
```

### Geo Service (Redis)

```bash
# Redis connection pooling
REDIS_POOL_SIZE=10          # Размер пула соединений
REDIS_MIN_IDLE_CONNS=5      # Минимум неактивных соединений
REDIS_MAX_RETRIES=3         # Максимум попыток переподключения
REDIS_DIAL_TIMEOUT=5s       # Таймаут установки соединения
REDIS_READ_TIMEOUT=3s       # Таймаут чтения
REDIS_WRITE_TIMEOUT=3s      # Таймаут записи
```

### Auth & Proxy Services (gRPC)

```bash
# gRPC connection pooling
GRPC_DIAL_TIMEOUT=30s           # Таймаут установки соединения
GRPC_MIN_CONNECT_TIMEOUT=30s    # Минимальное время соединения
GRPC_BACKOFF_BASE_DELAY=2s      # Базовая задержка при retry
GRPC_BACKOFF_MAX_DELAY=60s      # Максимальная задержка при retry
GRPC_BACKOFF_MULTIPLIER=1.6     # Множитель задержки
GRPC_MAX_RETRIES=3              # Максимум попыток
```

## Рекомендации по настройке

### Для разработки (низкая нагрузка)
```bash
DB_MAX_OPEN_CONNS=10
REDIS_POOL_SIZE=5
GRPC_DIAL_TIMEOUT=10s
```

### Для продакшена (высокая нагрузка)
```bash
DB_MAX_OPEN_CONNS=50
REDIS_POOL_SIZE=20
GRPC_DIAL_TIMEOUT=60s
```

### Для тестирования
```bash
DB_MAX_OPEN_CONNS=5
REDIS_POOL_SIZE=3
GRPC_DIAL_TIMEOUT=5s
```

## Запуск с .env файлом

Docker Compose автоматически загружает `.env` файл:

```bash
docker-compose up -d
```

## Проверка настроек

Проверить, что настройки загрузились правильно:

```bash
# Проверить ENV переменные в контейнере
docker-compose exec user env | grep DB_MAX_OPEN_CONNS
docker-compose exec geo env | grep REDIS_POOL_SIZE
docker-compose exec auth env | grep GRPC_DIAL_TIMEOUT
```

## Безопасность

⚠️ **Важно:**
- Никогда не коммитьте `.env` файл в git
- Добавьте `.env` в `.gitignore`
- Используйте разные значения для разных окружений
- Регулярно меняйте секретные ключи

## Troubleshooting

### Проблема: "load env fail"
- Проверьте синтаксис `.env` файла
- Убедитесь, что файл находится в корне проекта
- Проверьте права доступа к файлу

### Проблема: "JWT_SECRET must be at least 32 characters"
- Увеличьте длину JWT_SECRET до 32+ символов

### Проблема: "DADATA_API_KEY and DADATA_SECRET_KEY are required"
- Добавьте валидные ключи Dadata API 