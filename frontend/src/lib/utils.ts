import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { format } from 'date-fns';
import { es } from 'date-fns/locale';
import type { Currency } from '@/types/api';
import { CURRENCY_SYMBOLS } from './constants';

/**
 * Combina clases de Tailwind CSS de manera eficiente
 * Útil para componentes con estilos condicionales
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

/**
 * Formatea un monto con símbolo de moneda
 * @param amount - Monto numérico
 * @param currency - Moneda (ARS | USD | EUR)
 * @param decimals - Cantidad de decimales (default: 2)
 */
export function formatCurrency(
  amount: number,
  currency: Currency,
  decimals: number = 2
): string {
  const symbol = CURRENCY_SYMBOLS[currency];
  const formatted = amount.toFixed(decimals);
  return `${symbol} ${formatted}`;
}

/**
 * Formatea una fecha en formato YYYY-MM-DD a formato legible
 * @param dateString - Fecha en formato YYYY-MM-DD
 * @param formatString - Formato de salida (default: 'dd/MM/yyyy')
 */
export function formatDate(
  dateString: string,
  formatString: string = 'dd/MM/yyyy'
): string {
  const date = new Date(dateString + 'T00:00:00');
  return format(date, formatString, { locale: es });
}

/**
 * Convierte una fecha Date a formato YYYY-MM-DD para la API
 */
export function toApiDateFormat(date: Date): string {
  return format(date, 'yyyy-MM-dd');
}

/**
 * Obtiene el primer día del mes actual en formato YYYY-MM-DD
 */
export function getCurrentMonthStart(): string {
  const now = new Date();
  return format(new Date(now.getFullYear(), now.getMonth(), 1), 'yyyy-MM-dd');
}

/**
 * Obtiene el mes actual en formato YYYY-MM
 */
export function getCurrentMonth(): string {
  return format(new Date(), 'yyyy-MM');
}

/**
 * Calcula el porcentaje de progreso
 */
export function calculateProgress(current: number, target: number): number {
  if (target === 0) return 0;
  return Math.min((current / target) * 100, 100);
}
