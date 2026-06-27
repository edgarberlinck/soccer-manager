-- +goose Up
CREATE TABLE players (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    age INTEGER NOT NULL,

    pace SMALLINT NOT NULL,
    passing SMALLINT NOT NULL,
    shooting SMALLINT NOT NULL
);

-- +goose Down
DROP TABLE players;