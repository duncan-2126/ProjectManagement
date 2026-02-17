export interface TODO {
  id: string;
  file_path: string;
  line_number: number;
  column: number;
  type: string;
  content: string;
  author: string;
  email: string;
  created_at: string;
  updated_at: string;
  status: TODOStatus;
  priority: TODOPriority;
  category: string;
  assignee: string;
  due_date: string | null;
  estimate: number | null;
  hash: string;
}

export type TODOStatus = 'open' | 'in_progress' | 'blocked' | 'resolved' | 'wontfix' | 'closed';
export type TODOPriority = 'P0' | 'P1' | 'P2' | 'P3' | 'P4';

export interface Stats {
  total: number;
  by_status: Record<TODOStatus, number>;
  by_type: Record<string, number>;
  by_priority: Record<TODOPriority, number>;
}

export interface FilterOptions {
  status?: TODOStatus;
  priority?: TODOPriority;
  assignee?: string;
  type?: string;
  search?: string;
}

export interface TODOFormData {
  status?: TODOStatus;
  priority?: TODOPriority;
  assignee?: string;
  content?: string;
  category?: string;
  due_date?: string;
}
