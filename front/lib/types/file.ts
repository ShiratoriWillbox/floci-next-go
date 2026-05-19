/** POST /api/todos/:id/files のレスポンス（Go `gin.H`） */
export type TodoFileUploadInstruction = {
  file_id: number;
  upload_url: string;
  upload_method: string;
  /** Go `http.Header` の JSON（キー → 文字列の配列） */
  headers?: Record<string, string[]>;
};

/** GET /api/todos/:id/files/:file_id のレスポンス（プリサインド取得） */
export type TodoFileDownloadInstruction = {
  download_url: string;
  download_method: string;
  headers?: Record<string, string[]>;
};

/** PUT /api/todos/:id/files/:file_id のレスポンス（`dto.TodosFiles`） */
export type TodoFileLink = {
  id: number;
  todo_id: number;
  file_id: number;
  created_at: string;
  updated_at: string;
  file?: {
    id: number;
    name: string;
    created_at: string;
    updated_at: string;
  };
};
