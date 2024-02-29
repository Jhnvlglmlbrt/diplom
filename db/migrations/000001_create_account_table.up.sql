CREATE TABLE IF NOT EXISTS accounts (
    id serial PRIMARY KEY, 
    user_id UUID NOT NULL,
    notify_upfront INT DEFAULT 7,
    FOREIGN KEY (user_id) REFERENCES auth.users (id)
);