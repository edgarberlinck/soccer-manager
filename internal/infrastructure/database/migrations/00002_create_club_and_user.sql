-- +goose Up
CREATE table users (
  id UUID PRIMARY KEY,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  active boolean
);

CREATE table clubs (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id),
  name TEXT Not Null Unique
);

-- +goose Down
DROP Table clubs;
DROP Table users;