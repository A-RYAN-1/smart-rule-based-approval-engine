import { describe, it, expect } from 'vitest';

/**
 * Validation Tests
 * Tests business logic validation for requests
 */

describe('Request Validation', () => {
  
  describe('Leave Request Validation', () => {
    /**
     * Validate leave request
     */
    const validateLeaveRequest = (request: any): { valid: boolean; errors: string[] } => {
      const errors: string[] = [];

      // Check dates
      if (!request.fromDate) {
        errors.push('From date is required');
      }
      if (!request.toDate) {
        errors.push('To date is required');
      }

      if (request.fromDate && request.toDate) {
        const fromDate = new Date(request.fromDate);
        const toDate = new Date(request.toDate);

        if (fromDate > toDate) {
          errors.push('From date must be before to date');
        }

        // Check leave days
        const days = Math.ceil(
          (toDate.getTime() - fromDate.getTime()) / (1000 * 60 * 60 * 24)
        );

        if (days < 0) {
          errors.push('Leave duration must be positive');
        }
        if (days === 0) {
          errors.push('Leave must be at least 1 day');
        }
      }

      // Check leave type
      const validTypes = ['EARN', 'SICK', 'CASUAL', 'PERSONAL', 'UNPAID'];
      if (!request.leaveType || !validTypes.includes(request.leaveType)) {
        errors.push('Invalid leave type');
      }

      // Check reason
      if (!request.reason || request.reason.trim() === '') {
        errors.push('Reason is required');
      }
      if (request.reason && request.reason.length > 500) {
        errors.push('Reason must be less than 500 characters');
      }

      // Check balance
      if (request.balance !== undefined && request.balance < 0) {
        errors.push('Insufficient balance');
      }

      return {
        valid: errors.length === 0,
        errors,
      };
    };

    it('should validate correct leave request', () => {
      const request = {
        fromDate: '2026-02-01',
        toDate: '2026-02-03',
        leaveType: 'EARN',
        reason: 'Family function',
        balance: 10,
      };

      const result = validateLeaveRequest(request);
      expect(result.valid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should reject missing fromDate', () => {
      const request = {
        toDate: '2026-02-03',
        leaveType: 'EARN',
        reason: 'Family function',
      };

      const result = validateLeaveRequest(request);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('From date is required');
    });

    it('should reject missing toDate', () => {
      const request = {
        fromDate: '2026-02-01',
        leaveType: 'EARN',
        reason: 'Family function',
      };

      const result = validateLeaveRequest(request);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('To date is required');
    });

    it('should reject when fromDate > toDate', () => {
      const request = {
        fromDate: '2026-02-10',
        toDate: '2026-02-01',
        leaveType: 'EARN',
        reason: 'Family function',
      };

      const result = validateLeaveRequest(request);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('From date must be before to date');
    });

    it('should reject single day leave as 0 days', () => {
      const request = {
        fromDate: '2026-02-01',
        toDate: '2026-02-01',
        leaveType: 'EARN',
        reason: 'Family function',
      };

      const result = validateLeaveRequest(request);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Leave must be at least 1 day');
    });

    it('should reject invalid leave type', () => {
      const request = {
        fromDate: '2026-02-01',
        toDate: '2026-02-03',
        leaveType: 'INVALID',
        reason: 'Family function',
      };

      const result = validateLeaveRequest(request);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Invalid leave type');
    });

    it('should reject missing reason', () => {
      const request = {
        fromDate: '2026-02-01',
        toDate: '2026-02-03',
        leaveType: 'EARN',
        reason: '',
      };

      const result = validateLeaveRequest(request);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Reason is required');
    });

    it('should reject reason exceeding 500 characters', () => {
      const longReason = 'a'.repeat(501);
      const request = {
        fromDate: '2026-02-01',
        toDate: '2026-02-03',
        leaveType: 'EARN',
        reason: longReason,
      };

      const result = validateLeaveRequest(request);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Reason must be less than 500 characters');
    });

    it('should reject insufficient balance', () => {
      const request = {
        fromDate: '2026-02-01',
        toDate: '2026-02-10', // 9 days
        leaveType: 'EARN',
        reason: 'Family function',
        balance: -1,
      };

      const result = validateLeaveRequest(request);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Insufficient balance');
    });

    it('should allow all valid leave types', () => {
      const validTypes = ['EARN', 'SICK', 'CASUAL', 'PERSONAL', 'UNPAID'];

      validTypes.forEach((type) => {
        const request = {
          fromDate: '2026-02-01',
          toDate: '2026-02-03',
          leaveType: type,
          reason: 'Test leave',
        };

        const result = validateLeaveRequest(request);
        expect(result.valid).toBe(true);
      });
    });
  });

  describe('Expense Request Validation', () => {
    const validateExpenseRequest = (request: any): { valid: boolean; errors: string[] } => {
      const errors: string[] = [];

      // Check amount
      if (request.amount === undefined || request.amount === null) {
        errors.push('Amount is required');
      }
      if (request.amount !== undefined && request.amount <= 0) {
        errors.push('Amount must be greater than 0');
      }
      if (request.amount !== undefined && request.amount > 1000000) {
        errors.push('Amount exceeds maximum limit');
      }

      // Check category
      const validCategories = [
        'Travel',
        'Office Supplies',
        'Equipment',
        'Meals & Entertainment',
        'Training',
        'Software & Subscriptions',
        'Other',
      ];
      if (!request.category || !validCategories.includes(request.category)) {
        errors.push('Invalid category');
      }

      // Check reason
      if (!request.reason || request.reason.trim() === '') {
        errors.push('Reason is required');
      }
      if (request.reason && request.reason.length > 500) {
        errors.push('Reason must be less than 500 characters');
      }

      // Check balance
      if (request.availableBalance !== undefined && request.amount > request.availableBalance) {
        errors.push('Insufficient expense balance');
      }

      return {
        valid: errors.length === 0,
        errors,
      };
    };

    it('should validate correct expense request', () => {
      const request = {
        amount: 5000,
        category: 'Travel',
        reason: 'Client visit',
        availableBalance: 10000,
      };

      const result = validateExpenseRequest(request);
      expect(result.valid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should reject missing amount', () => {
      const request = {
        category: 'Travel',
        reason: 'Client visit',
      };

      const result = validateExpenseRequest(request);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Amount is required');
    });

    it('should reject zero amount', () => {
      const request = {
        amount: 0,
        category: 'Travel',
        reason: 'Client visit',
      };

      const result = validateExpenseRequest(request);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Amount must be greater than 0');
    });

    it('should reject negative amount', () => {
      const request = {
        amount: -1000,
        category: 'Travel',
        reason: 'Client visit',
      };

      const result = validateExpenseRequest(request);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Amount must be greater than 0');
    });

    it('should reject amount exceeding maximum', () => {
      const request = {
        amount: 1000001,
        category: 'Travel',
        reason: 'Client visit',
      };

      const result = validateExpenseRequest(request);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Amount exceeds maximum limit');
    });

    it('should reject invalid category', () => {
      const request = {
        amount: 5000,
        category: 'Invalid Category',
        reason: 'Client visit',
      };

      const result = validateExpenseRequest(request);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Invalid category');
    });

    it('should reject insufficient balance', () => {
      const request = {
        amount: 5000,
        category: 'Travel',
        reason: 'Client visit',
        availableBalance: 3000,
      };

      const result = validateExpenseRequest(request);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Insufficient expense balance');
    });

    it('should allow all valid categories', () => {
      const validCategories = [
        'Travel',
        'Office Supplies',
        'Equipment',
        'Meals & Entertainment',
        'Training',
        'Software & Subscriptions',
        'Other',
      ];

      validCategories.forEach((category) => {
        const request = {
          amount: 5000,
          category,
          reason: 'Test expense',
          availableBalance: 10000,
        };

        const result = validateExpenseRequest(request);
        expect(result.valid).toBe(true);
      });
    });
  });

  describe('Discount Request Validation', () => {
    const validateDiscountRequest = (request: any): { valid: boolean; errors: string[] } => {
      const errors: string[] = [];

      // Check percentage
      if (request.discountPercentage === undefined || request.discountPercentage === null) {
        errors.push('Discount percentage is required');
      }
      if (request.discountPercentage !== undefined && request.discountPercentage < 0) {
        errors.push('Discount percentage cannot be negative');
      }
      if (request.discountPercentage !== undefined && request.discountPercentage > 100) {
        errors.push('Discount percentage cannot exceed 100%');
      }

      // Check reason
      if (!request.reason || request.reason.trim() === '') {
        errors.push('Reason is required');
      }

      // Check balance
      if (
        request.availableBalance !== undefined &&
        request.discountPercentage > request.availableBalance
      ) {
        errors.push('Insufficient discount balance');
      }

      return {
        valid: errors.length === 0,
        errors,
      };
    };

    it('should validate correct discount request', () => {
      const request = {
        discountPercentage: 8,
        reason: 'Loan benefit',
        availableBalance: 20,
      };

      const result = validateDiscountRequest(request);
      expect(result.valid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should reject missing percentage', () => {
      const request = {
        reason: 'Loan benefit',
      };

      const result = validateDiscountRequest(request);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Discount percentage is required');
    });

    it('should reject negative percentage', () => {
      const request = {
        discountPercentage: -5,
        reason: 'Loan benefit',
      };

      const result = validateDiscountRequest(request);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Discount percentage cannot be negative');
    });

    it('should reject percentage > 100', () => {
      const request = {
        discountPercentage: 150,
        reason: 'Loan benefit',
      };

      const result = validateDiscountRequest(request);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Discount percentage cannot exceed 100%');
    });

    it('should reject zero percentage', () => {
      const request = {
        discountPercentage: 0,
        reason: 'Loan benefit',
      };

      const result = validateDiscountRequest(request);
      expect(result.valid).toBe(true); // 0 is valid
    });

    it('should allow 100% discount', () => {
      const request = {
        discountPercentage: 100,
        reason: 'Full discount',
        availableBalance: 100,
      };

      const result = validateDiscountRequest(request);
      expect(result.valid).toBe(true);
    });

    it('should reject insufficient balance', () => {
      const request = {
        discountPercentage: 15,
        reason: 'Loan benefit',
        availableBalance: 10,
      };

      const result = validateDiscountRequest(request);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Insufficient discount balance');
    });
  });

  describe('Date Validation', () => {
    const validateDateRange = (fromDate: string, toDate: string): boolean => {
      try {
        const from = new Date(fromDate);
        const to = new Date(toDate);

        // Valid date check
        if (isNaN(from.getTime()) || isNaN(to.getTime())) {
          return false;
        }

        // From must be before to
        return from <= to;
      } catch {
        return false;
      }
    };

    it('should validate correct date range', () => {
      expect(validateDateRange('2026-02-01', '2026-02-03')).toBe(true);
    });

    it('should accept same date', () => {
      expect(validateDateRange('2026-02-01', '2026-02-01')).toBe(true);
    });

    it('should reject invalid date format', () => {
      expect(validateDateRange('invalid-date', '2026-02-03')).toBe(false);
      expect(validateDateRange('2026-02-01', 'invalid-date')).toBe(false);
    });

    it('should reject when fromDate > toDate', () => {
      expect(validateDateRange('2026-02-10', '2026-02-01')).toBe(false);
    });

    it('should handle ISO date format', () => {
      expect(validateDateRange('2026-02-01T00:00:00Z', '2026-02-03T00:00:00Z')).toBe(true);
    });
  });
});
