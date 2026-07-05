<script setup lang="ts">
import { ref } from "vue";
import type { FormInstance, FormRules } from "element-plus";

export type MetricsAppFormData = {
  id?: string;
  name: string;
  description: string;
  allowedOriginsText: string;
  status: number;
};

const props = defineProps<{
  formInline: MetricsAppFormData;
  isEdit?: boolean;
}>();

const ruleFormRef = ref<FormInstance>();
const formData = ref({ ...props.formInline });

const rules: FormRules = {
  name: [{ required: true, message: "请输入应用名称", trigger: "blur" }]
};

defineExpose({
  getRef: () => ruleFormRef.value,
  getFormData: () => formData.value
});
</script>

<template>
  <el-form ref="ruleFormRef" :model="formData" :rules="rules" label-width="100px">
    <el-form-item label="名称" prop="name">
      <el-input v-model="formData.name" placeholder="如 Firefly 官网" />
    </el-form-item>
    <el-form-item label="描述">
      <el-input v-model="formData.description" type="textarea" :rows="2" />
    </el-form-item>
    <el-form-item label="允许来源">
      <el-input
        v-model="formData.allowedOriginsText"
        type="textarea"
        :rows="3"
        placeholder="每行一个 Origin，如 https://firefly.bit-labs.cn；留空表示不限制"
      />
    </el-form-item>
    <el-form-item v-if="isEdit" label="状态">
      <el-radio-group v-model="formData.status">
        <el-radio :value="1">启用</el-radio>
        <el-radio :value="2">禁用</el-radio>
      </el-radio-group>
    </el-form-item>
  </el-form>
</template>
