import { useState, useEffect } from 'react';
import type { TODO, TODOStatus } from '../types';
import { api } from '../services/api';
import { Layout } from '../components/Layout';
import { TODOCard } from '../components/TODOCard';
import { Loader2 } from 'lucide-react';
import {
  DndContext,
  DragOverlay,
  closestCorners,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core';
import type { DragStartEvent, DragEndEvent } from '@dnd-kit/core';
import {
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
  useSortable,
} from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';

interface KanbanColumnProps {
  id: TODOStatus;
  title: string;
  todos: TODO[];
}

function KanbanColumn({ id, title, todos }: KanbanColumnProps) {
  const { setNodeRef, transform, transition, isDragging } = useSortable({
    id,
    data: { type: 'column' },
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  return (
    <div ref={setNodeRef} style={style} className="flex-1 min-w-[280px] max-w-[350px]">
      <div className="bg-gray-100 dark:bg-gray-800 rounded-lg p-4 h-full">
        <div className="flex items-center justify-between mb-4">
          <h3 className="font-semibold text-gray-900 dark:text-white">{title}</h3>
          <span className="bg-gray-200 dark:bg-gray-700 text-gray-600 dark:text-gray-300 text-sm px-2 py-0.5 rounded-full">
            {todos.length}
          </span>
        </div>
        <SortableContext items={todos.map(t => t.id)} strategy={verticalListSortingStrategy}>
          <div className="space-y-3 min-h-[200px]">
            {todos.map(todo => (
              <SortableTODO key={todo.id} todo={todo} />
            ))}
          </div>
        </SortableContext>
      </div>
    </div>
  );
}

function SortableTODO({ todo }: { todo: TODO }) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: todo.id,
    data: { type: 'todo', todo },
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  if (isDragging) {
    return (
      <div ref={setNodeRef} style={style} className="opacity-50">
        <TODOCard todo={todo} isDragging />
      </div>
    );
  }

  return (
    <div ref={setNodeRef} style={style} {...attributes} {...listeners}>
      <TODOCard todo={todo} />
    </div>
  );
}

const columns: { id: TODOStatus; title: string }[] = [
  { id: 'open', title: 'Open' },
  { id: 'in_progress', title: 'In Progress' },
  { id: 'blocked', title: 'Blocked' },
  { id: 'resolved', title: 'Resolved' },
  { id: 'closed', title: 'Closed' },
];

export function Kanban() {
  const [todos, setTodos] = useState<TODO[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeId, setActiveId] = useState<string | null>(null);

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 8 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates })
  );

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const data = await api.getTODOs();
      setTodos(data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleDragStart = (event: DragStartEvent) => {
    setActiveId(event.active.id as string);
  };

  const handleDragEnd = async (event: DragEndEvent) => {
    const { active, over } = event;
    setActiveId(null);

    if (!over) return;

    const todoId = active.id as string;
    const overId = over.id as string;

    // Check if dropped on a column
    const targetColumn = columns.find(c => c.id === overId);
    if (targetColumn) {
      await handleStatusChange(todoId, targetColumn.id);
      return;
    }

    // Check if dropped on another todo
    const overTodo = todos.find(t => t.id === overId);
    if (overTodo) {
      await handleStatusChange(todoId, overTodo.status);
    }
  };

  const handleStatusChange = async (id: string, status: TODOStatus) => {
    try {
      await api.updateTODO(id, { status });
      setTodos(prev => prev.map(t => t.id === id ? { ...t, status } : t));
    } catch (err) {
      console.error('Failed to update status:', err);
    }
  };

  const getTodosByStatus = (status: TODOStatus) => todos.filter(t => t.status === status);

  const activeTodo = activeId ? todos.find(t => t.id === activeId) : null;

  return (
    <Layout>
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Kanban Board</h1>
          <p className="text-gray-600 dark:text-gray-400">Drag and drop to change status</p>
        </div>

        {loading ? (
          <div className="flex items-center justify-center min-h-[300px]">
            <Loader2 className="w-8 h-8 animate-spin text-blue-500" />
          </div>
        ) : (
          <DndContext
            sensors={sensors}
            collisionDetection={closestCorners}
            onDragStart={handleDragStart}
            onDragEnd={handleDragEnd}
          >
            <div className="flex gap-4 overflow-x-auto pb-4">
              {columns.map(column => (
                <KanbanColumn
                  key={column.id}
                  id={column.id}
                  title={column.title}
                  todos={getTodosByStatus(column.id)}
                />
              ))}
            </div>
            <DragOverlay>
              {activeTodo && <TODOCard todo={activeTodo} isDragging />}
            </DragOverlay>
          </DndContext>
        )}
      </div>
    </Layout>
  );
}
