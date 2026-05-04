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
   * 列水平对齐方式。同时作用于表头与单元格。默认 'center'。
   * 数字/金额列若需右对齐请显式传 align: 'right'，可与 numeric 叠加。
   */
  align?: 'left' | 'center' | 'right'
  /**
   * 是否数字列。设为 true 时启用 tabular-nums 等宽数字字形；
   * 不改变默认对齐方向，如需右对齐请同时传 align: 'right'。
   */
  numeric?: boolean
}
