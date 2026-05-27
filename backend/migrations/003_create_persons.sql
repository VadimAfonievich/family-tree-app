-- +goose Up
CREATE TABLE persons (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tree_id UUID NOT NULL REFERENCES trees(id) ON DELETE CASCADE,
    first_name TEXT NOT NULL,
    last_name TEXT,
    gender TEXT NOT NULL CHECK (gender IN ('male', 'female', 'other')),
    birth_date DATE,
    death_date DATE,
    photo_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_persons_tree_id ON persons(tree_id);

-- +goose Down
DROP TABLE IF EXISTS persons;
