<template>
  <span
    class="inline-flex shrink-0 items-center justify-center overflow-hidden rounded-full bg-white shadow-sm ring-1 ring-gray-200 dark:bg-white"
    :style="{ width: resolvedSize, height: resolvedSize }"
    :aria-label="altText"
  >
    <LibraryIcon
      v-if="iconDefinition"
      :icon="iconDefinition"
      :color="brandColor"
      class="h-full w-full"
    />
    <Icon v-else name="chat" size="xs" class="text-gray-500" />
  </span>
</template>

<script setup lang="ts">
import { computed, defineComponent, h } from 'vue'
import type { VNode } from 'vue'
import QqCircleFilled from '@ant-design/icons-svg/es/asn/QqCircleFilled'
import WechatFilled from '@ant-design/icons-svg/es/asn/WechatFilled'
import type { AbstractNode, IconDefinition } from '@ant-design/icons-svg/es/types'
import Icon from '@/components/icons/Icon.vue'
import { normalizeContactMethodType } from '@/utils/contactMethods'

function renderIconNode(node: AbstractNode): VNode {
  const attrs = { ...node.attrs }
  if (node.tag === 'svg') {
    attrs.width = '1em'
    attrs.height = '1em'
    attrs.fill = 'currentColor'
    attrs['aria-hidden'] = 'true'
  }
  if (node.tag === 'path' && !attrs.fill) {
    attrs.fill = 'currentColor'
  }
  return h(node.tag, attrs, node.children?.map((child) => renderIconNode(child)))
}

const LibraryIcon = defineComponent({
  props: {
    icon: {
      type: Object as () => IconDefinition,
      required: true,
    },
    color: {
      type: String,
      required: true,
    },
  },
  setup(componentProps) {
    return () => {
      const node =
        typeof componentProps.icon.icon === 'function'
          ? componentProps.icon.icon(componentProps.color, componentProps.color)
          : componentProps.icon.icon
      return h(
        'span',
        {
          class: 'inline-flex h-full w-full items-center justify-center',
          style: { color: componentProps.color },
        },
        [renderIconNode(node)],
      )
    }
  },
})

const props = withDefaults(defineProps<{
  type: string
  size?: string | number
}>(), {
  size: '22px',
})

const resolvedType = computed(() => normalizeContactMethodType(props.type))

const iconDefinition = computed<IconDefinition | null>(() => {
  if (resolvedType.value === 'qq') return QqCircleFilled
  if (resolvedType.value === 'wechat') return WechatFilled
  return null
})

const brandColor = computed(() => {
  if (resolvedType.value === 'qq') return '#12B7F5'
  if (resolvedType.value === 'wechat') return '#07C160'
  return '#6B7280'
})

const resolvedSize = computed(() =>
  typeof props.size === 'number' ? `${props.size}px` : props.size,
)

const altText = computed(() => {
  if (resolvedType.value === 'qq') return 'QQ'
  if (resolvedType.value === 'wechat') return 'WeChat'
  return 'Contact'
})
</script>
