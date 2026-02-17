import type { FilterOptions, TODOStatus, TODOPriority } from '../types';

interface FilterBarProps {
  filters: FilterOptions;
  onChange: (filters: FilterOptions) => void;
}

const statuses: TODOStatus[] = ['open', 'in_progress', 'blocked', 'resolved', 'wontfix', 'closed'];
const priorities: TODOPriority[] = ['P0', 'P1', 'P2', 'P3', 'P4'];

export function FilterBar({ filters, onChange }: FilterBarProps) {
  const handleChange = (key: keyof FilterOptions, value: string) => {
    onChange({ ...filters, [key]: value || undefined });
  };

  return (
    <div className="flex flex-wrap gap-3 items-center p-4 bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700">
      <div className="flex-1 min-w-[200px]">
        <input
          type="text"
          placeholder="Search todos..."
          value={filters.search || ''}
          onChange={(e) => handleChange('search', e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>

      <select
        value={filters.status || ''}
        onChange={(e) => handleChange('status', e.target.value)}
        className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
      >
        <option value="">All Statuses</option>
        {statuses.map(status => (
          <option key={status} value={status}>{status.replace('_', ' ')}</option>
        ))}
      </select>

      <select
        value={filters.priority || ''}
        onChange={(e) => handleChange('priority', e.target.value)}
        className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
      >
        <option value="">All Priorities</option>
        {priorities.map(priority => (
          <option key={priority} value={priority}>{priority}</option>
        ))}
      </select>

      <input
        type="text"
        placeholder="Assignee"
        value={filters.assignee || ''}
        onChange={(e) => handleChange('assignee', e.target.value)}
        className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 w-32"
      />

      <select
        value={filters.type || ''}
        onChange={(e) => handleChange('type', e.target.value)}
        className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
      >
        <option value="">All Types</option>
        <option value="TODO">TODO</option>
        <option value="FIXME">FIXME</option>
        <option value="HACK">HACK</option>
        <option value="BUG">BUG</option>
        <option value="NOTE">NOTE</option>
        <option value="XXX">XXX</option>
      </select>

      <button
        onClick={() => onChange({})}
        className="px-3 py-2 text-sm text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white"
      >
        Clear
      </button>
    </div>
  );
}
