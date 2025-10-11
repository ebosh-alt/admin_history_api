# Admin History API

HTTP‑сервис для работы с анкетами, пользователями и медиафайлами проекта “Истории”. Ниже приведено описание доступных REST‑эндпоинтов. Базовый префикс для всех вызовов: `/api`.

- Базовый формат запросов/ответов — JSON (`application/json`), если не указано multipart.
- Время и даты передаются в UTC. Все фильтры по датам принимают либо unix‑timestamp (секунды), либо строку `YYYY-MM-DD` / `YYYY-MM-DDTHH:MM:SS`.
- Swagger: `/swagger/index.html`

## Пользователи (`/api/users`)

| Метод | Путь           | Описание | Важные параметры |
|-------|----------------|----------|------------------|
| GET   | `/users/{id}`  | Получить пользователя по ID. | `id` – путь. |
| GET   | `/users`       | Список пользователей с фильтрами. | Query: `page`, `limit`, `status`, `accepted_offer`, `promocode`, `age_from`, `age_to`, `gender`, `map_binding`, `date_from`, `date_to`. |
| POST  | `/users/update`| Обновить пользователя. | JSON‑тело `UpdateUserRequest`. |

## Анкеты (`/api/questionnaires`)

| Метод | Путь                | Описание | Параметры |
|-------|---------------------|----------|-----------|
| GET   | `/questionnaires/{id}` | Получить анкету по ID. | `id`. |
| GET   | `/questionnaires`   | Список анкет с фильтрами. | `page`, `limit`, `payment`, `status`, `user_id`, `date_from`, `date_to`. |
| POST  | `/questionnaires/update` | Обновить анкету. | JSON `UpdateQuestionnaireRequest`. |
| POST  | `/questionnaires/media`  | Загрузить и сохранить медиа набора (демо/финал). | multipart/form-data или JSON, см. ниже. |

### `/questionnaires/media` (multipart)

Используйте `multipart/form-data`, чтобы сразу загрузить файлы и передать сцену/тип:

| Поле | Тип | Обязательность | Назначение |
|------|-----|----------------|------------|
| `questionnaire_id` | formData int | Да | ID анкеты. |
| `user_id` | formData int | Да | Telegram chat_id пользователя, которому отправляются демо. |
| `demo_photos[]` | файлы | Нет | Демо‑фото; будет сохранено как `scene=demo`. |
| `final_photos[]` | файлы | Нет | Финальные фото; сцены задаются либо через `final_photo_scene[]`, либо в JSON (`payload`). |
| `demo_video` | файл | Нет | Видео с водяным знаком. |
| `generated_video` | файл | Нет | Финальное видео (type `generated`). |
| `delivery_photo` | файл | Нет | Отдельное фото для отправки пользователю (scene/type `delivery`). |
| `demo_photo_path` | string[] | Нет | Пути к уже существующим демо‑фото (если файлы загружены заранее). |
| `final_photo_path` | string[] | Нет | Пути к готовым финальным фото. |
| `delivery_photo_path` | string | Нет | Путь к готовому delivery‑фото. |
| `demo_video_path` | string | Нет | Путь к демо‑видео. |
| `generated_video_path` | string | Нет | Путь к финальному видео. |
| `final_photo_scene[]` | string[] | Нет | Сцены для файлов `final_photos[]` и/или площадок без `scene` в payload. Порядок соответствует порядку файлов. |
| `payload` | string | Нет | JSON `SubmitQuestionnaireMediaRequest`, позволяет передать структуры `demo_photos`, `final_photos` и др. (например, сцены по индексам). |

Поведение:

- Все переданные файлы сохраняются через `storage.FS` (`photos` → `data/photos`, видео → `data/videos`), в ответе и БД фиксируются относительные пути.
- Типы медиа нормализуются: для фото допустимы только `original`, `generated`, `send`, `demo`; для видео — `send`, `demo`. Если приходит любое другое значение, оно автоматически заменяется на `send` (для фото — `original` при загрузке через `/photos/upload`).
- Демо‑контент добавляется в Telegram (если бот настроен) и записывается в БД с типом `demo`.
- `final_photo_scene[]` применяется к загруженным финальным файлам. Если сцен больше, чем файлов — лишние игнорируются; если сцен меньше — “result”.
- Можно комбинировать файлы и заранее сохранённые пути (`*_path`).
- Пример cURL:

```bash
curl -X POST http://localhost:3000/api/questionnaires/media \
  -F questionnaire_id=123 \
  -F user_id=456 \
  -F demo_photos[]=@demo1.jpg \
  -F final_photos[]=@final1.jpg \
  -F final_photos[]=@final2.jpg \
  -F final_photo_scene[]=forest \
  -F final_photo_scene[]=castle \
  -F generated_video=@result.mp4
```

JSON‑вариант (`Content-Type: application/json`) также поддерживается: передавайте `SubmitQuestionnaireMediaRequest`, где `final_photos` содержит `scene` и `type_photo`, а пути уже должны существовать.

## Фото (`/api/photos`)

| Метод | Путь            | Описание | Параметры |
|-------|-----------------|----------|-----------|
| GET   | `/photos`       | Получить фото анкеты. | `questionnaire_id` (обязательно), `type` (например, `demo`, `result`, `delivery`, `generated`, `all`). |
| POST  | `/photos/upload`| Загрузить отдельное фото. | multipart: `questionnaire_id`, `scene`, `type`, `file`. |

## Видео (`/api/videos`)

| Метод | Путь            | Описание | Параметры |
|-------|-----------------|----------|-----------|
| GET   | `/videos`       | Получить видео анкеты. | `questionnaire_id`, `type`. |
| POST  | `/videos/upload`| Загрузить отдельный ролик. | multipart: `questionnaire_id`, `type`, `file`. |

## Промокоды (`/api/promo-codes`)

| Метод | Путь                   | Описание | Параметры |
|-------|------------------------|----------|-----------|
| GET   | `/promo-codes/{id}`    | Получить промокод по ID. | `id`. |
| GET   | `/promo-codes`         | Список промокодов. | `page`, `limit`, `status`. |
| POST  | `/promo-codes`         | Создать промокод. | JSON `CreatePromoCodeRequest`. |
| POST  | `/promo-codes/update`  | Обновить промокод. | JSON `UpdatePromoCodeRequest` (ID обязателен). |

## Отзывы (`/api/reviews`)

| Метод | Путь           | Описание | Параметры |
|-------|----------------|----------|-----------|
| GET   | `/reviews/{id}`| Получить отзыв по ID. | `id`. |
| GET   | `/reviews`     | Список отзывов. | `page`, `limit`, `user_id`, `date_from`, `date_to`. |

## Дополнительно

- Swagger‑описание всегда актуально по `/swagger/index.html`.
- Для интеграции через gRPC используйте сгенерированные protobuf‑стабы из `pkg/proto/gen/go`.

Все изменения по API старайтесь сопровождать обновлением этой таблицы и swagger‑комментариев в хэндлерах.
