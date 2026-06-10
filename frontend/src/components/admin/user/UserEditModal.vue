<template>
  <BaseDialog
    :show="show"
    :title="t('admin.users.editUser')"
    width="normal"
    @close="$emit('close')"
  >
    <form v-if="user" id="edit-user-form" @submit.prevent="handleUpdateUser" class="space-y-5">
      <div>
        <label class="input-label">{{ t('admin.users.email') }}</label>
        <input v-model="form.email" type="email" class="input" />
      </div>
      <div>
        <label class="input-label">{{ t('admin.users.password') }}</label>
        <div class="flex gap-2">
          <div class="relative flex-1">
            <input v-model="form.password" type="text" class="input pr-10" :placeholder="t('admin.users.enterNewPassword')" />
            <button v-if="form.password" type="button" @click="copyPassword" class="absolute right-2 top-1/2 -translate-y-1/2 rounded-lg p-1 transition-colors hover:bg-gray-100 dark:hover:bg-dark-700" :class="passwordCopied ? 'text-green-500' : 'text-gray-400'">
              <Icon v-if="passwordCopied" name="check" size="sm" :stroke-width="2" />
              <Icon v-else name="clipboard" size="sm" />
            </button>
          </div>
          <button type="button" @click="generatePassword" class="btn btn-secondary px-3">
            <Icon name="refresh" size="md" />
          </button>
        </div>
      </div>
      <div>
        <label class="input-label">{{ t('admin.users.username') }}</label>
        <input v-model="form.username" type="text" class="input" />
      </div>
      <div>
        <label class="input-label">{{ t('admin.users.notes') }}</label>
        <textarea v-model="form.notes" rows="3" class="input"></textarea>
      </div>
      <div>
        <label class="input-label">{{ t('admin.users.columns.concurrency') }}</label>
        <input v-model.number="form.concurrency" type="number" class="input" />
      </div>
      <div>
        <label class="input-label">{{ t('admin.users.form.rpmLimit') }}</label>
        <input
          v-model.number="form.rpm_limit"
          type="number"
          min="0"
          step="1"
          class="input"
          :placeholder="t('admin.users.form.rpmLimitPlaceholder')"
        />
        <p class="input-hint">{{ t('admin.users.form.rpmLimitHint') }}</p>
      </div>
      <UserAttributeForm v-model="form.customAttributes" :user-id="user?.id" />
    </form>
    <template #footer>
      <div class="flex justify-end gap-3">
        <button @click="$emit('close')" type="button" class="btn btn-secondary">{{ t('common.cancel') }}</button>
        <button type="submit" form="edit-user-form" :disabled="submitting" class="btn btn-primary">
          {{ submitting ? t('admin.users.updating') : t('common.update') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useClipboard } from '@/composables/useClipboard'
import { adminAPI } from '@/api/admin'
import type { AdminUser, UserAttributeValuesMap } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import UserAttributeForm from '@/components/user/UserAttributeForm.vue'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{ show: boolean, user: AdminUser | null }>()
const emit = defineEmits(['close', 'success'])
const { t } = useI18n(); const appStore = useAppStore(); const { copyToClipboard } = useClipboard()

const submitting = ref(false); const passwordCopied = ref(false)
const form = reactive({ email: '', password: '', username: '', notes: '', concurrency: 1, rpm_limit: 0, customAttributes: {} as UserAttributeValuesMap })

watch(() => props.user, (u) => {
  if (u) {
    Object.assign(form, { email: u.email, password: '', username: u.username || '', notes: u.notes || '', concurrency: u.concurrency, rpm_limit: u.rpm_limit ?? 0, customAttributes: {} })
    passwordCopied.value = false
  }
}, { immediate: true })

const generatePassword = () => {
  const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789!@#$%^&*'
  let p = ''; for (let i = 0; i < 16; i++) p += chars.charAt(Math.floor(Math.random() * chars.length))
  form.password = p
}
const copyPassword = async () => {
  if (form.password && await copyToClipboard(form.password, t('admin.users.passwordCopied'))) {
    passwordCopied.value = true; setTimeout(() => passwordCopied.value = false, 2000)
  }
}
const handleUpdateUser = async () => {
  if (!props.user) return
  if (!form.email.trim()) {
    appStore.showError(t('admin.users.emailRequired'))
    return
  }
  if (form.concurrency < 1) {
    appStore.showError(t('admin.users.concurrencyMin'))
    return
  }
  submitting.value = true
  try {
    const data: any = { email: form.email, username: form.username, notes: form.notes, concurrency: form.concurrency, rpm_limit: form.rpm_limit }
    if (form.password.trim()) data.password = form.password.trim()
    await adminAPI.users.update(props.user.id, data)
    if (Object.keys(form.customAttributes).length > 0) await adminAPI.userAttributes.updateUserAttributeValues(props.user.id, form.customAttributes)
    appStore.showSuccess(t('admin.users.userUpdated'))
    emit('success'); emit('close')
  } catch (e: any) {
    appStore.showError(e.response?.data?.detail || t('admin.users.failedToUpdate'))
  } finally { submitting.value = false }
}
</script>
