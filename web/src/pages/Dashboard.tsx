import { useState, useEffect } from 'react';
import type { TODO, Stats, FilterOptions } from '../types';
import { api } from '../services/api';
import { Layout } from '../components/Layout';
import { FilterBar } from '../components/FilterBar';
import { TODOCard } from '../components/TODOCard';
import { Loader2, AlertCircle, Clock, CheckCircle, XCircle, BarChart } from 'lucide-react';

export function Dashboard() {
  const [todos, setTodos] = useState<TODO[]>([]);
  const [stats, setStats] = useState<Stats | null>(null);
  const [filters, setFilters] = useState<FilterOptions>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadData();
  }, [filters]);

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);
      const [todosData, statsData] = await Promise.all([
        api.getTODOs(filters),
        api.getStats(),
      ]);
      setTodos(todosData);
      setStats(statsData);
    } catch (err) {
      setError('Failed to load data. Make sure the server is running.');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const inProgressTodos = todos.filter(t => t.status === 'in_progress');
  const openTodos = todos.filter(t => t.status === 'open');

  const statusCards = stats ? [
    { label: 'Total', value: stats.total, icon: BarChart, color: 'bg-blue-500' },
    { label: 'Open', value: stats.by_status?.open || 0, icon: Clock, color: 'bg-gray-500' },
    { label: 'In Progress', value: stats.by_status?.in_progress || 0, icon: Loader2, color: 'bg-blue-500' },
    { label: 'Resolved', value: stats.by_status?.resolved || 0, icon: CheckCircle, color: 'bg-green-500' },
    { label: 'Blocked', value: stats.by_status?.blocked || 0, icon: XCircle, color: 'bg-red-500' },
    { label: 'Closed', value: stats.by_status?.closed || 0, icon: CheckCircle, color: 'bg-purple-500' },
  ] : [];

  if (error) {
    return (
      <Layout>
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="text-center">
            <AlertCircle className="w-12 h-12 text-red-500 mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-2">Connection Error</h2>
            <p className="text-gray-600 dark:text-gray-400 mb-4">{error}</p>
            <p className="text-sm text-gray-500 dark:text-gray-500">Start the server with: <code className="bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded">go run main.go serve</code></p>
          </div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Dashboard</h1>
          <p className="text-gray-600 dark:text-gray-400">Overview of your TODO items</p>
        </div>

        {loading && !todos.length ? (
          <div className="flex items-center justify-center min-h-[300px]">
            <Loader2 className="w-8 h-8 animate-spin text-blue-500" />
          </div>
        ) : (
          <>
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
              {statusCards.map(card => (
                <div key={card.label} className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
                  <div className="flex items-center gap-3">
                    <div className={`p-2 rounded-lg ${card.color}`}>
                      <card.icon className="w-5 h-5 text-white" />
                    </div>
                    <div>
                      <p className="text-2xl font-bold text-gray-900 dark:text-white">{card.value}</p>
                      <p className="text-sm text-gray-500 dark:text-gray-400">{card.label}</p>
                    </div>
                  </div>
                </div>
              ))}
            </div>

            <FilterBar filters={filters} onChange={setFilters} />

            {inProgressTodos.length > 0 && (
              <div>
                <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">In Progress ({inProgressTodos.length})</h2>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {inProgressTodos.map(todo => (
                    <TODOCard key={todo.id} todo={todo} />
                  ))}
                </div>
              </div>
            )}

            {openTodos.length > 0 && (
              <div>
                <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Open ({openTodos.length})</h2>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {openTodos.slice(0, 6).map(todo => (
                    <TODOCard key={todo.id} todo={todo} />
                  ))}
                </div>
              </div>
            )}

            {todos.length === 0 && (
              <div className="text-center py-12">
                <p className="text-gray-600 dark:text-gray-400">No TODOs found. Run <code className="bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded">todo scan</code> to scan for TODOs.</p>
              </div>
            )}
          </>
        )}
      </div>
    </Layout>
  );
}
