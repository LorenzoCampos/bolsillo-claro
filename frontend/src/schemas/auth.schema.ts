import { z } from 'zod';

// Schema para Login
export const loginSchema = z.object({
  email: z
    .string()
    .min(1, 'El email es requerido')
    .email('Email inválido')
    .toLowerCase()
    .trim(),
  password: z
    .string()
    .min(8, 'La contraseña debe tener mínimo 8 caracteres'),
});

export type LoginFormData = z.infer<typeof loginSchema>;

// Schema para Register
export const registerSchema = z.object({
  name: z
    .string()
    .min(1, 'El nombre es requerido')
    .max(100, 'El nombre es muy largo')
    .trim(),
  email: z
    .string()
    .min(1, 'El email es requerido')
    .email('Email inválido')
    .toLowerCase()
    .trim(),
  password: z
    .string()
    .min(8, 'La contraseña debe tener mínimo 8 caracteres')
    .max(100, 'La contraseña es muy larga'),
  confirmPassword: z
    .string()
    .min(1, 'Debes confirmar la contraseña'),
}).refine((data) => data.password === data.confirmPassword, {
  message: 'Las contraseñas no coinciden',
  path: ['confirmPassword'],
});

export type RegisterFormData = z.infer<typeof registerSchema>;
