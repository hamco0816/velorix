import alipayIcon from '@/assets/icons/alipay.svg'
import easypayIcon from '@/assets/icons/easypay.svg'
import stripeIcon from '@/assets/icons/stripe.svg'
import wxpayIcon from '@/assets/icons/wxpay.svg'

export type PaymentBrand = 'alipay' | 'wxpay' | 'stripe' | 'easypay'

export function normalizePaymentBrand(type: unknown): PaymentBrand | null {
  const value = String(type ?? '').trim().toLowerCase()
  if (!value) return null
  if (value.includes('alipay')) return 'alipay'
  if (value.includes('wxpay') || value.includes('wechat')) return 'wxpay'
  if (value.includes('stripe')) return 'stripe'
  if (value.includes('easypay')) return 'easypay'
  return null
}

export function paymentBrandIcon(type: unknown): string | null {
  switch (normalizePaymentBrand(type)) {
    case 'alipay':
      return alipayIcon
    case 'wxpay':
      return wxpayIcon
    case 'stripe':
      return stripeIcon
    case 'easypay':
      return easypayIcon
    default:
      return null
  }
}

export function hasPaymentBrandIcon(type: unknown): boolean {
  return paymentBrandIcon(type) !== null
}
