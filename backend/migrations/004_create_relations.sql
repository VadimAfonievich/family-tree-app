-- +goose Up
CREATE TABLE relations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tree_id UUID NOT NULL REFERENCES trees(id) ON DELETE CASCADE,
    person1_id UUID NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    person2_id UUID NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    relation_type TEXT NOT NULL CHECK (relation_type IN ('parent', 'spouse')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_different_persons CHECK (person1_id != person2_id)
);

CREATE INDEX idx_relations_tree_id ON relations(tree_id);
CREATE INDEX idx_relations_person1 ON relations(person1_id);
CREATE INDEX idx_relations_person2 ON relations(person2_id);

-- +goose Down
DROP TABLE IF EXISTS relations;
