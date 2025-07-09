# Connection Pooling в микросервисном проекте

## Обзор

В проекте настроен connection pooling для всех типов соединений:
- **PostgreSQL** - для user сервиса
- **Redis** - для geo сервиса  
- **gRPC** - для межсервисного взаимодействия
- **HTTP** - для внешних API (Dadata)

## Настройки по сервисам

### User Service (PostgreSQL)

```yaml
# Database connection pooling
DB_MAX_OPEN_CONNS: 25        # Максимум открытых соединений
DB_MAX_IDLE_CONNS: 5         # Максимум неактивных соединений в пуле
DB_CONN_MAX_LIFETIME: 5m     # Максимальное время жизни соединения
DB_CONN_MAX_IDLE_TIME: 5m    # Максимальное время неактивности соединения
```

**Рекомендации:**
- `MaxOpenConns`: 25-50 для средних нагрузок
- `MaxIdleConns`: 5-10% от MaxOpenConns
- `ConnMaxLifetime`: 5-15 минут
- `ConnMaxIdleTime`: 5-10 минут

### Geo Service (Redis)

```yaml
# Redis connection pooling
REDIS_POOL_SIZE: 10          # Размер пула соединений
REDIS_MIN_IDLE_CONNS: 5      # Минимум неактивных соединений
REDIS_MAX_RETRIES: 3         # Максимум попыток переподключения
REDIS_DIAL_TIMEOUT: 5s       # Таймаут установки соединения
REDIS_READ_TIMEOUT: 3s       # Таймаут чтения
REDIS_WRITE_TIMEOUT: 3s      # Таймаут записи
```

**Рекомендации:**
- `PoolSize`: 10-50 в зависимости от нагрузки
- `MinIdleConns`: 50% от PoolSize
- `MaxRetries`: 3-5 попыток
- Таймауты: 3-10 секунд

### gRPC Connections (Proxy, Auth)

```yaml
# gRPC connection pooling
GRPC_DIAL_TIMEOUT: 30s           # Таймаут установки соединения
GRPC_MIN_CONNECT_TIMEOUT: 30s    # Минимальное время соединения
GRPC_BACKOFF_BASE_DELAY: 2s      # Базовая задержка при retry
GRPC_BACKOFF_MAX_DELAY: 60s      # Максимальная задержка при retry
GRPC_BACKOFF_MULTIPLIER: 1.6     # Множитель задержки
GRPC_MAX_RETRIES: 3              # Максимум попыток
```

**Рекомендации:**
- `DialTimeout`: 30-60 секунд
- `MinConnectTimeout`: 30-60 секунд
- `BackoffBaseDelay`: 1-5 секунд
- `BackoffMaxDelay`: 30-120 секунд
- `BackoffMultiplier`: 1.5-2.0

### HTTP Client (Dadata API)

```go
httpClient := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,    // Максимум неактивных соединений
        MaxIdleConnsPerHost: 10,     // Максимум неактивных соединений на хост
        IdleConnTimeout:     90s,    // Таймаут неактивного соединения
        DisableCompression:  false,  // Включить сжатие
    },
}
```

## Мониторинг

### Prometheus метрики

Для мониторинга connection pooling используются стандартные метрики:

**PostgreSQL:**
- `pg_stat_activity` - активные соединения
- `pg_stat_database` - статистика по БД

**Redis:**
- `redis_connected_clients` - подключенные клиенты
- `redis_used_memory` - используемая память

**gRPC:**
- `grpc_client_started_total` - начатые запросы
- `grpc_client_handled_total` - обработанные запросы
- `grpc_client_msg_received_total` - полученные сообщения
- `grpc_client_msg_sent_total` - отправленные сообщения

### Grafana Dashboard

Создайте дашборд с панелями:
1. **Connection Pool Status** - статус пулов соединений
2. **Connection Errors** - ошибки соединений
3. **Response Times** - время отклика
4. **Throughput** - пропускная способность

## Troubleshooting

### Частые проблемы

1. **Too many connections**
   - Увеличьте `MaxOpenConns` для PostgreSQL
   - Проверьте утечки соединений

2. **Connection timeouts**
   - Увеличьте таймауты
   - Проверьте сетевую связность

3. **Slow responses**
   - Оптимизируйте размеры пулов
   - Проверьте нагрузку на БД/Redis

### Логирование

Все сервисы логируют информацию о connection pooling:

```json
{
  "level": "info",
  "msg": "Database connection established with connection pooling",
  "max_open_conns": 25,
  "max_idle_conns": 5,
  "conn_max_lifetime": "5m",
  "conn_max_idle_time": "5m"
}
```

## Производительность

### Бенчмарки

Типичные улучшения с connection pooling:
- **PostgreSQL**: 30-50% улучшение latency
- **Redis**: 20-40% улучшение throughput
- **gRPC**: 25-35% улучшение response time

### Настройка под нагрузку

Для высоких нагрузок:
1. Увеличьте размеры пулов
2. Настройте таймауты
3. Мониторьте метрики
4. Используйте load balancing

## Безопасность

1. **Ограничение соединений** - предотвращает DoS
2. **Таймауты** - избегает зависших соединений
3. **Retry логика** - устойчивость к сбоям
4. **Мониторинг** - раннее обнаружение проблем 