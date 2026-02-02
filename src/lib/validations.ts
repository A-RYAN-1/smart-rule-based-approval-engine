import { z } from 'zod';


export const sanitizeInput = (val: string): string => {
    return val.trim().replace(/<[^>]*>?/gm, '');
};


export const loginSchema = z.object({
  email: z
    .string({ message: 'Email is required' })
    .trim()
    .min(5, 'Email must be at least 5 characters')
    .max(254, 'Email must not exceed 254 characters')
    .email('Invalid email format')
    .toLowerCase()
    .refine(
      (email) => !email.endsWith('@example.com'),
      'Please use a valid email domain'
    ),

  password: z
    .string({ message: 'Password is required' })
    .min(6, 'Password must be at least 8 characters')
    .max(128, 'Password must not exceed 128 characters'),
});


export const registerSchema = z.object({
  name: z
    .string({ message: 'Name is required' })
    .trim()
    .min(2, 'Name must be at least 2 characters')
    .max(50, 'Name must not exceed 50 characters')
    .regex(
      /^[a-zA-Z\s'-]+$/,
      'Name can only contain letters, spaces, hyphens, and apostrophes'
    )
    .refine(
      (name) => !name.includes('  '),
      'Name cannot have consecutive spaces'
    ),

  email: z
    .string({ message: 'Email is required' })
    .trim()
    .min(5, 'Email must be at least 5 characters')
    .max(254, 'Email must not exceed 254 characters')
    .email('Invalid email format')
    .toLowerCase()
    .refine(
      (email) => !email.endsWith('@example.com'),
      'Please use a valid email domain'
    ),

  password: z
    .string({ message: 'Password is required' })
    .min(8, 'Password must be at least 8 characters')
    .max(128, 'Password must not exceed 128 characters')
    .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
    .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
    .regex(/[0-9]/, 'Password must contain at least one digit')
    .regex(
      /[@$!%*?&#^()\-_=+\[\]{};':"\\|,.<>/?]/,
      'Password must contain at least one special character'
    )
    .refine(
      (password) => new Set(password).size >= 5,
      'Password must contain at least 5 unique characters'
    ),
});


export const passwordChangeSchema = z.object({
  currentPassword: z
    .string({ message: 'Current password is required' })
    .min(1, 'Current password is required'),

  newPassword: z
    .string({ message: 'New password is required' })
    .min(8, 'New password must be at least 8 characters')
    .max(128, 'New password must not exceed 128 characters')
    .regex(/[a-z]/, 'New password must contain lowercase letter')
    .regex(/[A-Z]/, 'New password must contain uppercase letter')
    .regex(/[0-9]/, 'New password must contain digit')
    .regex(
      /[@$!%*?&#^()\-_=+\[\]{};':"\\|,.<>/?]/,
      'New password must contain special character'
    ),
}).refine(
  (data) => data.currentPassword !== data.newPassword,
  { message: 'New password must be different from current password', path: ['newPassword'] }
);

export const leaveRequestSchema = z.object({
    fromDate: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'Invalid date format (use YYYY-MM-DD)'),
    toDate: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'Invalid date format (use YYYY-MM-DD)'),
    leaveType: z.enum(['EARN', 'SICK', 'CASUAL', 'UNPAID']),
    reason: z.string().transform(sanitizeInput).pipe(
        z.string()
          .min(10, 'Please provide a more detailed reason (minimum 10 characters)')
          .max(1000, 'Reason is too long (maximum 1000 characters)')
    ),
}).refine(
  (data) => new Date(data.toDate) >= new Date(data.fromDate),
  { message: 'End date must be after or same as start date', path: ['toDate'] }
);

export const expenseRequestSchema = z.object({
    amount: z
      .number()
      .positive('Amount must be greater than 0')
      .max(1000000, 'Amount exceeds maximum limit'),
    category: z
      .string()
      .min(1, 'Category is required')
      .max(50, 'Category is too long'),
    reason: z.string().transform(sanitizeInput).pipe(
        z.string()
          .min(5, 'Description must be at least 5 characters')
          .max(500, 'Description is too long')
    ),
});

export const discountRequestSchema = z.object({
    discountPercentage: z
      .number()
      .min(1, 'Discount must be at least 1%')
      .max(25, 'Discount cannot exceed 25%'),
    reason: z.string().transform(sanitizeInput).pipe(
        z.string()
          .min(5, 'Justification must be at least 5 characters')
          .max(500, 'Justification is too long')
    ),
});
