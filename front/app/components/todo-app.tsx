"use client";

import { useCallback, useMemo, useState } from "react";

import {
  createTodo,
  deleteTodo,
  getTodoFileDownload,
  listTodos,
  updateTodo,
  uploadAndAttachTodoFile,
} from "@/lib/api";
import type { Todo } from "@/lib/types/todo";

type Props = {
  initialTodos: Todo[];
  initialError: string | null;
};

function fileBusyToken(todoId: number, fileId: number) {
  return `${todoId}-${fileId}`;
}

export function TodoApp({ initialTodos, initialError }: Props) {
  const [todos, setTodos] = useState<Todo[]>(initialTodos);
  const [error, setError] = useState<string | null>(initialError);
  const [draft, setDraft] = useState("");
  const [busyId, setBusyId] = useState<number | null>(null);
  const [fileBusyKey, setFileBusyKey] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const sorted = useMemo(() => [...todos].sort((a, b) => a.id - b.id), [todos]);

  const reload = useCallback(async () => {
    setError(null);
    try {
      const rows = await listTodos();
      setTodos(rows);
    } catch (e) {
      setError(e instanceof Error ? e.message : "一覧の取得に失敗しました");
    }
  }, []);

  const onAdd = async (e: React.FormEvent) => {
    e.preventDefault();
    const title = draft.trim();
    if (!title || submitting) return;
    setSubmitting(true);
    setError(null);
    try {
      const row = await createTodo({ title });
      setTodos((prev) => [...prev, { ...row, files: row.files ?? [] }]);
      setDraft("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "作成に失敗しました");
    } finally {
      setSubmitting(false);
    }
  };

  const onToggle = async (row: Todo) => {
    setBusyId(row.id);
    setError(null);
    try {
      const next = await updateTodo(row.id, { completed: !row.completed });
      setTodos((prev) => prev.map((t) => (t.id === next.id ? next : t)));
    } catch (e) {
      setError(e instanceof Error ? e.message : "更新に失敗しました");
    } finally {
      setBusyId(null);
    }
  };

  const onDelete = async (id: number) => {
    setBusyId(id);
    setError(null);
    try {
      await deleteTodo(id);
      setTodos((prev) => prev.filter((t) => t.id !== id));
    } catch (e) {
      setError(e instanceof Error ? e.message : "削除に失敗しました");
    } finally {
      setBusyId(null);
    }
  };

  const onAttachFile = async (todoId: number, e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    e.target.value = "";
    if (!file) return;

    setBusyId(todoId);
    setError(null);
    try {
      await uploadAndAttachTodoFile(todoId, file);
      const rows = await listTodos();
      setTodos(rows);
    } catch (err) {
      setError(err instanceof Error ? err.message : "ファイルの添付に失敗しました");
    } finally {
      setBusyId(null);
    }
  };

  const onDownloadFile = async (todoId: number, fileId: number) => {
    const key = fileBusyToken(todoId, fileId);
    setFileBusyKey(key);
    setError(null);
    try {
      const d = await getTodoFileDownload(todoId, fileId);
      window.open(d.download_url, "_blank", "noopener,noreferrer");
    } catch (err) {
      setError(err instanceof Error ? err.message : "ダウンロード URL の取得に失敗しました");
    } finally {
      setFileBusyKey(null);
    }
  };

  return (
    <div className="mx-auto flex w-full max-w-lg flex-col gap-8 px-4 py-10">
      <header className="space-y-1">
        <h1 className="text-2xl font-semibold tracking-tight">Todo（API 連携）</h1>
        <p className="text-sm text-zinc-500 dark:text-zinc-400">
          <code className="rounded bg-zinc-100 px-1.5 py-0.5 text-xs dark:bg-zinc-800">
            GET /api/todos
          </code>{" "}
          で各 Todo に紐づくファイル一覧が返り、
          <code className="rounded bg-zinc-100 px-1.5 py-0.5 text-xs dark:bg-zinc-800">
            GET /api/todos/:id/files/:file_id
          </code>{" "}
          でプリサインド取得 URL が取れます。
        </p>
      </header>

      {error ? (
        <div
          className="flex flex-col gap-2 rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-800 dark:border-red-900 dark:bg-red-950/40 dark:text-red-200"
          role="alert"
        >
          <span>{error}</span>
          <button
            type="button"
            onClick={() => reload()}
            className="self-start rounded-md bg-red-100 px-3 py-1.5 text-xs font-medium text-red-900 hover:bg-red-200 dark:bg-red-900/60 dark:text-red-100 dark:hover:bg-red-900"
          >
            再読み込み
          </button>
        </div>
      ) : null}

      <form onSubmit={onAdd} className="flex gap-2">
        <input
          className="min-w-0 flex-1 rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm ring-zinc-400 outline-none focus:ring-2 dark:border-zinc-700 dark:bg-zinc-950"
          placeholder="タイトルを入力…"
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          disabled={submitting}
          aria-label="新しい Todo のタイトル"
        />
        <button
          type="submit"
          disabled={submitting || !draft.trim()}
          className="shrink-0 rounded-lg bg-foreground px-4 py-2 text-sm font-medium text-background disabled:opacity-40"
        >
          追加
        </button>
      </form>

      <ul className="flex flex-col gap-2">
        {sorted.length === 0 ? (
          <li className="rounded-lg border border-dashed border-zinc-200 py-10 text-center text-sm text-zinc-500 dark:border-zinc-800 dark:text-zinc-400">
            Todo がありません。上のフォームから追加してください。
          </li>
        ) : (
          sorted.map((row) => {
            const busy = busyId === row.id;
            const files = row.files ?? [];
            return (
              <li
                key={row.id}
                className="flex items-start gap-3 rounded-lg border border-zinc-200 bg-white p-3 dark:border-zinc-800 dark:bg-zinc-950"
              >
                <input
                  type="checkbox"
                  checked={row.completed}
                  disabled={busy}
                  onChange={() => onToggle(row)}
                  className="mt-1 size-4 rounded border-zinc-300"
                  aria-label={`「${row.title}」を完了にする`}
                />
                <div className="min-w-0 flex-1 space-y-2">
                  <p
                    className={`text-sm leading-snug ${
                      row.completed ? "text-zinc-400 line-through dark:text-zinc-500" : ""
                    }`}
                  >
                    {row.title}
                  </p>
                  <p className="text-xs text-zinc-400 dark:text-zinc-500">id: {row.id}</p>
                  {files.length > 0 ? (
                    <ul className="flex flex-col gap-1.5 text-xs text-zinc-600 dark:text-zinc-400">
                      {files.map((f) => {
                        const fKey = fileBusyToken(row.id, f.id);
                        const fileBusy = fileBusyKey === fKey;
                        return (
                          <li
                            key={f.id}
                            className="flex flex-wrap items-center justify-between gap-2 rounded-md border border-zinc-100 bg-zinc-50/80 px-2 py-1.5 dark:border-zinc-800 dark:bg-zinc-900/40"
                          >
                            <span className="min-w-0 truncate" title={f.name}>
                              {f.name}{" "}
                              <span className="text-zinc-400 dark:text-zinc-500">#{f.id}</span>
                            </span>
                            <button
                              type="button"
                              disabled={fileBusy || busy}
                              onClick={() => void onDownloadFile(row.id, f.id)}
                              className="shrink-0 rounded border border-zinc-200 bg-white px-2 py-0.5 text-[11px] text-zinc-700 hover:bg-zinc-100 disabled:opacity-40 dark:border-zinc-600 dark:bg-zinc-950 dark:text-zinc-200 dark:hover:bg-zinc-800"
                            >
                              {fileBusy ? "…" : "ダウンロード"}
                            </button>
                          </li>
                        );
                      })}
                    </ul>
                  ) : null}
                  <label className="inline-flex cursor-pointer items-center gap-2">
                    <span className="rounded-md border border-zinc-200 px-2 py-1 text-xs text-zinc-600 hover:bg-zinc-50 dark:border-zinc-700 dark:text-zinc-300 dark:hover:bg-zinc-900">
                      ファイルを添付
                    </span>
                    <input
                      type="file"
                      className="sr-only"
                      disabled={busy}
                      onChange={(e) => void onAttachFile(row.id, e)}
                      aria-label={`Todo ${row.id} にファイルを添付`}
                    />
                  </label>
                </div>
                <button
                  type="button"
                  disabled={busy}
                  onClick={() => onDelete(row.id)}
                  className="shrink-0 rounded-md border border-zinc-200 px-2 py-1 text-xs text-zinc-600 hover:bg-zinc-50 disabled:opacity-40 dark:border-zinc-700 dark:text-zinc-300 dark:hover:bg-zinc-900"
                >
                  削除
                </button>
              </li>
            );
          })
        )}
      </ul>
    </div>
  );
}
