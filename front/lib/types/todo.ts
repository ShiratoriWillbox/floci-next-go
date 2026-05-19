/** Go `dto.Todo` の JSON（`files` は `todos_files` 経由で結合された `dto.File`） */
export type TodoFile = {
  id: number;
  name: string;
  created_at: string;
  updated_at: string;
};

export type Todo = {
  id: number;
  title: string;
  completed: boolean;
  created_at: string;
  updated_at: string;
  files?: TodoFile[];
};
