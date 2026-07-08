/** 解析 owl router.PageSuccess：list 在 data，total 在根级 */
export function pickPageList<T>(res: unknown): { list: T[]; total: number } {
  const r = res as { data?: { list?: T[] }; total?: number };
  return { list: r.data?.list ?? [], total: Number(r.total ?? 0) };
}
