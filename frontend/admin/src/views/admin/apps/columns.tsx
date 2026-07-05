import { tableIndexColumn, type TablePaginationInput } from "@bit-labs.cn/owl-ui/utils/tableIndexColumn";

const statusMap: Record<number, string> = {
  1: "启用",
  2: "禁用"
};

export function createColumns(options: { pagination?: TablePaginationInput } = {}) {
  return [
    tableIndexColumn(options.pagination),
    { label: "名称", prop: "name", minWidth: 140 },
    { label: "AppKey", prop: "appKey", minWidth: 220, showOverflowTooltip: true },
    {
      label: "状态",
      prop: "status",
      width: 80,
      cellRenderer: ({ row }) => <span>{statusMap[row.status] ?? row.status}</span>
    },
    { label: "描述", prop: "description", minWidth: 160, showOverflowTooltip: true },
    { label: "操作", fixed: "right", width: 260, slot: "operation" }
  ] satisfies TableColumnList;
}
