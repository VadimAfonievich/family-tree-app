import { ArrowLeft, GitBranch, Plus, RefreshCw, Trash2, UserPlus, Users } from 'lucide-react';
import { FormEvent, PointerEvent, useCallback, useEffect, useMemo, useState } from 'react';
import {
  authTelegram,
  createPerson,
  createRelation,
  createTree,
  deletePerson,
  getStoredToken,
  getTree,
  listTrees,
  updatePerson,
} from './api';
import { configureBackButton, configureMainButton, getInitData, initTelegram } from './telegram';
import type { FullTree, Gender, Person, Relation, Tree } from './types';

type View = 'trees' | 'tree';
type ModalMode = 'person' | 'edit' | 'parent' | 'child' | 'spouse' | null;

const emptyPerson = {
  first_name: '',
  last_name: '',
  gender: 'other' as Gender,
  birth_date: '',
  death_date: '',
  photo_url: '',
};

type PersonForm = typeof emptyPerson;

export default function App() {
  const [view, setView] = useState<View>('trees');
  const [trees, setTrees] = useState<Tree[]>([]);
  const [activeTree, setActiveTree] = useState<FullTree | null>(null);
  const [selected, setSelected] = useState<Person | null>(null);
  const [treeTitle, setTreeTitle] = useState('');
  const [form, setForm] = useState<PersonForm>(emptyPerson);
  const [modal, setModal] = useState<ModalMode>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const loadTrees = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      setTrees(await listTrees());
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось загрузить деревья');
    } finally {
      setLoading(false);
    }
  }, []);

  const loadTree = useCallback(async (id: string) => {
    setLoading(true);
    setError('');
    try {
      const data = await getTree(id);
      setActiveTree(data);
      setView('tree');
      setSelected(data.persons[0] ?? null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось загрузить дерево');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    initTelegram();
    async function boot() {
      try {
        if (!getStoredToken()) {
          await authTelegram(getInitData());
        }
        await loadTrees();
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Ошибка авторизации');
      } finally {
        setLoading(false);
      }
    }
    void boot();
  }, [loadTrees]);

  const goBack = useCallback(() => {
    if (view === 'tree') {
      setView('trees');
      setActiveTree(null);
      setSelected(null);
      void loadTrees();
    }
  }, [loadTrees, view]);

  useEffect(() => {
    configureBackButton(view === 'tree', goBack);
  }, [goBack, view]);

  useEffect(() => {
    const onMain = () => {
      if (view === 'trees') {
        void submitTree();
      } else {
        openModal('person');
      }
    };
    configureMainButton(view === 'trees' ? 'Создать дерево' : 'Добавить человека', true, onMain);
    return () => configureMainButton('', false, onMain);
  });

  async function submitTree(event?: FormEvent) {
    event?.preventDefault();
    if (!treeTitle.trim()) return;
    const tree = await createTree(treeTitle.trim());
    setTreeTitle('');
    setTrees((current) => [tree, ...current]);
    await loadTree(tree.id);
  }

  function openModal(mode: ModalMode, person = selected) {
    setModal(mode);
    if (mode === 'edit' && person) {
      setForm({
        first_name: person.first_name,
        last_name: person.last_name,
        gender: person.gender,
        birth_date: person.birth_date?.slice(0, 10) ?? '',
        death_date: person.death_date?.slice(0, 10) ?? '',
        photo_url: person.photo_url,
      });
    } else {
      setForm(emptyPerson);
    }
  }

  async function submitPerson(event: FormEvent) {
    event.preventDefault();
    if (!activeTree || !modal || !form.first_name.trim()) return;

    const input = {
      tree_id: activeTree.tree.id,
      first_name: form.first_name.trim(),
      last_name: form.last_name.trim(),
      gender: form.gender,
      birth_date: form.birth_date || undefined,
      death_date: form.death_date || undefined,
      photo_url: form.photo_url.trim(),
    };

    if (modal === 'edit' && selected) {
      await updatePerson(selected.id, input);
    } else {
      const person = await createPerson(input);
      if (selected && modal === 'parent') {
        await createRelation(activeTree.tree.id, person.id, selected.id, 'parent');
      }
      if (selected && modal === 'child') {
        await createRelation(activeTree.tree.id, selected.id, person.id, 'parent');
      }
      if (selected && modal === 'spouse') {
        await createRelation(activeTree.tree.id, selected.id, person.id, 'spouse');
      }
      setSelected(person);
    }

    setModal(null);
    await loadTree(activeTree.tree.id);
  }

  async function removeSelected() {
    if (!selected || !activeTree) return;
    await deletePerson(selected.id);
    setSelected(null);
    await loadTree(activeTree.tree.id);
  }

  const content = view === 'trees' ? (
    <TreesView trees={trees} title={treeTitle} setTitle={setTreeTitle} onCreate={submitTree} onOpen={loadTree} loading={loading} />
  ) : activeTree ? (
    <TreeView
      data={activeTree}
      selected={selected}
      onSelect={setSelected}
      onBack={goBack}
      onRefresh={() => loadTree(activeTree.tree.id)}
      onAdd={() => openModal('person')}
      onAddParent={() => openModal('parent')}
      onAddChild={() => openModal('child')}
      onAddSpouse={() => openModal('spouse')}
      onEdit={() => openModal('edit')}
      onDelete={removeSelected}
    />
  ) : null;

  return (
    <main className="safe-page mx-auto flex max-w-5xl flex-col gap-4">
      {error && <div className="rounded-lg bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>}
      {content}
      {modal && (
        <PersonModal
          mode={modal}
          form={form}
          setForm={setForm}
          onClose={() => setModal(null)}
          onSubmit={submitPerson}
        />
      )}
    </main>
  );
}

function TreesView({
  trees,
  title,
  setTitle,
  onCreate,
  onOpen,
  loading,
}: {
  trees: Tree[];
  title: string;
  setTitle: (value: string) => void;
  onCreate: (event: FormEvent) => void;
  onOpen: (id: string) => void;
  loading: boolean;
}) {
  return (
    <>
      <header className="flex items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold">Семейные деревья</h1>
          <p className="text-sm text-hint">Создай дерево и добавь родственников.</p>
        </div>
        <div className="grid h-12 w-12 place-items-center rounded-lg bg-panel shadow-soft">
          <GitBranch className="text-accent" />
        </div>
      </header>

      <form onSubmit={onCreate} className="flex gap-2">
        <input className="field" value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Название дерева" />
        <button className="btn-primary shrink-0" type="submit" aria-label="Создать дерево">
          <Plus size={18} />
        </button>
      </form>

      <section className="grid gap-3">
        {loading && <div className="rounded-lg bg-panel p-4 text-sm text-hint">Загрузка...</div>}
        {!loading && trees.length === 0 && <div className="rounded-lg bg-panel p-4 text-sm text-hint">Пока нет ни одного дерева.</div>}
        {trees.map((tree) => (
          <button key={tree.id} className="rounded-lg bg-panel p-4 text-left shadow-sm transition active:scale-[0.99]" onClick={() => onOpen(tree.id)}>
            <div className="font-semibold">{tree.title}</div>
            <div className="mt-1 text-xs text-hint">{new Date(tree.created_at).toLocaleDateString()}</div>
          </button>
        ))}
      </section>
    </>
  );
}

function TreeView({
  data,
  selected,
  onSelect,
  onBack,
  onRefresh,
  onAdd,
  onAddParent,
  onAddChild,
  onAddSpouse,
  onEdit,
  onDelete,
}: {
  data: FullTree;
  selected: Person | null;
  onSelect: (person: Person) => void;
  onBack: () => void;
  onRefresh: () => void;
  onAdd: () => void;
  onAddParent: () => void;
  onAddChild: () => void;
  onAddSpouse: () => void;
  onEdit: () => void;
  onDelete: () => void;
}) {
  return (
    <>
      <header className="flex items-center justify-between gap-2">
        <button className="btn-soft w-11 px-0" onClick={onBack} aria-label="Назад">
          <ArrowLeft size={18} />
        </button>
        <div className="min-w-0 flex-1">
          <h1 className="truncate text-xl font-bold">{data.tree.title}</h1>
          <p className="text-sm text-hint">{data.persons.length} чел.</p>
        </div>
        <button className="btn-soft w-11 px-0" onClick={onRefresh} aria-label="Обновить">
          <RefreshCw size={18} />
        </button>
        <button className="btn-primary w-11 px-0" onClick={onAdd} aria-label="Добавить человека">
          <UserPlus size={18} />
        </button>
      </header>

      <FamilyGraph persons={data.persons} relations={data.relations} selected={selected} onSelect={onSelect} />
      <PersonCard
        person={selected}
        onAddParent={onAddParent}
        onAddChild={onAddChild}
        onAddSpouse={onAddSpouse}
        onEdit={onEdit}
        onDelete={onDelete}
      />
    </>
  );
}

function FamilyGraph({
  persons,
  relations,
  selected,
  onSelect,
}: {
  persons: Person[];
  relations: Relation[];
  selected: Person | null;
  onSelect: (person: Person) => void;
}) {
  const [pan, setPan] = useState({ x: 0, y: 0 });
  const [scale, setScale] = useState(1);
  const [drag, setDrag] = useState<{ x: number; y: number } | null>(null);

  const layout = useMemo(() => buildLayout(persons, relations), [persons, relations]);

  function onPointerDown(event: PointerEvent<SVGSVGElement>) {
    setDrag({ x: event.clientX - pan.x, y: event.clientY - pan.y });
  }

  function onPointerMove(event: PointerEvent<SVGSVGElement>) {
    if (!drag) return;
    setPan({ x: event.clientX - drag.x, y: event.clientY - drag.y });
  }

  function zoom(delta: number) {
    setScale((value) => Math.min(1.8, Math.max(0.55, value + delta)));
  }

  if (persons.length === 0) {
    return (
      <section className="grid min-h-[380px] place-items-center rounded-lg bg-panel p-6 text-center text-hint">
        <div>
          <Users className="mx-auto mb-2 text-accent" />
          Add the first person to start the tree.
        </div>
      </section>
    );
  }

  return (
    <section className="relative h-[52vh] min-h-[380px] overflow-hidden rounded-lg bg-panel shadow-sm">
      <div className="absolute right-3 top-3 z-10 flex gap-2">
        <button className="btn-soft h-9 w-9 px-0" onClick={() => zoom(-0.15)} aria-label="Уменьшить">-</button>
        <button className="btn-soft h-9 w-9 px-0" onClick={() => zoom(0.15)} aria-label="Увеличить">+</button>
      </div>
      <svg
        className="h-full w-full touch-none"
        onPointerDown={onPointerDown}
        onPointerMove={onPointerMove}
        onPointerUp={() => setDrag(null)}
        onPointerCancel={() => setDrag(null)}
      >
        <g transform={`translate(${pan.x + 40} ${pan.y + 40}) scale(${scale})`}>
          {layout.lines.map((line) => (
            <line key={line.id} x1={line.x1} y1={line.y1} x2={line.x2} y2={line.y2} stroke={line.type === 'spouse' ? '#ef8f2f' : '#8aa0b5'} strokeWidth="2.5" strokeLinecap="round" />
          ))}
          {layout.nodes.map((node) => (
            <g key={node.person.id} transform={`translate(${node.x} ${node.y})`} onClick={(event) => { event.stopPropagation(); onSelect(node.person); }}>
              <rect width="142" height="64" rx="8" fill={selected?.id === node.person.id ? 'var(--tg-accent)' : '#ffffff'} stroke="#d8dee8" />
              <text x="14" y="27" fontSize="14" fontWeight="700" fill={selected?.id === node.person.id ? '#ffffff' : '#161b22'}>
                {node.person.first_name}
              </text>
              <text x="14" y="47" fontSize="12" fill={selected?.id === node.person.id ? '#eaf3ff' : '#6b7280'}>
                {
                 node.person.last_name ||
                  (node.person.gender === 'male'
                    ? 'Муж.'
                    : node.person.gender === 'female'
                    ? 'Жен.'
                    : 'Друг.')
                }
              </text>
            </g>
          ))}
        </g>
      </svg>
    </section>
  );
}

function buildLayout(persons: Person[], relations: Relation[]) {
  const children = new Map<string, string[]>();
  const hasParent = new Set<string>();
  relations.filter((r) => r.relation_type === 'parent').forEach((relation) => {
    children.set(relation.person1_id, [...(children.get(relation.person1_id) ?? []), relation.person2_id]);
    hasParent.add(relation.person2_id);
  });

  const levels = new Map<string, number>();
  const byID = new Map(persons.map((person) => [person.id, person]));
  const roots = persons.filter((person) => !hasParent.has(person.id));

  function visit(id: string, level: number) {
    levels.set(id, Math.max(levels.get(id) ?? 0, level));
    for (const child of children.get(id) ?? []) {
      visit(child, level + 1);
    }
  }

  (roots.length ? roots : persons).forEach((person) => visit(person.id, 0));

  relations
    .filter((r) => r.relation_type === 'spouse')
    .forEach((relation) => {
      const level1 = levels.get(relation.person1_id);
      const level2 = levels.get(relation.person2_id);

      if (level1 !== undefined && level2 === undefined) {
        levels.set(relation.person2_id, level1);
      } else if (level2 !== undefined && level1 === undefined) {
        levels.set(relation.person1_id, level2);
      }
    });

  const grouped = new Map<number, Person[]>();
  persons.forEach((person) => {
    const level = levels.get(person.id) ?? 0;
    grouped.set(level, [...(grouped.get(level) ?? []), person]);
  });

  const spouseMap = new Map<string, string>();

  relations
    .filter((r) => r.relation_type === 'spouse')
    .forEach((r) => {
      spouseMap.set(r.person1_id, r.person2_id);
      spouseMap.set(r.person2_id, r.person1_id);
    });

  const positioned = new Set<string>();

  const nodes: {
    person: Person;
    x: number;
    y: number;
  }[] = [];

  grouped.forEach((group, level) => {
    let x = 0;

    for (const person of group) {
      if (positioned.has(person.id)) continue;

      const spouseId = spouseMap.get(person.id);
      const spouse = spouseId
        ? byID.get(spouseId)
        : null;

      nodes.push({
        person,
        x,
        y: level * 130,
      });

      positioned.add(person.id);

      if (spouse && !positioned.has(spouse.id)) {
        nodes.push({
          person: spouse,
          x: x + 176,
          y: level * 130,
        });

        positioned.add(spouse.id);

        x += 352;
      } else {
        x += 176;
      }
    }
  });

  const nodeMap = new Map(nodes.map((node) => [node.person.id, node]));
  const lines = relations.flatMap((relation) => {
    const a = nodeMap.get(relation.person1_id);
    const b = nodeMap.get(relation.person2_id);
    if (!a || !b) return [];
    return [{
      id: relation.id,
      type: relation.relation_type,
      x1: a.x + 71,
      y1: a.y + (relation.relation_type === 'parent' ? 64 : 32),
      x2: b.x + 71,
      y2: b.y + (relation.relation_type === 'parent' ? 0 : 32),
    }];
  });

  return { nodes, lines };
}

function PersonCard({
  person,
  onAddParent,
  onAddChild,
  onAddSpouse,
  onEdit,
  onDelete,
}: {
  person: Person | null;
  onAddParent: () => void;
  onAddChild: () => void;
  onAddSpouse: () => void;
  onEdit: () => void;
  onDelete: () => void;
}) {
  if (!person) {
    return <section className="rounded-lg bg-panel p-4 text-sm text-hint">Выберите человека для просмотра информации.</section>;
  }

  return (
    <section className="rounded-lg bg-panel p-4 shadow-sm">
      <div className="flex gap-3">
        <div className="grid h-16 w-16 shrink-0 place-items-center overflow-hidden rounded-lg bg-slate-100 text-lg font-bold text-accent">
          {person.photo_url ? <img className="h-full w-full object-cover" src={person.photo_url} alt="" /> : person.first_name[0]}
        </div>
        <div className="min-w-0 flex-1">
          <h2 className="truncate text-lg font-bold">{person.first_name} {person.last_name}</h2>
          <p className="text-sm text-hint">{person.gender}</p>
          <p className="mt-1 text-sm">{formatLife(person)}</p>
        </div>
      </div>
      <div className="mt-4 grid grid-cols-2 gap-2">
        <button className="btn-soft" onClick={onAddParent}>Добавить родителя</button>
        <button className="btn-soft" onClick={onAddChild}>Добавить ребенка</button>
        <button className="btn-soft" onClick={onAddSpouse}>Добавить супруга</button>
        <button className="btn-soft" onClick={onEdit}>Изменить</button>
        <button className="btn-soft col-span-2 text-red-600" onClick={onDelete}>
          <Trash2 size={16} /> Удалить
        </button>
      </div>
    </section>
  );
}

function PersonModal({
  mode,
  form,
  setForm,
  onClose,
  onSubmit,
}: {
  mode: ModalMode;
  form: PersonForm;
  setForm: (value: PersonForm) => void;
  onClose: () => void;
  onSubmit: (event: FormEvent) => void;
}) {
  const title = mode === 'edit' ? 'Редактировать человека' : mode === 'parent' ? 'Добавить родителя' : mode === 'child' ? 'Добавить ребёнка' : mode === 'spouse' ? 'Добавить супруга' : 'Добавить человека';

  return (
    <div className="fixed inset-0 z-30 grid place-items-end bg-black/30 p-3">
      <form onSubmit={onSubmit} className="w-full max-w-lg rounded-lg bg-panel p-4 shadow-soft">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-bold">{title}</h2>
          <button type="button" className="btn-soft h-9 px-3" onClick={onClose}>Закрыть</button>
        </div>
        <div className="grid gap-3">
          <input className="field" required value={form.first_name} onChange={(e) => setForm({ ...form, first_name: e.target.value })} placeholder="Имя" />
          <input className="field" value={form.last_name} onChange={(e) => setForm({ ...form, last_name: e.target.value })} placeholder="Фамилия" />
          <select className="field" value={form.gender} onChange={(e) => setForm({ ...form, gender: e.target.value as Gender })}>
            <option value="female">Жен</option>
            <option value="male">Муж</option>
            <option value="other">Другое</option>
          </select>
          <input className="field" type="date" value={form.birth_date} onChange={(e) => setForm({ ...form, birth_date: e.target.value })} />
          <input className="field" type="date" value={form.death_date} onChange={(e) => setForm({ ...form, death_date: e.target.value })} />
          <input className="field" value={form.photo_url} onChange={(e) => setForm({ ...form, photo_url: e.target.value })} placeholder="Фото ссылка" />
          <button className="btn-primary" type="submit">Сохранить</button>
        </div>
      </form>
    </div>
  );
}

function formatLife(person: Person) {
  const birth = person.birth_date ? new Date(person.birth_date).getFullYear() : '';
  const death = person.death_date ? new Date(person.death_date).getFullYear() : '';
  if (!birth && !death) return 'Даты неизвестны';
  return `${birth || '?'} - ${death || ''}`;
}
