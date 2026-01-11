-- +goose up
CREATE TABLE posts (
    id uuid PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

);