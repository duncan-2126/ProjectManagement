import type { TODO, Stats, FilterOptions, TODOFormData } from '../types';

const API_BASE = '/api';

export const api = {
  async getTODOs(filters?: FilterOptions): Promise<TODO[]> {
    const params = new URLSearchParams();
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value) params.append(key, value);
      });
    }
    const response = await fetch(`${API_BASE}/todos?${params}`);
    if (!response.ok) throw new Error('Failed to fetch TODOs');
    const data = await response.json();
    return data.todos || [];
  },

  async getTODO(id: string): Promise<TODO> {
    const response = await fetch(`${API_BASE}/todo/${id}`);
    if (!response.ok) throw new Error('Failed to fetch TODO');
    return response.json();
  },

  async updateTODO(id: string, data: TODOFormData): Promise<TODO> {
    const response = await fetch(`${API_BASE}/todo/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    if (!response.ok) throw new Error('Failed to update TODO');
    return response.json();
  },

  async deleteTODO(id: string): Promise<void> {
    const response = await fetch(`${API_BASE}/todo/${id}`, {
      method: 'DELETE',
    });
    if (!response.ok) throw new Error('Failed to delete TODO');
  },

  async getStats(): Promise<Stats> {
    const response = await fetch(`${API_BASE}/stats`);
    if (!response.ok) throw new Error('Failed to fetch stats');
    return response.json();
  },

  async searchTODOs(query: string): Promise<TODO[]> {
    const response = await fetch(`${API_BASE}/search?q=${encodeURIComponent(query)}`);
    if (!response.ok) throw new Error('Failed to search TODOs');
    const data = await response.json();
    return data.todos || [];
  },
};
