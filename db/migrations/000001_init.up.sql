CREATE TABLE IF NOT EXISTS users(
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    password VARCHAR(72) NOT NULL,
    image_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    friend_count INTEGER DEFAULT 0 NOT NULL
);

CREATE TABLE IF NOT EXISTS user_credentials(
    id SERIAL PRIMARY KEY,
    credential_type VARCHAR(10) NOT NULL,
    credential_value VARCHAR(255) UNIQUE NOT NULL,
    user_id INTEGER REFERENCES users(id) NOT NULL
);


CREATE TABLE IF NOT EXISTS relationships(
    id SERIAL PRIMARY KEY,
    user_first_id INTEGER REFERENCES users(id) NOT NULL,
    user_second_id INTEGER REFERENCES users(id) NOT NULL
    -- CONSTRAINT check_not_self_relationship CHECK (user_first_id <> user_second_id)
);

CREATE UNIQUE INDEX unique_relationship 
ON relationships (LEAST(user_first_id, user_second_id), GREATEST(user_first_id, user_second_id));

CREATE TABLE IF NOT EXISTS posts(
    id SERIAL PRIMARY KEY,
    html TEXT NOT NULL,
    user_id INTEGER REFERENCES users(id) NOT NULL,
    tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS comments(
    id SERIAL PRIMARY KEY,
    comment TEXT NOT NULL,
    post_id INTEGER REFERENCES posts(id) NOT NULL,
    user_id INTEGER REFERENCES users(id) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP not NULL
);
