-- Миграция для создания начальной схемы базы данных блог-платформы

-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Таблица постов
CREATE TABLE IF NOT EXISTS posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    author_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT fk_posts_author
        FOREIGN KEY (author_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- Таблица комментариев
CREATE TABLE IF NOT EXISTS comments (
    id SERIAL PRIMARY KEY,
    content TEXT NOT NULL,
    post_id INTEGER NOT NULL,
    author_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT fk_comments_post
        FOREIGN KEY (post_id)
        REFERENCES posts(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_comments_author
        FOREIGN KEY (author_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- Индексы для оптимизации запросов

-- Индекс на posts.author_id для быстрого поиска постов пользователя
CREATE INDEX IF NOT EXISTS idx_posts_author ON posts(author_id);

-- Индекс на comments.post_id для быстрого поиска комментариев к посту
CREATE INDEX IF NOT EXISTS idx_comments_post ON comments(post_id);

-- Индекс на comments.author_id для быстрого поиска комментариев пользователя
CREATE INDEX IF NOT EXISTS idx_comments_author ON comments(author_id);

-- Индекс на posts.created_at для сортировки по дате
CREATE INDEX IF NOT EXISTS idx_posts_created ON posts(created_at DESC);

-- Дополнительный индекс на username для быстрого поиска (часто используемый запрос)
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- Дополнительный индекс на email для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
