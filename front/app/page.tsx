import { listTodos } from "@/lib/api";

import { TodoApp } from "./components/todo-app";

export default async function Home() {
  let initialTodos: Awaited<ReturnType<typeof listTodos>> = [];
  let initialError: string | null = null;

  try {
    initialTodos = await listTodos();
  } catch (e) {
    initialError = e instanceof Error ? e.message : "バックエンドからの取得に失敗しました";
  }

  return <TodoApp initialTodos={initialTodos} initialError={initialError} />;
}
