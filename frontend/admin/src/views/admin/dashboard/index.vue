<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import VChart from "vue-echarts";
import { use } from "echarts/core";
import { CanvasRenderer } from "echarts/renderers";
import { LineChart, BarChart, PieChart } from "echarts/charts";
import { GridComponent, TooltipComponent, LegendComponent } from "echarts/components";
import dayjs from "dayjs";
import { metricsAppAPI, type MetricsAppRow } from "../../../api/app";
import { metricsQueryAPI, type OverviewData, type TimeseriesPoint, type TopNItem, type EventRow } from "../../../api/query";
import { pickPageList } from "../../../utils/page";
import { createEventColumns } from "./eventColumns";
import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";

use([CanvasRenderer, LineChart, BarChart, PieChart, GridComponent, TooltipComponent, LegendComponent]);

defineOptions({ name: "MetricsDashboard" });

const apps = ref<MetricsAppRow[]>([]);
const appId = ref("");
const fromDay = ref(dayjs().subtract(29, "day").format("YYYY-MM-DD"));
const toDay = ref(dayjs().format("YYYY-MM-DD"));
const overview = ref<OverviewData | null>(null);
const timeseries = ref<TimeseriesPoint[]>([]);
const topPlatform = ref<TopNItem[]>([]);
const topOS = ref<TopNItem[]>([]);
const events = ref<EventRow[]>([]);
const loading = ref(false);
const eventTotal = ref(0);
const eventPage = ref(1);
const eventPageSize = ref(10);

const eventColumns = createEventColumns();

function lastMonthRange() {
  return {
    fromDay: dayjs().subtract(29, "day").format("YYYY-MM-DD"),
    toDay: dayjs().format("YYYY-MM-DD")
  };
}

function buildDayRange(from: string, to: string): string[] {
  const days: string[] = [];
  let cur = dayjs(from);
  const end = dayjs(to);
  while (cur.isBefore(end) || cur.isSame(end, "day")) {
    days.push(cur.format("YYYY-MM-DD"));
    cur = cur.add(1, "day");
  }
  return days;
}

const lineOption = computed(() => {
  const { fromDay: trendFrom, toDay: trendTo } = lastMonthRange();
  const days = buildDayRange(trendFrom, trendTo);
  const names = [...new Set(timeseries.value.map(p => p.name))];
  return {
    tooltip: { trigger: "axis" },
    legend: { data: names },
    grid: { left: 40, right: 20, bottom: 30, top: 40 },
    xAxis: {
      type: "category",
      data: days,
      axisLabel: { rotate: days.length > 14 ? 45 : 0, hideOverlap: true }
    },
    yAxis: { type: "value", minInterval: 1 },
    series: names.map(name => ({
      name,
      type: "line",
      smooth: true,
      showSymbol: false,
      data: days.map(d => timeseries.value.find(p => p.day === d && p.name === name)?.count ?? 0)
    }))
  };
});

const platformPie = computed(() => ({
  tooltip: { trigger: "item" },
  series: [
    {
      type: "pie",
      radius: "65%",
      data: topPlatform.value.map(i => ({ name: i.dimValue || "unknown", value: i.count }))
    }
  ]
}));

const osBar = computed(() => ({
  tooltip: { trigger: "axis" },
  grid: { left: 40, right: 20, bottom: 30, top: 20 },
  xAxis: { type: "category", data: topOS.value.map(i => i.dimValue || "unknown") },
  yAxis: { type: "value", minInterval: 1 },
  series: [{ type: "bar", data: topOS.value.map(i => i.count) }]
}));

async function loadApps() {
  const { data } = await metricsAppAPI.getList({ page: 1, pageSize: 100 });
  apps.value = (data?.list ?? []) as MetricsAppRow[];
  if (!appId.value && apps.value.length) {
    appId.value = apps.value[0].id;
  }
}

async function refresh() {
  if (!appId.value) return;
  loading.value = true;
  const params = { appId: appId.value, fromDay: fromDay.value, toDay: toDay.value };
  try {
    const [ov, ts, tp, tos, ev] = await Promise.all([
      metricsQueryAPI.overview(params),
      metricsQueryAPI.timeseries({ appId: appId.value, ...lastMonthRange() }),
      metricsQueryAPI.topn({ ...params, name: "page_view", dimKey: "platform", limit: 8 }),
      metricsQueryAPI.topn({ ...params, name: "download", dimKey: "os", limit: 8 }),
      metricsQueryAPI.events({ ...params, page: eventPage.value, pageSize: eventPageSize.value })
    ]);
    overview.value = ov.data as OverviewData;
    timeseries.value = (ts.data ?? []) as TimeseriesPoint[];
    topPlatform.value = (tp.data ?? []) as TopNItem[];
    topOS.value = (tos.data ?? []) as TopNItem[];
    const page = pickPageList<EventRow>(ev);
    events.value = page.list;
    eventTotal.value = page.total;
  } finally {
    loading.value = false;
  }
}

watch(appId, refresh);
onMounted(async () => {
  await loadApps();
  await refresh();
});
</script>

<template>
  <div class="main metrics-dashboard">
    <el-card shadow="never" class="mb-4">
      <el-form inline>
        <el-form-item label="应用">
          <el-select v-model="appId" placeholder="选择应用" style="width: 220px">
            <el-option v-for="a in apps" :key="a.id" :label="a.name" :value="a.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="开始">
          <el-date-picker v-model="fromDay" type="date" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item label="结束">
          <el-date-picker v-model="toDay" type="date" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="() => { eventPage = 1; refresh(); }">查询</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-row :gutter="16" class="mb-4">
      <el-col :span="6">
        <el-statistic title="首页访问(PV)" :value="overview?.pageViews ?? 0" />
      </el-col>
      <el-col :span="6">
        <el-statistic title="独立访客(UV)" :value="overview?.uniqueUsers ?? 0" />
      </el-col>
      <el-col :span="6">
        <el-statistic title="下载次数" :value="overview?.downloads ?? 0" />
      </el-col>
      <el-col :span="6">
        <el-statistic title="总事件数" :value="overview?.totalEvents ?? 0" />
      </el-col>
    </el-row>

    <el-row :gutter="16" class="mb-4">
      <el-col :span="14">
        <el-card shadow="never" header="趋势（按天，最近 30 天）">
          <VChart :option="lineOption" autoresize style="height: 320px" />
        </el-card>
      </el-col>
      <el-col :span="10">
        <el-card shadow="never" header="访问平台分布">
          <VChart :option="platformPie" autoresize style="height: 320px" />
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" class="mb-4">
      <el-col :span="24">
        <el-card shadow="never" header="下载按 OS">
          <VChart :option="osBar" autoresize style="height: 260px" />
        </el-card>
      </el-col>
    </el-row>

    <PureTableBar title="事件浏览器" :columns="eventColumns" @refresh="refresh">
      <template v-slot="{ size, dynamicColumns }">
        <pure-table
          row-key="id"
          :loading="loading"
          :size="size"
          :data="events"
          :columns="dynamicColumns"
          :pagination="{ total: eventTotal, pageSize: eventPageSize, currentPage: eventPage }"
          @page-size-change="(v: number) => { eventPageSize = v; refresh(); }"
          @page-current-change="(v: number) => { eventPage = v; refresh(); }"
        />
      </template>
    </PureTableBar>
  </div>
</template>

<style scoped>
.metrics-dashboard :deep(.el-statistic__number) {
  font-size: 28px;
}
.mb-4 {
  margin-bottom: 16px;
}
</style>
