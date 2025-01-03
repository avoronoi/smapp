CREATE TABLE posts (
    id BINARY(16) PRIMARY KEY,
    author_id BINARY(16) NOT NULL,
    body TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    -- Index to speed up ORDER BY when fetching paginated feed
    INDEX created_at_id_index (created_at DESC, id)
);

CREATE TABLE likes (
    id BINARY(16) PRIMARY KEY,
    entity_type ENUM('posts', 'comments') NOT NULL,
    entity_id BINARY(16) NOT NULL,
    author_id BINARY(16) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY author_entity_unique (entity_type, entity_id, author_id)
);

CREATE TABLE comments (
    id BINARY(16) PRIMARY KEY,
    post_id BINARY(16) NOT NULL,
    author_id BINARY(16) NOT NULL,
    body TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    -- Index to speed up ORDER BY when fetching paginated comments for a post
    INDEX created_at_id_index (created_at DESC, id),
    FOREIGN KEY (post_id) REFERENCES posts(id)
);

CREATE TABLE likes_count (
    id BINARY(16) PRIMARY KEY,
    entity_type ENUM('posts', 'comments') NOT NULL,
    entity_id BINARY(16) NOT NULL,
    UNIQUE KEY entity_unique (entity_type, entity_id),
    count INT UNSIGNED NOT NULL DEFAULT 0
);

CREATE TABLE comments_count (
    id BINARY(16) PRIMARY KEY,
    post_id BINARY(16) NOT NULL,
    count INT UNSIGNED NOT NULL DEFAULT 0,
    UNIQUE KEY post_id_unique (post_id),
    FOREIGN KEY (post_id) REFERENCES posts(id)
);

CREATE TABLE images (
    id BINARY(16) PRIMARY KEY,
    post_id BINARY(16) NOT NULL,
    position INT UNSIGNED NOT NULL,
    s3_bucket VARCHAR(63) NOT NULL,
    s3_key VARCHAR(1024) NOT NULL
);