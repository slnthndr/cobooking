# CoBooking - Микросервисная платформа бронирования рабочих мест

Полноценная backend-платформа для сети коворкингов, построенная на микросервисной архитектуре с использованием паттернов Event-Driven Architecture (EDA) и Cache-Aside.

## 🚀 Стек технологий

*   **Язык:** Go (Golang 1.22)
*   **База данных:** PostgreSQL 15 (Транзакции, оптимистичные блокировки)
*   **Кэширование & Rate Limiting:** Redis 7
*   **Брокер сообщений:** RabbitMQ (Асинхронное взаимодействие микросервисов)
*   **Object Storage (Медиа):** MinIO (S3-совместимое хранилище)
*   **Роутинг & API:** API Gateway, Chi Router, JWT Authentication
*   **Инфраструктура:** Docker, Docker Compose, Multi-stage builds
*   **Observability:** Prometheus, Grafana

## 🏗 Архитектура

Система состоит из API Gateway и 5 независимых микросервисов:

1.  **API Gateway:** Единая точка входа. Валидация JWT токенов, маршрутизация, сбор метрик (Prometheus), Rate Limiting.
2.  **MS Auth:** Регистрация, аутентификация (Bcrypt), генерация Access/Refresh JWT токенов.
3.  **MS Places:** Управление рабочими пространствами. Реализует паттерн **Cache-Aside** через Redis для ускорения отдачи данных, загрузка фото в S3.
4.  **MS Booking:** Управление жизненным циклом бронирований. Использует ACID транзакции PostgreSQL для защиты от двойного бронирования (Double-booking protection).
5.  **MS Payments:** Инициализация платежей и прием вебхуков от провайдеров.
6.  **MS Notifications:** Фоновый воркер. Читает очередь `EventsQueue` из RabbitMQ и рассылает email/push уведомления пользователям.

## ⚙️ Быстрый старт

Для запуска проекта вам потребуется только установленный `Docker` и `Docker Compose`.

```bash
# 1. Клонируем репозиторий
git clone https://github.com/ВАШ_НИК/cobooking.git
cd cobooking

# 2. Запускаем всю инфраструктуру
docker compose up --build -d
