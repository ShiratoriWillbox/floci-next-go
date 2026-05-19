import type { TodoFileLink, TodoFileUploadInstruction, TodoFileDownloadInstruction } from "@/lib/types/file";
import type { Todo } from "@/lib/types/todo";

/** ブラウザは同一オリジンの `/api`（Next の rewrite）。サーバーは Go に直結。 */
export function apiBase(): string {
  if (typeof window !== "undefined") {
    return "";
  }
  return (process.env.SERVER_API_URL ?? "http://127.0.0.1:8080").replace(/\/$/, "");
}

export function apiUrl(path: string): string {
  const p = path.startsWith("/") ? path : `/${path}`;
  const base = apiBase();
  return base ? `${base}${p}` : p;
}

async function readErrorMessage(res: Response): Promise<string> {
  try {
    const body: unknown = await res.json();
    if (
      body &&
      typeof body === "object" &&
      "error" in body &&
      typeof (body as { error: unknown }).error === "string"
    ) {
      return (body as { error: string }).error;
    }
  } catch {
    /* ignore */
  }
  return res.statusText || `HTTP ${res.status}`;
}

export async function listTodos(): Promise<Todo[]> {
  const res = await fetch(apiUrl("/api/todos"), { cache: "no-store" });
  if (!res.ok) throw new Error(await readErrorMessage(res));
  return res.json() as Promise<Todo[]>;
}

export async function createTodo(input: { title: string; completed?: boolean }): Promise<Todo> {
  const res = await fetch(apiUrl("/api/todos"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      title: input.title,
      ...(input.completed !== undefined ? { completed: input.completed } : {}),
    }),
  });
  if (!res.ok) throw new Error(await readErrorMessage(res));
  return res.json() as Promise<Todo>;
}

export async function updateTodo(
  id: number,
  patch: { title?: string; completed?: boolean },
): Promise<Todo> {
  const res = await fetch(apiUrl(`/api/todos/${id}`), {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(patch),
  });
  if (!res.ok) throw new Error(await readErrorMessage(res));
  return res.json() as Promise<Todo>;
}

export async function deleteTodo(id: number): Promise<void> {
  const res = await fetch(apiUrl(`/api/todos/${id}`), {
    method: "DELETE",
  });
  if (res.status === 204) return;
  if (!res.ok) throw new Error(await readErrorMessage(res));
}

function presignedHeadersInit(raw: TodoFileUploadInstruction["headers"]): Headers {
  const h = new Headers();
  if (!raw) return h;
  for (const [key, values] of Object.entries(raw)) {
    if (!Array.isArray(values)) continue;
    for (const v of values) {
      if (typeof v === "string" && v.length > 0) {
        h.append(key, v);
      }
    }
  }
  return h;
}

/** POST /api/todos/:id/files — アップロード用プリサインド URL を取得 */
export async function createTodoFileUpload(
  todoId: number,
  input?: { name?: string },
): Promise<TodoFileUploadInstruction> {
  const res = await fetch(apiUrl(`/api/todos/${todoId}/files`), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input?.name ? { name: input.name } : {}),
  });
  if (!res.ok) throw new Error(await readErrorMessage(res));
  return res.json() as Promise<TodoFileUploadInstruction>;
}

/** プリサインド URL へオブジェクト本文を PUT（S3 側の CORS 設定が必要） */
export async function putBlobToPresignedUpload(
  instruction: TodoFileUploadInstruction,
  body: Blob,
): Promise<void> {
  const method = instruction.upload_method || "PUT";
  const res = await fetch(instruction.upload_url, {
    method,
    headers: presignedHeadersInit(instruction.headers),
    body,
  });
  if (!res.ok) {
    const hint =
      res.status === 0 || res.type === "opaque"
        ? "（ブラウザのネットワーク制限・S3 の CORS 等を確認してください）"
        : "";
    throw new Error(
      `オブジェクトストレージへのアップロードに失敗しました (${res.status})${hint}`,
    );
  }
}

/** PUT /api/todos/:id/files/:file_id — アップロード済みファイルを Todo に紐づけ */
export async function attachTodoFile(todoId: number, fileId: number): Promise<TodoFileLink> {
  const res = await fetch(apiUrl(`/api/todos/${todoId}/files/${fileId}`), {
    method: "PUT",
  });
  if (!res.ok) throw new Error(await readErrorMessage(res));
  return res.json() as Promise<TodoFileLink>;
}

/**
 * 1) プリサインド URL 取得 → 2) PUT でバイナリ送信 → 3) Todo に紐づけ
 */
export async function uploadAndAttachTodoFile(
  todoId: number,
  file: File,
): Promise<TodoFileLink> {
  const name = file.name?.trim() || undefined;
  const instruction = await createTodoFileUpload(todoId, name ? { name } : undefined);
  await putBlobToPresignedUpload(instruction, file);
  return attachTodoFile(todoId, instruction.file_id);
}

/** GET /api/todos/:id/files/:file_id — ダウンロード用プリサインド URL */
export async function getTodoFileDownload(
  todoId: number,
  fileId: number,
): Promise<TodoFileDownloadInstruction> {
  const res = await fetch(apiUrl(`/api/todos/${todoId}/files/${fileId}`), { cache: "no-store" });
  if (!res.ok) throw new Error(await readErrorMessage(res));
  return res.json() as Promise<TodoFileDownloadInstruction>;
}
