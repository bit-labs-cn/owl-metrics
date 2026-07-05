import { http } from "@bit-labs.cn/owl-ui/utils/http";

export type OverviewData = {
  pageViews: number;
  uniqueUsers: number;
  downloads: number;
  totalEvents: number;
  byName: { name: string; count: number }[];
  fromDay: string;
  toDay: string;
};

export type TimeseriesPoint = {
  day: string;
  name: string;
  count: number;
  uniques?: number;
};

export type TopNItem = {
  dimValue: string;
  count: number;
};

export type EventRow = {
  id: string;
  name: string;
  platform: string;
  anonId: string;
  url: string;
  os: string;
  serverTs: number;
  day: string;
};

class MetricsQueryAPI {
  overview = (params?: object) =>
    http.request<Result>("get", "/api/v1/metrics/admin/overview", { params }, { silentMessage: true });

  timeseries = (params?: object) =>
    http.request<Result>("get", "/api/v1/metrics/admin/timeseries", { params }, { silentMessage: true });

  topn = (params?: object) =>
    http.request<Result>("get", "/api/v1/metrics/admin/topn", { params }, { silentMessage: true });

  events = (params?: object) =>
    http.request<ResultTable>("get", "/api/v1/metrics/admin/events", { params }, { silentMessage: true });
}

export const metricsQueryAPI = new MetricsQueryAPI();
