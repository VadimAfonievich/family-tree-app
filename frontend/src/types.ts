export type Gender = 'male' | 'female' | 'other';
export type RelationType = 'parent' | 'spouse';

export interface Tree {
  id: string;
  owner_id: string;
  title: string;
  created_at: string;
}

export interface Person {
  id: string;
  tree_id: string;
  first_name: string;
  last_name: string;
  gender: Gender;
  birth_date: string | null;
  death_date: string | null;
  photo_url: string;
  created_at: string;
}

export interface Relation {
  id: string;
  tree_id: string;
  person1_id: string;
  person2_id: string;
  relation_type: RelationType;
  created_at: string;
}

export interface FullTree {
  tree: Tree;
  persons: Person[];
  relations: Relation[];
}
