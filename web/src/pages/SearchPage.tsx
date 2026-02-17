import { useState } from 'react';
import type { TODO } from '../types';
import { api } from '../services/api';
import { Layout } from '../components/Layout';
import { TODOCard } from '../components/TODOCard';
import { Loader2, Search } from 'lucide-react';

export function SearchPage() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<TODO[]>([]);
  const [loading, setLoading] = useState(false);
  const [searched, setSearched] = useState(false);

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!query.trim()) return;

    try {
      setLoading(true);
      setSearched(true);
      const data = await api.searchTODOs(query);
      setResults(data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Layout>
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Search</h1>
          <p className="text-gray-600 dark:text-gray-400">Search for TODO items by content, assignee, or file</p>
        </div>

        <form onSubmit={handleSearch} className="flex gap-3">
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search todos..."
              className="w-full pl-10 pr-4 py-3 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <button
            type="submit"
            disabled={loading || !query.trim()}
            className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            {loading && <Loader2 className="w-4 h-4 animate-spin" />}
            Search
          </button>
        </form>

        {searched && !loading && (
          <div>
            <p className="text-gray-600 dark:text-gray-400 mb-4">
              Found {results.length} result{results.length !== 1 ? 's' : ''}
            </p>

            {results.length > 0 ? (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {results.map(todo => (
                  <TODOCard key={todo.id} todo={todo} />
                ))}
              </div>
            ) : (
              <div className="text-center py-12">
                <p className="text-gray-600 dark:text-gray-400">No results found</p>
              </div>
            )}
          </div>
        )}
      </div>
    </Layout>
  );
}
