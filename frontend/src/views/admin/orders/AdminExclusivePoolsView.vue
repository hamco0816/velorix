<template>
  <AppLayout wide>
    <div class="space-y-5">
      <!-- 选择独享池：inline label + 控件单行对齐 -->
      <div class="surface-card flex flex-wrap items-center gap-3 px-5 py-3">
        <label class="flex items-center gap-2 text-sm font-medium text-gray-700 dark:text-gray-300">
          <Icon name="badge" size="sm" class="text-gray-400" />
          {{ t('payment.admin.exclusivePools.poolGroup') }}
        </label>
        <div class="w-full sm:w-64">
          <Select v-model="selectedGroupID" :options="groupOptions" @change="loadInventoryAndSeats" />
        </div>
        <button @click="loadInventoryAndSeats" :disabled="!selectedGroupID || loading" class="btn btn-secondary btn-sm shrink-0" :title="t('common.refresh')">
          <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
        </button>
        <button @click="grantDialogOpen = true" :disabled="!selectedGroupID" class="btn btn-primary btn-sm ml-auto shrink-0 whitespace-nowrap">
          <Icon name="plus" size="sm" class="mr-1.5" />
          <span class="hidden sm:inline">{{ t('payment.admin.exclusivePools.grantButton') }}</span>
          <span class="sm:hidden">{{ t('common.create') }}</span>
        </button>
      </div>

      <!-- 库存看板：移动端 2 列 -->
      <div v-if="inventory" class="grid grid-cols-2 gap-3 sm:gap-4 lg:grid-cols-4">
        <div class="metric-card">
          <span class="metric-icon bg-violet-50 text-violet-600 dark:bg-violet-500/15 dark:text-violet-300">
            <Icon name="badge" size="sm" :stroke-width="1.75" />
          </span>
          <div class="min-w-0 flex-1">
            <p class="text-[11px] font-medium text-gray-500 dark:text-dark-400">{{ t('payment.admin.exclusivePools.statTotal') }}</p>
            <p class="mt-1 text-xl font-semibold leading-tight tabular-nums sm:text-[24px] text-gray-900 dark:text-white">{{ inventory.total }}</p>
          </div>
        </div>
        <div class="metric-card">
          <span class="metric-icon bg-emerald-50 text-emerald-600 dark:bg-emerald-500/15 dark:text-emerald-300">
            <Icon name="check" size="sm" :stroke-width="1.75" />
          </span>
          <div class="min-w-0 flex-1">
            <p class="text-[11px] font-medium text-gray-500 dark:text-dark-400" :title="t('payment.admin.exclusivePools.statSchedulableHint')">
              {{ t('payment.admin.exclusivePools.statAvailableNow') }}
            </p>
            <p class="mt-1 text-xl font-semibold leading-tight tabular-nums sm:text-[24px] text-emerald-600 dark:text-emerald-400">
              {{ typeof inventory.schedulable === 'number' ? inventory.schedulable : inventory.free }}
            </p>
            <p v-if="typeof inventory.schedulable === 'number' && inventory.schedulable !== inventory.free"
              class="mt-1 text-[11px] tabular-nums text-gray-500 dark:text-gray-400">
              {{ t('payment.admin.exclusivePools.statFreeRaw', { n: inventory.free }) }}
            </p>
          </div>
        </div>
        <div class="metric-card">
          <span class="metric-icon bg-sky-50 text-sky-600 dark:bg-sky-500/15 dark:text-sky-300">
            <Icon name="link" size="sm" :stroke-width="1.75" />
          </span>
          <div class="min-w-0 flex-1">
            <p class="text-[11px] font-medium text-gray-500 dark:text-dark-400">{{ t('payment.admin.exclusivePools.statUsed') }}</p>
            <p class="mt-1 text-xl font-semibold leading-tight tabular-nums sm:text-[24px] text-sky-600 dark:text-sky-400">{{ inventory.used }}</p>
          </div>
        </div>
        <!-- 7 天内到期：可点击筛选 seat 列表，运营预警入口 -->
        <button
          type="button"
          class="metric-card text-left"
          :class="expiringFilter ? 'ring-2 ring-amber-300 dark:ring-amber-500/50' : ''"
          :disabled="inventory.expiring_in_7 === 0"
          @click="toggleExpiringFilter"
        >
          <span class="metric-icon bg-amber-50 text-amber-600 dark:bg-amber-500/15 dark:text-amber-300">
            <Icon name="clock" size="sm" :stroke-width="1.75" />
          </span>
          <div class="min-w-0 flex-1">
            <p class="text-[11px] font-medium text-gray-500 dark:text-dark-400">{{ t('payment.admin.exclusivePools.statExpiring') }}</p>
            <p class="mt-1 text-xl font-semibold leading-tight tabular-nums sm:text-[24px] text-amber-600 dark:text-amber-400">{{ inventory.expiring_in_7 }}</p>
            <p v-if="inventory.expiring_in_7 > 0" class="mt-1 text-[11px] text-amber-700/80 dark:text-amber-300/70">
              {{ expiringFilter ? t('payment.admin.exclusivePools.statExpiringClickClear') : t('payment.admin.exclusivePools.statExpiringClickFilter') }}
            </p>
          </div>
        </button>
      </div>

      <!-- 7 天内到期筛选条 -->
      <div v-if="expiringFilter" class="flex items-center justify-between rounded-lg border border-amber-300 bg-amber-50 px-4 py-2 dark:border-amber-700 dark:bg-amber-900/20">
        <p class="text-sm text-amber-800 dark:text-amber-200">
          <Icon name="clock" size="sm" class="mr-1 inline-block" />
          {{ t('payment.admin.exclusivePools.expiringFilterActive') }}
        </p>
        <button class="text-xs font-semibold text-amber-700 hover:underline dark:text-amber-300" @click="expiringFilter = false">
          {{ t('payment.admin.exclusivePools.expiringFilterClear') }}
        </button>
      </div>

      <!-- seat 列表：顶部加搜索 + 状态筛选 -->
      <div class="surface-card overflow-hidden">
        <!-- 列表筛选条 -->
        <div class="flex flex-wrap items-center gap-2 border-b border-gray-200/60 px-5 py-3 dark:border-dark-700/60">
          <div class="relative w-full sm:w-64">
            <Icon name="search" size="sm" class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input
              v-model="seatSearch"
              type="text"
              :placeholder="t('payment.admin.exclusivePools.searchPlaceholder')"
              class="input w-full pl-9"
            />
          </div>
          <div class="w-[9.5rem]">
            <Select v-model="seatStatusFilter" :options="seatStatusOptions" />
          </div>
          <span class="ml-auto text-xs text-gray-500 dark:text-dark-400">
            {{ t('payment.admin.exclusivePools.totalCount', { n: filteredSeats.length }) }}
          </span>
        </div>

        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200/60 dark:divide-dark-700/60">
            <thead class="bg-gray-50/60 dark:bg-dark-800/60">
              <tr>
                <th class="px-4 py-3 text-left text-[13px] font-medium text-gray-500 dark:text-dark-400">ID</th>
                <th class="px-4 py-3 text-left text-[13px] font-medium text-gray-500 dark:text-dark-400">{{ t('payment.admin.exclusivePools.colUser') }}</th>
                <th class="px-4 py-3 text-left text-[13px] font-medium text-gray-500 dark:text-dark-400">{{ t('payment.admin.exclusivePools.colAccount') }}</th>
                <th class="px-4 py-3 text-left text-[13px] font-medium text-gray-500 dark:text-dark-400">{{ t('payment.admin.exclusivePools.colStatus') }}</th>
                <th class="px-4 py-3 text-left text-[13px] font-medium text-gray-500 dark:text-dark-400">{{ t('payment.admin.exclusivePools.colExpiresAt') }}</th>
                <th class="px-4 py-3 text-right text-[13px] font-medium text-gray-500 dark:text-dark-400">{{ t('payment.admin.exclusivePools.colActions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700/60">
              <tr v-for="seat in filteredSeats" :key="seat.id" class="text-sm transition-colors hover:bg-gray-50/60 dark:hover:bg-dark-800/40">
                <td class="px-4 py-3 font-mono text-gray-500 dark:text-gray-400">#{{ seat.id }}</td>
                <td class="px-4 py-3">
                  <div class="flex flex-col">
                    <span class="font-medium text-gray-900 dark:text-white">{{ userLabel(seat.user_id) }}</span>
                    <span class="font-mono text-[11px] text-gray-400">#{{ seat.user_id }}</span>
                  </div>
                </td>
                <td class="px-4 py-3">
                  <div class="flex flex-col">
                    <span class="text-gray-900 dark:text-gray-200">{{ accountLabel(seat.account_id) }}</span>
                    <span class="font-mono text-[11px] text-gray-400">#{{ seat.account_id }}</span>
                  </div>
                </td>
                <td class="px-4 py-3">
                  <span class="inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium" :class="statusPillClass(seat.status)">
                    <span class="relative flex h-1.5 w-1.5">
                      <span
                        v-if="seat.status === 'active'"
                        class="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-70"
                      ></span>
                      <span
                        class="relative inline-flex h-1.5 w-1.5 rounded-full"
                        :class="statusDotClass(seat.status)"
                      ></span>
                    </span>
                    {{ t(`exclusiveSeats.status.${seat.status}`) }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <div class="flex flex-col items-start gap-1">
                    <span class="text-sm tabular-nums" :class="expiresColor(seat)">{{ formatDate(seat.expires_at) }}</span>
                    <!-- 倒计时 chip：< 3 天 rose / 3-7 天 amber，更醒目 -->
                    <span v-if="seat.status === 'active' && expiryDays(seat) !== null" class="inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[11px] font-medium" :class="expiryChipClass(seat)">
                      <Icon name="clock" size="xs" />
                      {{ expiryChipLabel(seat) }}
                    </span>
                  </div>
                </td>
                <td class="px-4 py-3 text-right">
                  <div class="flex justify-end gap-1">
                    <button v-if="seat.status === 'active'" @click="openExtend(seat)" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-gray-600 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-dark-700 dark:hover:text-gray-100" :title="t('payment.admin.exclusivePools.actExtend')">
                      <Icon name="calendar" size="xs" />
                      {{ t('payment.admin.exclusivePools.actExtend') }}
                    </button>
                    <button v-if="seat.status === 'active'" @click="openSwap(seat)" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-sky-600 transition-colors hover:bg-sky-50 dark:text-sky-400 dark:hover:bg-sky-500/10" :title="t('payment.admin.exclusivePools.actSwap')">
                      <Icon name="refresh" size="xs" />
                      {{ t('payment.admin.exclusivePools.actSwap') }}
                    </button>
                    <button v-if="seat.status === 'active'" @click="openRelease(seat)" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-rose-600 transition-colors hover:bg-rose-50 dark:text-rose-400 dark:hover:bg-rose-500/10" :title="t('payment.admin.exclusivePools.actRelease')">
                      <Icon name="trash" size="xs" />
                      {{ t('payment.admin.exclusivePools.actRelease') }}
                    </button>
                  </div>
                </td>
              </tr>
              <tr v-if="filteredSeats.length === 0">
                <td colspan="6" class="py-16">
                  <div class="flex flex-col items-center gap-3">
                    <div class="flex h-16 w-16 items-center justify-center rounded-2xl bg-gray-100 text-gray-400 dark:bg-dark-700 dark:text-dark-500">
                      <Icon name="inbox" size="lg" />
                    </div>
                    <p class="text-sm font-semibold text-gray-900 dark:text-gray-100">
                      {{ expiringFilter ? t('payment.admin.exclusivePools.emptyExpiring') : (seatSearch || seatStatusFilter ? t('payment.admin.exclusivePools.emptyFiltered') : t('payment.admin.exclusivePools.empty')) }}
                    </p>
                    <p v-if="!filteredSeats.length && (seatSearch || seatStatusFilter || expiringFilter)" class="text-xs text-gray-500 dark:text-dark-400">
                      {{ t('payment.admin.exclusivePools.emptyFilteredHint') }}
                    </p>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- 赠送 dialog -->
    <BaseDialog :show="grantDialogOpen" :title="t('payment.admin.exclusivePools.grantTitle')" @close="grantDialogOpen = false">
      <div class="space-y-3">
        <div>
          <label class="input-label">{{ t('payment.admin.exclusivePools.targetUserID') }}</label>
          <!-- 用户搜索：输入 ID/邮箱/用户名实时匹配；选中后下方显示标识防送错 -->
          <div class="relative mt-1">
            <input
              v-model="userSearchQuery"
              type="text"
              :placeholder="t('payment.admin.exclusivePools.userSearchPlaceholder')"
              class="input w-full"
              @input="onUserSearchInput"
              @focus="userSearchFocused = true"
              @blur="onUserSearchBlur"
            />
            <div
              v-if="userSearchFocused && userSearchResults.length > 0"
              class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md border border-gray-200 bg-white shadow-lg dark:border-dark-600 dark:bg-dark-800"
            >
              <button
                v-for="user in userSearchResults"
                :key="user.id"
                type="button"
                class="block w-full px-3 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-dark-700"
                @mousedown.prevent="selectUser(user)"
              >
                <span class="font-mono text-gray-500">#{{ user.id }}</span>
                <span class="ml-2 font-medium text-gray-900 dark:text-white">{{ user.username || user.email }}</span>
                <span v-if="user.username && user.email" class="ml-2 text-xs text-gray-500">{{ user.email }}</span>
              </button>
            </div>
          </div>
          <p v-if="selectedUser" class="mt-1.5 text-xs">
            <span class="inline-flex items-center gap-1 rounded-md bg-emerald-50 px-2 py-0.5 font-medium text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300">
              <Icon name="check" size="xs" />
              {{ t('payment.admin.exclusivePools.userSearchSelected') }}: #{{ selectedUser.id }} {{ selectedUser.username || selectedUser.email }}
            </span>
          </p>
          <p v-else-if="userSearchQuery && userSearchResults.length === 0 && !userSearchLoading" class="mt-1 text-xs text-amber-600 dark:text-amber-400">
            {{ t('payment.admin.exclusivePools.userSearchNoMatch') }}
          </p>
        </div>
        <div>
          <label class="input-label">{{ t('payment.admin.exclusivePools.targetPlanID') }}</label>
          <Select v-model="grantForm.plan_id" :options="grantPlanOptions" class="mt-1 w-full" />
          <p v-if="grantPlanOptions.length === 0" class="mt-1 text-xs text-amber-600 dark:text-amber-400">
            {{ t('payment.admin.exclusivePools.noPlansForPool') }}
          </p>
        </div>
        <div>
          <label class="input-label">{{ t('payment.admin.exclusivePools.validityDays') }}</label>
          <input v-model.number="grantForm.validity_days" type="number" min="1" class="input mt-1 w-full" />
          <!-- 快捷按钮：常用 7/30/90/180/365 天，省去手动输入 -->
          <div class="mt-2 flex flex-wrap gap-1.5">
            <button
              v-for="d in [7, 30, 90, 180, 365]"
              :key="d"
              type="button"
              class="rounded-md px-2 py-0.5 text-xs font-medium transition-colors"
              :class="grantForm.validity_days === d
                ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-300 dark:hover:bg-dark-600'"
              @click="grantForm.validity_days = d"
            >
              {{ d }}{{ t('payment.admin.daySuffix') }}
            </button>
          </div>
        </div>
        <div>
          <label class="input-label">{{ t('payment.admin.exclusivePools.notes') }}</label>
          <textarea v-model="grantForm.notes" rows="2" class="input mt-1 w-full" />
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="grantDialogOpen = false">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" :disabled="actionLoading" @click="confirmGrant">
            {{ actionLoading ? t('common.processing') : t('payment.admin.exclusivePools.grantButton') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- 释放 dialog -->
    <BaseDialog :show="!!releaseTarget" :title="t('payment.admin.exclusivePools.releaseTitle')" width="narrow" @close="releaseTarget = null">
      <div class="space-y-3">
        <p class="text-sm text-gray-600 dark:text-gray-300">
          {{ t('payment.admin.exclusivePools.releaseHint', { id: releaseTarget?.id }) }}
        </p>
        <div>
          <label class="input-label">{{ t('payment.admin.exclusivePools.reason') }}</label>
          <textarea v-model="releaseReason" rows="2" class="input mt-1 w-full" />
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="releaseTarget = null">{{ t('common.cancel') }}</button>
          <button class="btn btn-danger" :disabled="actionLoading" @click="confirmRelease">
            {{ actionLoading ? t('common.processing') : t('payment.admin.exclusivePools.actRelease') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- 延期 dialog -->
    <BaseDialog :show="!!extendTarget" :title="t('payment.admin.exclusivePools.extendTitle')" width="narrow" @close="extendTarget = null">
      <div class="space-y-3">
        <div>
          <label class="input-label">{{ t('payment.admin.exclusivePools.extendDays') }}</label>
          <input v-model.number="extendDays" type="number" class="input mt-1 w-full" />
          <p class="mt-1 text-xs text-gray-500">{{ t('payment.admin.exclusivePools.extendHint') }}</p>
          <!-- 快捷延期：+7/+30/+90/+180/+365 / -7 缩减 -->
          <div class="mt-2 flex flex-wrap gap-1.5">
            <button
              v-for="d in [7, 30, 90, 180, 365]"
              :key="d"
              type="button"
              class="rounded-md px-2 py-0.5 text-xs font-medium transition-colors"
              :class="extendDays === d
                ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-300 dark:hover:bg-dark-600'"
              @click="extendDays = d"
            >
              +{{ d }}{{ t('payment.admin.daySuffix') }}
            </button>
          </div>
        </div>
        <!-- 预览：当前到期日 + 输入天数 = 新到期日，让管理员所见即所得 -->
        <div v-if="extendTarget" class="rounded-lg bg-gray-50 p-3 text-xs dark:bg-dark-800">
          <div class="flex justify-between py-0.5">
            <span class="text-gray-500">{{ t('payment.admin.exclusivePools.extendPreviewCurrent') }}</span>
            <span class="font-medium text-gray-700 dark:text-gray-300">{{ formatDate(extendTarget.expires_at) }}</span>
          </div>
          <div class="flex justify-between py-0.5">
            <span class="text-gray-500">{{ t('payment.admin.exclusivePools.extendPreviewDelta') }}</span>
            <span :class="['font-medium', extendDays > 0 ? 'text-emerald-600' : extendDays < 0 ? 'text-red-600' : 'text-gray-500']">
              {{ extendDays > 0 ? `+${extendDays}` : extendDays }} {{ t('payment.admin.exclusivePools.daysSuffix') }}
            </span>
          </div>
          <div class="mt-1 flex justify-between border-t border-gray-200 pt-1 dark:border-dark-700">
            <span class="font-semibold text-gray-700 dark:text-gray-200">{{ t('payment.admin.exclusivePools.extendPreviewNew') }}</span>
            <span :class="['font-semibold', extendPreviewInvalid ? 'text-red-600' : 'text-primary-600 dark:text-primary-400']">
              {{ extendPreviewDate }}
            </span>
          </div>
          <p v-if="extendPreviewInvalid" class="mt-1 text-[11px] text-red-600">
            {{ t('payment.admin.exclusivePools.extendPreviewInvalidPast') }}
          </p>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="extendTarget = null">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" :disabled="actionLoading || extendPreviewInvalid || extendDays === 0" @click="confirmExtend">
            {{ actionLoading ? t('common.processing') : t('payment.admin.exclusivePools.actExtend') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- 换号确认 dialog（替代之前的浏览器 confirm()），明确"释放旧账号 + 分配新空闲号" -->
    <BaseDialog :show="!!swapTarget" :title="t('payment.admin.exclusivePools.swapTitle')" width="narrow" @close="swapTarget = null">
      <div v-if="swapTarget" class="space-y-3">
        <p class="text-sm text-gray-700 dark:text-gray-200">
          {{ t('payment.admin.exclusivePools.swapDialogHint') }}
        </p>
        <div class="rounded-lg bg-gray-50 p-3 text-xs dark:bg-dark-800">
          <div class="py-0.5"><span class="text-gray-500">Seat:</span> <span class="font-mono">#{{ swapTarget.id }}</span></div>
          <div class="py-0.5"><span class="text-gray-500">{{ t('payment.admin.exclusivePools.colUser') }}:</span> <span class="font-medium">{{ userLabel(swapTarget.user_id) }} (#{{ swapTarget.user_id }})</span></div>
          <div class="py-0.5"><span class="text-gray-500">{{ t('payment.admin.exclusivePools.swapDialogOldAccount') }}:</span> <span class="font-medium">{{ accountLabel(swapTarget.account_id) }} (#{{ swapTarget.account_id }})</span></div>
        </div>
        <p class="flex items-start gap-1.5 text-xs text-amber-700 dark:text-amber-400">
          <Icon name="exclamationTriangle" size="xs" class="mt-0.5 flex-shrink-0" />
          <span>{{ t('payment.admin.exclusivePools.swapDialogWarn') }}</span>
        </p>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="swapTarget = null">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" :disabled="actionLoading" @click="confirmSwap">
            {{ actionLoading ? t('common.processing') : t('payment.admin.exclusivePools.actSwap') }}
          </button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminPaymentAPI } from '@/api/admin/payment'
import adminAPI from '@/api/admin'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { AdminExclusiveSeat, ExclusivePoolInventory, SubscriptionPlan } from '@/types/payment'
import type { AdminGroup } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const actionLoading = ref(false)
const groups = ref<AdminGroup[]>([])
const selectedGroupID = ref<number | null>(null)
const inventory = ref<ExclusivePoolInventory | null>(null)
const seats = ref<AdminExclusiveSeat[]>([])

const grantDialogOpen = ref(false)
const grantForm = ref({ user_id: 0, plan_id: 0, validity_days: 30, notes: '' })

// 当前所选独享池下的套餐列表（exclusive 类型 + 同 group_id），用于赠送 dialog 的下拉
const allPlans = ref<SubscriptionPlan[]>([])
const grantPlanOptions = computed(() => {
  if (!selectedGroupID.value) return []
  return allPlans.value
    .filter(p => p.kind === 'exclusive' && p.group_id === selectedGroupID.value)
    .map(p => ({ value: p.id, label: `#${p.id} · ${p.name} — ¥${p.price} / ${p.validity_days}${p.validity_unit}` }))
})

// 用户搜索（输入 ID/邮箱/用户名 实时匹配），防止管理员只看 ID 送错人
const userSearchQuery = ref('')
const userSearchResults = ref<Array<{ id: number; email?: string; username?: string }>>([])
const userSearchLoading = ref(false)
const userSearchFocused = ref(false)
const selectedUser = ref<{ id: number; email?: string; username?: string } | null>(null)
let userSearchTimer: ReturnType<typeof setTimeout> | null = null
let userSearchSeq = 0
function onUserSearchInput() {
  // 输入即清空已选用户（重新选）+ 防抖 300ms 调 admin users API
  selectedUser.value = null
  grantForm.value.user_id = 0
  if (userSearchTimer) clearTimeout(userSearchTimer)
  const q = userSearchQuery.value.trim()
  if (!q) {
    userSearchResults.value = []
    return
  }
  userSearchLoading.value = true
  const seq = ++userSearchSeq
  userSearchTimer = setTimeout(async () => {
    try {
      const res = await adminAPI.users.list(1, 8, { search: q })
      if (seq === userSearchSeq) {
        userSearchResults.value = (res?.items || []).map((u) => ({
          id: u.id,
          email: u.email,
          username: u.username,
        }))
      }
    } catch {
      if (seq === userSearchSeq) userSearchResults.value = []
    } finally {
      if (seq === userSearchSeq) userSearchLoading.value = false
    }
  }, 300)
}
function selectUser(user: { id: number; email?: string; username?: string }) {
  selectedUser.value = user
  grantForm.value.user_id = user.id
  userSearchQuery.value = `#${user.id} ${user.username || user.email || ''}`.trim()
  userSearchResults.value = []
  userSearchFocused.value = false
}
function onUserSearchBlur() {
  // 延迟关闭下拉，让 mousedown 选项点击事件先触发
  setTimeout(() => { userSearchFocused.value = false }, 200)
}
function resetGrantForm() {
  grantForm.value = { user_id: 0, plan_id: 0, validity_days: 30, notes: '' }
  selectedUser.value = null
  userSearchQuery.value = ''
  userSearchResults.value = []
}

const releaseTarget = ref<AdminExclusiveSeat | null>(null)
const releaseReason = ref('')

const extendTarget = ref<AdminExclusiveSeat | null>(null)
const extendDays = ref(7)

// 换号 dialog 目标（替代原 confirm()）
const swapTarget = ref<AdminExclusiveSeat | null>(null)

// 7 天内到期筛选开关：点击库存看板的"到期"卡片切换
const expiringFilter = ref(false)

// 用户/账号信息缓存：seat 列表只有 id，这里按 id 单独拉一次显示名称，避免送错人/认错号
const userInfoCache = ref<Map<number, { id: number; username?: string; email?: string }>>(new Map())
const accountInfoCache = ref<Map<number, { id: number; name?: string }>>(new Map())

const groupOptions = computed(() =>
  groups.value
    .filter(g => g.subscription_type === 'subscription')
    .map(g => ({ value: g.id, label: `${g.name} — ${g.platform}` })),
)

async function loadGroups() {
  try {
    groups.value = await adminAPI.groups.getAll()
  } catch { /* ignore */ }
}

// 拉取所有套餐用于赠送 dialog 的下拉选择；按当前 selectedGroupID 二次过滤。
async function loadPlans() {
  try {
    const res = await adminPaymentAPI.getPlans()
    allPlans.value = res.data
  } catch { /* ignore */ }
}

async function loadInventoryAndSeats() {
  if (!selectedGroupID.value) return
  loading.value = true
  try {
    const [invRes, seatsRes] = await Promise.all([
      adminPaymentAPI.getPoolInventory(selectedGroupID.value),
      adminPaymentAPI.listSeats({ group_id: selectedGroupID.value, limit: 200 }),
    ])
    inventory.value = invRes.data
    seats.value = seatsRes.data.items || []
    // 后台异步补全显示名称（不阻塞主表渲染），失败回退为 user-#N / acc-#N
    void hydrateLabels(seats.value)
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

// 按 seat 收集 user_id / account_id，未缓存的逐个拉详情；并发上限通过 await 顺序自然限制
async function hydrateLabels(list: AdminExclusiveSeat[]) {
  const userIds = new Set<number>()
  const accIds = new Set<number>()
  for (const s of list) {
    if (s.user_id && !userInfoCache.value.has(s.user_id)) userIds.add(s.user_id)
    if (s.account_id && !accountInfoCache.value.has(s.account_id)) accIds.add(s.account_id)
  }
  const userPromises = Array.from(userIds).map(async (id) => {
    try {
      const u = await adminAPI.users.getById(id)
      userInfoCache.value.set(id, { id, username: u.username, email: u.email })
    } catch {
      userInfoCache.value.set(id, { id })
    }
  })
  const accPromises = Array.from(accIds).map(async (id) => {
    try {
      const a = await adminAPI.accounts.getById(id)
      accountInfoCache.value.set(id, { id, name: a.name })
    } catch {
      accountInfoCache.value.set(id, { id })
    }
  })
  await Promise.all([...userPromises, ...accPromises])
  // 触发响应式更新（Map.set 已是 reactive，但显式 trigger 确保模板重渲染）
  userInfoCache.value = new Map(userInfoCache.value)
  accountInfoCache.value = new Map(accountInfoCache.value)
}

function userLabel(id: number): string {
  const u = userInfoCache.value.get(id)
  if (!u) return `user-#${id}`
  return u.username || u.email || `user-#${id}`
}

function accountLabel(id: number): string {
  const a = accountInfoCache.value.get(id)
  if (!a) return `acc-#${id}`
  return a.name || `acc-#${id}`
}

// 距离到期的剩余毫秒；负数表示已过期
function msUntilExpiry(seat: AdminExclusiveSeat): number {
  if (!seat.expires_at) return Number.POSITIVE_INFINITY
  const d = new Date(seat.expires_at)
  if (isNaN(d.getTime())) return Number.POSITIVE_INFINITY
  return d.getTime() - Date.now()
}

// 到期时间颜色：<0 红、<3天 红、<7天 黄、其他默认
function expiresColor(seat: AdminExclusiveSeat): string {
  const ms = msUntilExpiry(seat)
  const day = 24 * 60 * 60 * 1000
  if (ms < 0) return 'text-red-600 dark:text-red-400 font-medium'
  if (ms < 3 * day) return 'text-red-600 dark:text-red-400'
  if (ms < 7 * day) return 'text-amber-600 dark:text-amber-400'
  return 'text-gray-700 dark:text-gray-300'
}

function isExpiringSoon(seat: AdminExclusiveSeat): boolean {
  const ms = msUntilExpiry(seat)
  return ms >= 0 && ms < 7 * 24 * 60 * 60 * 1000
}

// 搜索（user 名称 / user id / account 名称 / account id）+ 状态筛选 + 到期筛选
const seatSearch = ref('')
const seatStatusFilter = ref<string>('')
const seatStatusOptions = computed(() => [
  { value: '', label: t('payment.admin.exclusivePools.allStatuses') },
  { value: 'active', label: t('exclusiveSeats.status.active') },
  { value: 'expired', label: t('exclusiveSeats.status.expired') },
  { value: 'refunded', label: t('exclusiveSeats.status.refunded') }
])

const filteredSeats = computed(() => {
  let list = seats.value
  if (expiringFilter.value) {
    list = list.filter((s) => s.status === 'active' && isExpiringSoon(s))
  }
  if (seatStatusFilter.value) {
    list = list.filter((s) => s.status === seatStatusFilter.value)
  }
  const q = seatSearch.value.trim().toLowerCase()
  if (q) {
    list = list.filter((s) => {
      const userText = `${userLabel(s.user_id)} ${s.user_id}`.toLowerCase()
      const accountText = `${accountLabel(s.account_id)} ${s.account_id}`.toLowerCase()
      return userText.includes(q) || accountText.includes(q)
    })
  }
  return list
})

function toggleExpiringFilter() {
  if (inventory.value && inventory.value.expiring_in_7 === 0) return
  expiringFilter.value = !expiringFilter.value
}

// 延期 dialog 预览：当前到期日 + extendDays 后的新到期日，以及"提前结束导致到期日为过去"的非法态
const extendPreviewDate = computed<string>(() => {
  if (!extendTarget.value || !extendTarget.value.expires_at) return '-'
  const base = new Date(extendTarget.value.expires_at)
  if (isNaN(base.getTime())) return '-'
  const next = new Date(base.getTime() + (extendDays.value || 0) * 24 * 60 * 60 * 1000)
  return next.toLocaleString()
})
const extendPreviewInvalid = computed<boolean>(() => {
  if (!extendTarget.value || !extendTarget.value.expires_at) return false
  if (extendDays.value >= 0) return false
  const base = new Date(extendTarget.value.expires_at)
  if (isNaN(base.getTime())) return false
  const next = base.getTime() + extendDays.value * 24 * 60 * 60 * 1000
  return next <= Date.now()
})

async function confirmGrant() {
  if (!grantForm.value.user_id || !grantForm.value.plan_id) {
    appStore.showError('user_id / plan_id required')
    return
  }
  if (!selectedGroupID.value) {
    appStore.showError('select an exclusive pool first')
    return
  }
  // 前端二次保护：只允许选当前池下的 exclusive 套餐；同时把 group_id 透传给后端做权威校验
  const matched = grantPlanOptions.value.some(opt => opt.value === grantForm.value.plan_id)
  if (!matched) {
    appStore.showError(t('payment.admin.exclusivePools.planMustBelongToPool'))
    return
  }
  actionLoading.value = true
  try {
    await adminPaymentAPI.grantSeat({
      ...grantForm.value,
      group_id: selectedGroupID.value,
    })
    appStore.showSuccess(t('common.success'))
    grantDialogOpen.value = false
    resetGrantForm()
    await loadInventoryAndSeats()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

function openRelease(seat: AdminExclusiveSeat) {
  releaseTarget.value = seat
  releaseReason.value = ''
}

async function confirmRelease() {
  if (!releaseTarget.value) return
  actionLoading.value = true
  try {
    await adminPaymentAPI.releaseSeat(releaseTarget.value.id, releaseReason.value)
    appStore.showSuccess(t('common.success'))
    releaseTarget.value = null
    await loadInventoryAndSeats()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

function openSwap(seat: AdminExclusiveSeat) {
  swapTarget.value = seat
}

async function confirmSwap() {
  if (!swapTarget.value) return
  const seatId = swapTarget.value.id
  actionLoading.value = true
  try {
    await adminPaymentAPI.swapSeatAccount(seatId)
    appStore.showSuccess(t('common.success'))
    swapTarget.value = null
    await loadInventoryAndSeats()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

function openExtend(seat: AdminExclusiveSeat) {
  extendTarget.value = seat
  extendDays.value = 7
}

async function confirmExtend() {
  if (!extendTarget.value || extendDays.value === 0) return
  actionLoading.value = true
  try {
    await adminPaymentAPI.extendSeat(extendTarget.value.id, extendDays.value)
    appStore.showSuccess(t('common.success'))
    extendTarget.value = null
    await loadInventoryAndSeats()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

function statusPillClass(status: string): string {
  switch (status) {
    case 'active': return 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300'
    case 'expired': return 'bg-amber-50 text-amber-700 dark:bg-amber-500/15 dark:text-amber-300'
    case 'refunded': return 'bg-violet-50 text-violet-700 dark:bg-violet-500/15 dark:text-violet-300'
    default: return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300'
  }
}

function statusDotClass(status: string): string {
  switch (status) {
    case 'active': return 'bg-emerald-500'
    case 'expired': return 'bg-amber-500'
    case 'refunded': return 'bg-violet-500'
    default: return 'bg-gray-400'
  }
}

// 到期倒计时（天数）：null = 没有到期时间
function expiryDays(seat: AdminExclusiveSeat): number | null {
  const ms = msUntilExpiry(seat)
  if (!isFinite(ms)) return null
  return Math.ceil(ms / (24 * 60 * 60 * 1000))
}

function expiryChipClass(seat: AdminExclusiveSeat): string {
  const days = expiryDays(seat)
  if (days === null || days >= 7) return ''
  if (days < 0) return 'bg-rose-50 text-rose-700 dark:bg-rose-500/15 dark:text-rose-300'
  if (days < 3) return 'bg-rose-50 text-rose-700 dark:bg-rose-500/15 dark:text-rose-300'
  return 'bg-amber-50 text-amber-700 dark:bg-amber-500/15 dark:text-amber-300'
}

function expiryChipLabel(seat: AdminExclusiveSeat): string {
  const days = expiryDays(seat)
  if (days === null || days >= 7) return ''
  if (days < 0) return t('payment.admin.exclusivePools.alreadyExpired')
  if (days === 0) return t('payment.admin.exclusivePools.expiresToday')
  return t('payment.admin.exclusivePools.expiresInDays', { n: days })
}

function formatDate(s: string): string {
  if (!s) return '-'
  const d = new Date(s)
  if (isNaN(d.getTime())) return s
  return d.toLocaleString()
}

onMounted(() => {
  loadGroups()
  loadPlans()
})
</script>

<style scoped>
.metric-card {
  @apply flex items-start gap-3 rounded-2xl border border-gray-200/70 bg-white p-4 shadow-[0_1px_2px_rgba(15,23,42,0.04)] transition-all duration-200;
  @apply dark:border-dark-700/60 dark:bg-dark-800/40;
  @apply hover:-translate-y-0.5 hover:border-gray-300 hover:shadow-[0_4px_16px_rgba(15,23,42,0.06)] dark:hover:border-dark-600;
}

.metric-card[disabled] {
  @apply pointer-events-none opacity-60;
}

.metric-icon {
  @apply flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl;
}
</style>
