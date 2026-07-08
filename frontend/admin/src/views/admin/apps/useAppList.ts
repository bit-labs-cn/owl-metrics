import { reactive, ref, onMounted } from "vue";
import { metricsAppAPI, type MetricsAppRow } from "../../../api/app";
import { pickPageList } from "../../../utils/page";
import { createColumns } from "./columns";

export function useAppList() {
  const loading = ref(false);
  const dataList = ref<MetricsAppRow[]>([]);
  const pagination = reactive({ total: 0, pageSize: 10, currentPage: 1, background: true });

  const columns = createColumns({ pagination });

  async function onSearch() {
    loading.value = true;
    try {
      const res = await metricsAppAPI.getList({
        page: pagination.currentPage,
        pageSize: pagination.pageSize
      });
      const page = pickPageList<MetricsAppRow>(res);
      dataList.value = page.list;
      pagination.total = page.total;
    } finally {
      loading.value = false;
    }
  }

  function handleSizeChange(val: number) {
    pagination.pageSize = val;
    onSearch();
  }

  function handleCurrentChange(val: number) {
    pagination.currentPage = val;
    onSearch();
  }

  onMounted(onSearch);

  return { loading, dataList, columns, pagination, onSearch, handleSizeChange, handleCurrentChange };
}
