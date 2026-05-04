/**
 * Common component types
 */

export interface Column {
  key: string
  label: string
  sortable?: boolean
  class?: string
  formatter?: (value: any, row: any) => string
  /**
   * 列水平对齐方式。同时作用于表头与单元格。默认 'left'。
   * 数字列推荐使用 numeric: true（自动右对齐 + 等宽数字）。
   */
  align?: 'left' | 'center' | 'right'
  /**
   * 是否数字列。设为 true 时单元格右对齐并启用 tabular-nums 等宽数字字形，
   * 优先级高于 align。
   */
  numeric?: boolean
}
