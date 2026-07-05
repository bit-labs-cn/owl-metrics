import { http } from "@bit-labs.cn/owl-ui/utils/http";

export type MetricsAppRow = {
  id: string;
  name: string;
  appKey: string;
  appSecret?: string;
  allowedOrigins: string[];
  status: number;
  description?: string;
  createdAt?: number;
};

class MetricsAppAPI {
  getList = (params?: object) =>
    http.request<ResultTable>("get", "/api/v1/metrics/admin/apps", { params }, { silentMessage: true });

  create = (data?: object) => http.request<Result>("post", "/api/v1/metrics/admin/apps", { data });

  detail = (id: string | number) => http.request<Result>("get", `/api/v1/metrics/admin/apps/${id}`);

  update = (id: string | number, data?: object) =>
    http.request<Result>("put", `/api/v1/metrics/admin/apps/${id}`, { data });

  remove = (id: string | number) => http.request<Result>("delete", `/api/v1/metrics/admin/apps/${id}`);

  regenerateSecret = (id: string | number) =>
    http.request<Result>("put", `/api/v1/metrics/admin/apps/${id}/regenerate-secret`);
}

export const metricsAppAPI = new MetricsAppAPI();
