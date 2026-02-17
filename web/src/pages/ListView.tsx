import { useState, useEffect } from 'react';
import type { TODO, FilterOptions } from '../types';
import { api } from '../services/api';
import { Layout } from '../components/Layout';
import { FilterBar } from '../components/FilterBar';
import { Loader2, FileText, User } from 'lucide-react';

export function ListView() {
  const [todos, setTodos] = useState<TODO[]>([]);
  const [filters, setFilters] = useState<FilterOptions>({});
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadData();
  }, [filters]);

  const loadData = async () => {
    try {
      setLoading(true);
      const data = await api.getTODOs(filters);
      setTodos(data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const priorityColors: Record<string, string> = {
    P0: 'text-red-600 dark:text-red-400',
    P1: 'text-orange-600 dark:text-orange-400',
    P2: 'text-yellow-600 dark:text-yellow-400',
    P3: 'text-blue-600 dark:text-blue-400',
    P4: 'text-gray-600 dark:text-gray-400',
  };

  const statusColors: Record<string, string> = {
    open: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300',
    in_progress: 'bg-blue-100 text-blue-800 dark:bg-blue-900/50 dark:text-blue-300',
    blocked: 'bg-red-100 text-red-800 dark:bg-red-900/50 dark:text-red-300',
    resolved: 'bg-green-100 text-green-800 dark:bg-green-900/50 dark:text-green-300',
    wontfix: 'bg-purple-100 text-purple-800 dark:bg-purple-900/50 dark:text-purple-300',
    closed: 'bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400',
  };

  return (
    <Layout>
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">List View</h1>
          <p className="text-gray-600 dark:text-gray-400">All TODO items in a table view</p>
        </div>

        <FilterBar filters={filters} onChange={setFilters} />

        {loading ? (
          <div className="flex items-center justify-center min-h-[300px]">
            <Loader2 className="w-8 h-8 animate-spin text-blue-500" />
          </div>
        ) : todos.length > 0 ? (
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-gray-50 dark:bg-gray-700/50">
                  <tr>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">ID</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Type</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Content</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Priority</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Assignee</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Location</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                  {todos.map(todo => (
                    <tr key={todo.id} className="hover:bg-gray-50 dark:hover:bg-gray-700/30">
                      <td className="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{todo.id.slice(0, 8)}</td>
                      <td className="px-4 py-3">
                        <span className="px-2 py-1 text-xs font-medium rounded bg-blue-100 text-blue-800 dark:bg-blue-900/50 dark:text-blue-300">
                          {todo.type}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-900 dark:text-white max-w-xs truncate">{todo.content}</td>
                      <td className={`px-4 py-3 text-sm font-medium ${priorityColors[todo.priority] || priorityColors.P3}`}>{todo.priority}</td>
                      <td className="px-4 py-3">
                        <span className={`px-2 py-1 text-xs font-medium rounded ${statusColors[todo.status] || statusColors.open}`}>
                          {todo.status.replace('_', ' ')}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-600 dark:text-gray-300">
                        {todo.assignee ? (
                          <span className="flex items-center gap-1">
                            <User className="w-3 h-3" />
                            {todo.assignee}
                          </span>
                        ) : '-'}
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">
                        <span className="flex items-center gap-1">
                          <FileText className="w-3 h-3" />
                          {todo.file_path.split('/').pop()}:{todo.line_number}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        ) : (
          <div className="text-center py-12">
            <p className="text-gray-600 dark:text-gray-400">No TODOs found</p>
          </div>
        )}
      </div>
    </Layout>
  );
}
