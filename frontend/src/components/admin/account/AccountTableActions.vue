<template>
  <!-- 操作按钮组：靠右排列；筛选条已占满上一行，这里只放操作按钮 -->
  <div class="flex flex-wrap items-center justify-end gap-2">
    <slot name="before"></slot>
    <button @click="$emit('refresh')" :disabled="loading" class="btn btn-secondary shrink-0" :title="t('common.refresh')">
      <Icon name="refresh" size="md" :class="[loading ? 'animate-spin' : '']" />
    </button>
    <slot name="after"></slot>
    <button @click="$emit('sync')" class="btn btn-secondary shrink-0 whitespace-nowrap">
      <span class="hidden sm:inline">{{ t('admin.accounts.syncFromCrs') }}</span>
      <span class="sm:hidden">{{ t('admin.accounts.syncFromCrsShort') }}</span>
    </button>
    <slot name="beforeCreate"></slot>
    <button @click="$emit('create')" class="btn btn-primary shrink-0 whitespace-nowrap">
      <Icon name="plus" size="sm" class="mr-1.5" />
      <span class="hidden sm:inline">{{ t('admin.accounts.createAccount') }}</span>
      <span class="sm:hidden">{{ t('common.create') }}</span>
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
