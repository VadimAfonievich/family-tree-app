import type { FullTree, Gender, Person, Relation, RelationType, Tree } from './types';

const API_URL = import.meta.env.VITE_API_URL ?? '';

export function getStoredToken() {
  return localStorage.getItem('family_tree_token');
}

export function setStoredToken(token: string) {
  localStorage.setItem('family_tree_token', token);
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getStoredToken();
  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error ?? `Request failed: ${res.status}`);
  }

  return res.json() as Promise<T>;
}

export async function authTelegram(initData: string) {
  const data = await request<{ token: string }>('/api/auth/telegram', {
    method: 'POST',
    body: JSON.stringify({ init_data: initData }),
  });
  setStoredToken(data.token);
  return data.token;
}

export function listTrees() {
  return request<Tree[]>('/api/trees');
}

export function createTree(title: string) {
  return request<Tree>('/api/trees', {
    method: 'POST',
    body: JSON.stringify({ title }),
  });
}

export function getTree(id: string) {
  return request<FullTree>(`/api/trees/${id}`);
}

export interface PersonInput {
  tree_id: string;
  first_name: string;
  last_name?: string;
  gender: Gender;
  birth_date?: string;
  death_date?: string;
  photo_url?: string;
}

export function createPerson(input: PersonInput) {
  return request<Person>('/api/persons', {
    method: 'POST',
    body: JSON.stringify(input),
  });
}

export function updatePerson(id: string, input: Omit<PersonInput, 'tree_id'>) {
  return request<Person>(`/api/persons/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(input),
  });
}

export function deletePerson(id: string) {
  return request<{ message: string }>(`/api/persons/${id}`, { method: 'DELETE' });
}

export function createRelation(treeId: string, person1Id: string, person2Id: string, relationType: RelationType) {
  return request<Relation>('/api/relations', {
    method: 'POST',
    body: JSON.stringify({
      tree_id: treeId,
      person1_id: person1Id,
      person2_id: person2Id,
      relation_type: relationType,
    }),
  });
}
