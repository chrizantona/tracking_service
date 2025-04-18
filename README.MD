# **Проект: Система трекинга заказов в реальном времени**

---

## **Описание проекта**

Система позволяет клиентам создавать заказы, курьерам — принимать и выполнять их, а администраторам — контролировать выполнение заказов.

---

## **Технологический стек**

* **Backend:** Go, Gin, Docker, PostgreSQL, PostGIS, Kafka, WebSocket, JWT
* **Frontend:** HTML, CSS, JS, React, Axios, Google Maps API
* **Тесты:** GoMock, Dockertest, k6

---

## **Сущности**

### **1. User** (Базовая сущность)

| Поле      | Тип    | Описание                                           |
| ------------- | --------- | ---------------------------------------------------------- |
| id            | UUID      | Уникальный идентификатор            |
| email         | String    | Почта пользователя                        |
| password_hash | String    | Пароль (захеширован)                      |
| role          | Enum      | Роль пользователя (CLIENT, COURIER, ADMIN) |
| created_at    | Timestamp | Дата создания                                  |
| updated_at    | Timestamp | Дата обновления                              |

---

### **2. Client** (Наследуется от User)

| Поле | Тип | Описание                             |
| -------- | ------ | -------------------------------------------- |
| user_id  | UUID   | Ссылка на `User.id`                |
| name     | String | Имя клиента                        |
| phone    | String | Номер телефона                  |
| address  | String | Основной адрес доставки |

---

### **3. Courier** (Наследуется от User)

| Поле | Тип   | Описание                                       |
| -------- | -------- | ------------------------------------------------------ |
| user_id  | UUID     | Ссылка на `User.id`                          |
| name     | String   | Имя курьера                                  |
| status   | Enum     | Статус курьера (AVAILABLE, BUSY, OFFLINE) |
| location | geometry | Координаты курьера (PostGIS)          |
| rating   | Float    | Рейтинг курьера                          |

---

### **4. Admin** (Наследуется от User)

| Поле    | Тип | Описание                        |
| ----------- | ------ | --------------------------------------- |
| user_id     | UUID   | Ссылка на `User.id`           |
| permissions | JSON   | Права администратора |

---

### **5. Order**

| Поле         | Тип                                                    | Описание                                |
| ---------------- | --------------------------------------------------------- | ----------------------------------------------- |
| id               | UUID                                                      | Уникальный идентификатор |
| client_id        | UUID                                                      | Ссылка на `Client.id`                 |
| courier_id       | UUID (nullable)                                           | Ссылка на `Courier.id`                |
| status           | Enum (CREATED, ASSIGNED, IN_TRANSIT, DELIVERED, CANCELED) | Статус заказа                       |
| delivery_address | String                                                    | Адрес доставки                     |
| delivery_coords  | geometry (PostGIS)                                        | Координаты доставки           |
| created_at       | Timestamp                                                 | Время создания заказа        |
| updated_at       | Timestamp                                                 | Время обновления заказа    |

---

### **6. OrderStatusLog**

| Поле   | Тип    | Описание                                |
| ---------- | --------- | ----------------------------------------------- |
| id         | UUID      | Уникальный идентификатор |
| order_id   | UUID      | Ссылка на `Order.id`                  |
| status     | Enum      | Статус заказа                       |
| created_at | Timestamp | Время обновления статуса  |

---

### **7. Notification**

| Поле   | Тип    | Описание                                |
| ---------- | --------- | ----------------------------------------------- |
| id         | UUID      | Уникальный идентификатор |
| user_id    | UUID      | Получатель уведомления     |
| message    | String    | Текст уведомления               |
| is_read    | Boolean   | Прочитано или нет                |
| created_at | Timestamp | Время создания                     |

---

### **8. Rating**

| Поле   | Тип    | Описание                                |
| ---------- | --------- | ----------------------------------------------- |
| id         | UUID      | Уникальный идентификатор |
| courier_id | UUID      | Курьер                                    |
| order_id   | UUID      | Заказ                                      |
| rating     | Integer   | Оценка (1-5)                              |
| comment    | String    | Комментарий                          |
| created_at | Timestamp | Время создания                     |

---

## **Юзкейсы**

### **1. Клиент**

* Регистрация и логин
* Создание заказа
* Отслеживание статуса заказа в реальном времени
* Отмена заказа (если статус позволяет)
* Оценка курьера после доставки

---

### **2. Курьер**

* Регистрация и логин
* Установка статуса (доступен/занят)
* Получение уведомлений о новых заказах
* Принятие и выполнение заказа
* Обновление статуса в реальном времени

---

### **3. Админ**

* Логин
* Просмотр всех заказов и пользователей
* Блокировка пользователей
* Аналитика по статусам и заказам
* Редактирование статусов заказов

---

## **Роудмап с тестированием (8 недель)**

| Неделя | Задачи                                                        | Результат                                                                        | Тесты                                                                                                             |
| ------------ | ------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| **1**  | Архитектура, базовый сервер, JWT, PostGIS   | Сервер с авторизацией                                                  | Юнит-тесты для конфигураций сервера и подключения к БД                   |
| **2**  | API пользователей (регистрация, логин) | Рабочее API с авторизацией                                            | Юнит-тесты для регистрации и логина (валидные/невалидные данные) |
| **3**  | API заказов (CRUD)                                           | API для работы с заказами                                               | Юнит-тесты для CRUD-операций, проверки прав доступа                             |
| **4**  | API курьеров + поиск через PostGIS                | Логика поиска ближайших курьеров                             | Интеграционные тесты для PostGIS, поиск курьеров                                    |
| **5**  | Назначение через Kafka + старт фронта     | Автоматическое назначение заказа                            | Интеграционные тесты на Kafka (отправка и приём событий)                     |
| **6**  | WebSocket + трекинг курьера                           | Реальное обновление статусов в реальном времени | Интеграционные тесты для WebSocket и обновлений статуса                        |
| **7**  | Рейтинги и отмены заказов                     | Учет рейтинга при назначении заказов                      | Юнит-тесты на логику рейтинга и отмены                                                 |
| **8**  | Docker + нагрузочное тестирование            | Полный проект в Docker, выдерживает нагрузку              | Нагрузочные тесты через k6 (10k запросов, 1000+ соединений)                     |

---

## **Тестирование**

* [Юнит-тесты (GoMock)](https://github.com/uber-go/mock)
* [Интеграционные тесты (Dockertest)](https://github.com/ory/dockertest)
* Нагрузочные тесты через k6 (10k запросов, 1000+ соединений)

---

## **Запуск тестов**

Перейдите в корень репозитория и выполните:

```bash
go mod download
go test ./...
```
