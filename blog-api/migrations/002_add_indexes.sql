-- Индексы для таблицы users
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_created ON users (created);

-- Индексы для таблицы posts
CREATE INDEX IF NOT EXISTS idx_posts_author_id ON posts (author_id);
CREATE INDEX IF NOT EXISTS idx_posts_created ON posts (created DESC);
CREATE INDEX IF NOT EXISTS idx_posts_updated ON posts (updated DESC);
CREATE INDEX IF NOT EXISTS idx_posts_title ON posts USING GIN (to_tsvector('english', title));
CREATE INDEX IF NOT EXISTS idx_posts_content ON posts USING GIN (to_tsvector('english', content));

-- Составной индекс для поиска постов по автору и дате
CREATE INDEX IF NOT EXISTS idx_posts_author_created ON posts (author_id, created DESC);

-- Индексы для таблицы comments
CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments (post_id);
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments (user_id);
CREATE INDEX IF NOT EXISTS idx_comments_created ON comments (created DESC);

-- Составной индекс для быстрого получения комментариев к посту в хронологическом порядке
CREATE INDEX IF NOT EXISTS idx_comments_post_created ON comments (post_id, created ASC);

-- +migrate Down
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_created;

DROP INDEX IF EXISTS idx_postsauthor_id;
DROP INDEX IF EXISTS idx_posts_created;
DROP INDEX IF EXISTS idx_posts_updated;
DROP INDEX IF EXISTS idx_posts_title;
DROP INDEX IF EXISTS idx_posts_content;
DROP INDEX IF EXISTS idx_posts_author_created;

DROP INDEX IF EXISTS idx_comments_post_id;
DROP INDEX IF EXISTS idx_comments_user_id;
DROP INDEX IF EXISTS idx_comments_created;
DROP INDEX IF EXISTS idx_comments_post_created;