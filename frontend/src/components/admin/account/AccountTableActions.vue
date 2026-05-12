<template>
  <!-- 操作按钮组：靠右排列；筛选条已占满上一行，这里只放操作按钮 -->
  <div class="flex flex-wrap items-center justify-end gap-2">
    <slot name="before"></slot>
    <button @click="$emit('refresh')" :disabled="loading" class="btn btn-secondary" :title="t('common.refresh')">
      <Icon name="refresh" size="md" :class="[loading ? 'animate-spin' : '']" />
    </button>
    <slot name="after"></slot>
    <button @click="$emit('sync')" class="btn btn-secondary">{{ t('admin.accounts.syncFromCrs') }}</button>
    <slot name="beforeCreate"></slot>
    <button @click="$emit('create')" class="btn btn-primary">
      <Icon name="plus" size="sm" class="mr-1.5" />
      {{ t('admin.accounts.createAccount') }}
    </button>
    <slot name="afterCreate"></slot>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

defineProps(['loading'])
defineEmits(['refresh', 'sync', 'create'])

const { t } = useI18n()
</script>
