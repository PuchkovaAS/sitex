# 🖥️ Sitex — Внутренняя система управления задачами и сотрудниками

> Современное веб-приложение на Go для планирования задач, учёта рабочего времени и управления доступом сотрудников.  
> Идеально подходит для небольших команд и внутренних корпоративных решений.

---

## 🌟 Основные возможности

- ✅ Авторизация и аутентификация пользователей
- ✅ Управление задачами и назначение ответственных
- ✅ Календарь рабочих дней и учет отпусков
- ✅ Ролевая модель доступа (админ / пользователь)
- ✅ Адаптивный UI с использованием Tailwind CSS
- ✅ Работа через SSH + Docker + PostgreSQL
- ✅ Автоматическая инициализация БД из дампа

---

## 🚀 Быстрый старт

### Требования

- Go 1.21+
- PostgreSQL 16
- Docker & Docker Compose (опционально, но рекомендуется)

---

### ▶️ Запуск через Docker (рекомендуется)

```bash
# 1. Клонируй репозиторий
git clone https://github.com/PuchkovaAS/sitex.git
cd sitex

# 2. Подготовь .env
cp .env.example .env

# 3. Запусти приложение
docker-compose up --build
```

➡️ Приложение доступно по адресу: [http://localhost:3000](http://localhost:3000)

> 💡 При первом запуске автоматически загрузится структура БД и демо-данные из `./init/dump.sql`.

---

### ▶️ Локальный запуск (без Docker)

```bash
# 1. Установи зависимости
go mod download

# 2. Создай и настрой .env
cp .env.example .env
nano .env  # укажи свои настройки БД

# 3. Запусти миграции (если есть) или загрузи дамп вручную
psql -U postgres -f ./init/dump.sql

# 4. Запусти сервер
go run main.go
```

---

## 📁 Структура проекта

```
sitex/
├── cmd/               # точка входа (если вынесено)
├── internal/          # бизнес-логика: auth, user, pages
├── pkg/               # переиспользуемые компоненты: logger, middleware, database
├── public/            # статические файлы: CSS, JS, изображения
├── templates/         # шаблоны (если используешь templ/html)
├── init/              # SQL-дампы для инициализации БД
├── migrations/        # (опционально) миграции с помощью goose/migrate
├── docker-compose.yml
├── .env.example
└── main.go
```

---

## 🛠️ Технологии

| Категория     | Технология                          |
| ------------- | ----------------------------------- |
| Язык          | Go 1.21+                            |
| Веб-фреймворк | [Fiber](https://gofiber.io)         |
| Шаблонизация  | [templ](https://templ.guide)        |
| База данных   | PostgreSQL 16                       |
| Логирование   | zerolog                             |
| Контейнеры    | Docker, Docker Compose              |
| Сессии        | PostgreSQL-backed sessions          |
| UI            | Tailwind CSS, Alpine.js (если есть) |

---

## 🧩 Как работает авторизация

- Используется middleware сессий через `github.com/gofiber/session`
- Данные сессий хранятся в PostgreSQL (`sessions` таблица)
- Пользователь идентифицируется по email, сохранённому в `c.Locals("email")`
- Есть middleware для проверки аутентификации и роли

---

## 📖 Примеры API (если есть)

```bash
# Получить список пользователей (требует авторизации)
curl -H "Cookie: session_id=..." http://localhost:3000/api/users

# Создать задачу
curl -X POST http://localhost:3000/api/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "Новая задача", "assignee_id": 1}'
```

---

## 🤝 Как внести вклад

Мы рады любым улучшениям! Вот как это сделать:

1. 🍴 Форкните репозиторий
2. 🌿 Создайте ветку: `git checkout -b feature/your-awesome-feature`
3. 💾 Закоммитьте изменения: `git commit -m "Add some feature"`
4. ⬆️ Запушьте: `git push origin feature/your-awesome-feature`
5. 💌 Откройте Pull Request

---

## ❓ Частые вопросы / Troubleshooting

### ❗ При `git push` запрашивает логин/пароль

Убедись, что используешь SSH:

```bash
git remote set-url origin git@github.com:PuchkovaAS/sitex.git
```

### ❗ Ошибка подключения к БД

Проверь `.env` и убедись, что PostgreSQL запущен и доступен на `localhost:5432`.

### ❗ Не применяется дамп при запуске

Удали volume и пересоздай контейнеры:

```bash
docker-compose down -v
docker-compose up --build
```

---

## 📜 Лицензия

Этот проект распространяется под лицензией **MIT**.  
См. файл [LICENSE](LICENSE) для подробностей.

---

## 👤 Автор

**Анастасия Пучкова**  
📧 puchkovaac@rambler.ru
🐙 [GitHub](https://github.com/PuchkovaAS)

---

## 🎁 Благодарности

- Команде Fiber за отличный фреймворк 🚀
- Сообществу Go за вдохновение и поддержку 💙
- Всем, кто делает open-source лучше каждый день 🌍
