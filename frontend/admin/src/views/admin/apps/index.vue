<script setup lang="ts">
import { ref, h } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { useAppList } from "./useAppList";
import AppForm from "./AppForm.vue";
import type { MetricsAppFormData } from "./AppForm.vue";
import type { MetricsAppRow } from "../../../api/app";
import { metricsAppAPI } from "../../../api/app";
import { addDialog } from "@bit-labs.cn/owl-ui/components/ReDialog";
import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";
import AddFill from "@iconify-icons/ri/add-circle-line";
import EditPen from "@iconify-icons/ep/edit-pen";
import Delete from "@iconify-icons/ep/delete";
import Key from "@iconify-icons/ep/key";

defineOptions({ name: "MetricsApps" });

const appFormRef = ref();
const { loading, dataList, columns, pagination, onSearch, handleSizeChange, handleCurrentChange } = useAppList();

function parseOrigins(text: string) {
  return text
    .split("\n")
    .map(s => s.trim())
    .filter(Boolean);
}

function formatOrigins(origins?: string[]) {
  return (origins ?? []).join("\n");
}

function openDialog(title = "新增", row?: MetricsAppRow) {
  const isEdit = title === "编辑";
  addDialog({
    title: `${title}指标应用`,
    props: {
      formInline: {
        id: row?.id,
        name: row?.name ?? "",
        description: row?.description ?? "",
        allowedOriginsText: formatOrigins(row?.allowedOrigins),
        status: row?.status ?? 1
      } satisfies MetricsAppFormData,
      isEdit
    },
    width: "560px",
    draggable: true,
    closeOnClickModal: false,
    contentRenderer: ({ options }) =>
      h(AppForm, {
        ref: appFormRef,
        formInline: options.props.formInline,
        isEdit: options.props.isEdit
      }),
    beforeSure: done => {
      const FormRef = appFormRef.value.getRef();
      const cur = appFormRef.value.getFormData() as MetricsAppFormData;
      FormRef.validate((valid: boolean) => {
        if (!valid) return;
        const payload = {
          name: cur.name,
          description: cur.description,
          allowedOrigins: parseOrigins(cur.allowedOriginsText),
          status: cur.status
        };
        const req = isEdit ? metricsAppAPI.update(cur.id!, payload) : metricsAppAPI.create(payload);
        req.then(res => {
          if (!isEdit && res?.data?.appKey) {
            ElMessage.success(`创建成功 AppKey: ${res.data.appKey}`);
          }
          done();
          onSearch();
        });
      });
    }
  });
}

function handleDelete(row: MetricsAppRow) {
  metricsAppAPI.remove(row.id).then(() => onSearch());
}

async function showSecret(row: MetricsAppRow) {
  const { data } = await metricsAppAPI.detail(row.id);
  const secret = data?.appSecret ?? "";
  await ElMessageBox.alert(
    `AppKey: ${data?.appKey}\nAppSecret: ${secret}`,
    "接入凭证",
    { confirmButtonText: "知道了", type: "info" }
  );
}

async function regenerateSecret(row: MetricsAppRow) {
  const { data } = await metricsAppAPI.regenerateSecret(row.id);
  ElMessage.success(`新 AppSecret: ${data?.appSecret ?? ""}`);
}
</script>

<template>
  <div class="main">
    <PureTableBar title="指标应用" :columns="columns" @refresh="onSearch">
      <template #buttons>
        <el-button type="primary" :icon="useRenderIcon(AddFill)" @click="openDialog()">新增应用</el-button>
      </template>
      <template v-slot="{ size, dynamicColumns }">
        <pure-table
          row-key="id"
          align-whole="center"
          table-layout="auto"
          :loading="loading"
          :size="size"
          :data="dataList"
          :columns="dynamicColumns"
          :pagination="pagination"
          @page-size-change="handleSizeChange"
          @page-current-change="handleCurrentChange"
        >
          <template #operation="{ row }">
            <el-button link type="primary" :icon="useRenderIcon(Key)" @click="showSecret(row)">凭证</el-button>
            <el-button link type="primary" :icon="useRenderIcon(EditPen)" @click="openDialog('编辑', row)">编辑</el-button>
            <el-button link type="warning" @click="regenerateSecret(row)">重置密钥</el-button>
            <el-popconfirm title="确认删除？" @confirm="handleDelete(row)">
              <template #reference>
                <el-button link type="danger" :icon="useRenderIcon(Delete)">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </pure-table>
      </template>
    </PureTableBar>
  </div>
</template>
