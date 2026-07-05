import dayjs from "dayjs";
import type { EventRow } from "../../../api/query";

export function createEventColumns() {
  return [
    { label: "事件", prop: "name", minWidth: 120 },
    { label: "平台", prop: "platform", width: 90 },
    { label: "OS", prop: "os", width: 90 },
    { label: "URL", prop: "url", minWidth: 200, showOverflowTooltip: true },
    {
      label: "时间",
      prop: "serverTs",
      width: 170,
      cellRenderer: ({ row }: { row: EventRow }) => (
        <span>{row.serverTs ? dayjs(row.serverTs).format("YYYY-MM-DD HH:mm:ss") : "-"}</span>
      )
    }
  ] satisfies TableColumnList;
}
