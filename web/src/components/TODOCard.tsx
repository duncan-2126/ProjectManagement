import type { TODO } from '../types';
import { FileText, User, Calendar, AlertCircle } from 'lucide-react';

interface TODOCardProps {
  todo: TODO;
  onClick?: () => void;
  isDragging?: boolean;
}

const priorityColors: Record<string, string> = {
  P0: 'bg-red-100 text-red-800 dark:bg-red-900/50 dark:text-red-300',
  P1: 'bg-orange-100 text-orange-800 dark:bg-orange-900/50 dark:text-orange-300',
  P2: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/50 dark:text-yellow-300',
  P3: 'bg-blue-100 text-blue-800 dark:bg-blue-900/50 dark:text-blue-300',
  P4: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300',
};

const statusColors: Record<string, string> = {
  open: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300',
  in_progress: 'bg-blue-100 text-blue-800 dark:bg-blue-900/50 dark:text-blue-300',
  blocked: 'bg-red-100 text-red-800 dark:bg-red-900/50 dark:text-red-300',
  resolved: 'bg-green-100 text-green-800 dark:bg-green-900/50 dark:text-green-300',
  wontfix: 'bg-purple-100 text-purple-800 dark:bg-purple-900/50 dark:text-purple-300',
  closed: 'bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400',
};

const typeIcons: Record<string, string> = {
  TODO: 'bg-blue-100 text-blue-800 dark:bg-blue-900/50 dark:text-blue-300',
  FIXME: 'bg-red-100 text-red-800 dark:bg-red-900/50 dark:text-red-300',
  HACK: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/50 dark:text-yellow-300',
  BUG: 'bg-red-100 text-red-800 dark:bg-red-900/50 dark:text-red-300',
  NOTE: 'bg-green-100 text-green-800 dark:bg-green-900/50 dark:text-green-300',
  XXX: 'bg-purple-100 text-purple-800 dark:bg-purple-900/50 dark:text-purple-300',
};

export function TODOCard({ todo, onClick, isDragging }: TODOCardProps) {
  const formatDate = (date: string | null) => {
    if (!date) return null;
    return new Date(date).toLocaleDateString();
  };

  return (
    <div
      onClick={onClick}
      className={`bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4 cursor-pointer transition-all hover:shadow-md ${
        isDragging ? 'shadow-lg ring-2 ring-blue-500' : ''
      }`}
    >
      <div className="flex items-start justify-between gap-2 mb-2">
        <span className={`px-2 py-0.5 rounded text-xs font-medium ${typeIcons[todo.type] || 'bg-gray-100 text-gray-800'}`}>
          {todo.type}
        </span>
        <span className={`px-2 py-0.5 rounded text-xs font-medium ${priorityColors[todo.priority] || priorityColors.P3}`}>
          {todo.priority}
        </span>
      </div>

      <p className="text-sm text-gray-900 dark:text-white mb-3 line-clamp-2">{todo.content}</p>

      <div className="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
        <div className="flex items-center gap-1">
          <FileText className="w-3 h-3" />
          <span className="truncate max-w-[120px]">{todo.file_path.split('/').pop()}:{todo.line_number}</span>
        </div>

        {todo.assignee && (
          <div className="flex items-center gap-1">
            <User className="w-3 h-3" />
            <span className="truncate max-w-[80px]">{todo.assignee}</span>
          </div>
        )}
      </div>

      {todo.due_date && (
        <div className={`flex items-center gap-1 mt-2 text-xs ${
          new Date(todo.due_date) < new Date() ? 'text-red-500' : 'text-gray-500 dark:text-gray-400'
        }`}>
          <Calendar className="w-3 h-3" />
          <span>{formatDate(todo.due_date)}</span>
          {new Date(todo.due_date) < new Date() && <AlertCircle className="w-3 h-3" />}
        </div>
      )}

      <div className="mt-2">
        <span className={`px-2 py-0.5 rounded text-xs font-medium ${statusColors[todo.status] || statusColors.open}`}>
          {todo.status.replace('_', ' ')}
        </span>
      </div>
    </div>
  );
}
