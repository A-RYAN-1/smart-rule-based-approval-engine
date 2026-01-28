import { describe, it, expect, beforeEach } from 'vitest';
import { 
  transformLeaveRequest, 
  transformExpenseRequest, 
  transformDiscountRequest,
  normalizeStatus 
} from '@/lib/transformers';
import { RequestStatus, LeaveRequest, ExpenseRequest, DiscountRequest } from '@/types';

describe('Transformers', () => {
  
  describe('normalizeStatus', () => {
    it('should normalize approved status variants', () => {
      expect(normalizeStatus('APPROVED')).toBe('approved');
      expect(normalizeStatus('approved')).toBe('approved');
      expect(normalizeStatus('Approved')).toBe('approved');
    });

    it('should normalize rejected status variants', () => {
      expect(normalizeStatus('REJECTED')).toBe('rejected');
      expect(normalizeStatus('rejected')).toBe('rejected');
    });

    it('should normalize auto_approved status', () => {
      expect(normalizeStatus('AUTO_APPROVED')).toBe('auto_approved');
      expect(normalizeStatus('auto_approved')).toBe('auto_approved');
    });

    it('should normalize auto_rejected status', () => {
      expect(normalizeStatus('AUTO_REJECTED')).toBe('auto_rejected');
      expect(normalizeStatus('auto_rejected')).toBe('auto_rejected');
    });

    it('should normalize pending status', () => {
      expect(normalizeStatus('PENDING')).toBe('pending');
      expect(normalizeStatus('pending')).toBe('pending');
    });

    it('should normalize cancelled status', () => {
      expect(normalizeStatus('CANCELLED')).toBe('cancelled');
      expect(normalizeStatus('cancelled')).toBe('cancelled');
    });

    it('should handle undefined/null gracefully', () => {
      expect(normalizeStatus(undefined as any)).toBe('pending');
      expect(normalizeStatus(null as any)).toBe('pending');
      expect(normalizeStatus('')).toBe('pending');
    });

    it('should handle unknown status', () => {
      expect(normalizeStatus('UNKNOWN_STATUS')).toBe('pending');
    });
  });

  describe('transformLeaveRequest', () => {
    const mockLeaveData = {
      id: 1,
      user_id: 123,
      employee: 'John Doe',
      from_date: '2026-02-01T10:00:00Z',
      to_date: '2026-02-03T10:00:00Z',
      leave_type: 'EARN',
      reason: 'Family function',
      status: 'PENDING',
      created_at: '2026-01-20T10:00:00Z',
      updated_at: '2026-01-20T10:00:00Z',
    };

    it('should transform valid leave request correctly', () => {
      const result = transformLeaveRequest(mockLeaveData);
      
      expect(result.id).toBe(1);
      expect(result.userId).toBe(123);
      expect(result.userName).toBe('John Doe');
      expect(result.leaveType).toBe('EARN');
      expect(result.reason).toBe('Family function');
      expect(result.status).toBe('pending');
    });

    it('should handle missing userName fields', () => {
      const data = {
        ...mockLeaveData,
        employee: undefined,
        user_name: undefined,
        name: 'Fallback Name',
      };
      
      const result = transformLeaveRequest(data);
      expect(result.userName).toBe('Fallback Name');
    });

    it('should parse dates correctly', () => {
      const result = transformLeaveRequest(mockLeaveData);
      
      expect(result.fromDate).toBe('2026-02-01T10:00:00Z');
      expect(result.toDate).toBe('2026-02-03T10:00:00Z');
      expect(result.createdAt).toBe('2026-01-20T10:00:00Z');
    });

    it('should use fallback date when date is missing', () => {
      const data = {
        ...mockLeaveData,
        created_at: undefined,
        createdAt: undefined,
      };
      
      const result = transformLeaveRequest(data);
      expect(result.createdAt).toBeDefined();
      // Should have some date, not null
      expect(typeof result.createdAt).toBe('string');
    });

    it('should handle reason extraction from various fields', () => {
      const data = {
        ...mockLeaveData,
        reason: 'Test reason',
      };
      
      const result = transformLeaveRequest(data);
      expect(result.reason).toBe('Test reason');
    });

    it('should handle auto_approved status', () => {
      const data = {
        ...mockLeaveData,
        status: 'AUTO_APPROVED',
      };
      
      const result = transformLeaveRequest(data);
      expect(result.status).toBe('auto_approved');
    });
  });

  describe('transformExpenseRequest', () => {
    const mockExpenseData = {
      id: 1,
      user_id: 456,
      employee: 'Jane Smith',
      amount: 5000,
      category: 'Travel',
      reason: 'Client visit',
      status: 'APPROVED',
      created_at: '2026-01-18T15:00:00Z',
      updated_at: '2026-01-18T15:02:00Z',
    };

    it('should transform valid expense request correctly', () => {
      const result = transformExpenseRequest(mockExpenseData);
      
      expect(result.id).toBe(1);
      expect(result.userId).toBe(456);
      expect(result.amount).toBe(5000);
      expect(result.category).toBe('Travel');
      expect(result.status).toBe('approved');
    });

    it('should handle negative amount edge case', () => {
      const data = {
        ...mockExpenseData,
        amount: -1000,
      };
      
      const result = transformExpenseRequest(data);
      expect(result.amount).toBe(-1000); // Store as-is, validation happens elsewhere
    });

    it('should handle zero amount', () => {
      const data = {
        ...mockExpenseData,
        amount: 0,
      };
      
      const result = transformExpenseRequest(data);
      expect(result.amount).toBe(0);
    });

    it('should handle missing userName', () => {
      const data = {
        ...mockExpenseData,
        employee: undefined,
        user_name: 'Fallback User',
      };
      
      const result = transformExpenseRequest(data);
      expect(result.userName).toBe('Fallback User');
    });

    it('should handle auto_rejected status', () => {
      const data = {
        ...mockExpenseData,
        status: 'AUTO_REJECTED',
      };
      
      const result = transformExpenseRequest(data);
      expect(result.status).toBe('auto_rejected');
    });
  });

  describe('transformDiscountRequest', () => {
    const mockDiscountData = {
      id: 1,
      user_id: 789,
      employee: 'Bob Wilson',
      discount_percentage: 8,
      reason: 'Loan benefit',
      status: 'PENDING',
      created_at: '2026-01-23T09:00:00Z',
      updated_at: '2026-01-23T09:00:00Z',
    };

    it('should transform valid discount request correctly', () => {
      const result = transformDiscountRequest(mockDiscountData);
      
      expect(result.id).toBe(1);
      expect(result.userId).toBe(789);
      expect(result.discountPercentage).toBe(8);
      expect(result.reason).toBe('Loan benefit');
      expect(result.status).toBe('pending');
    });

    it('should handle zero discount', () => {
      const data = {
        ...mockDiscountData,
        discount_percentage: 0,
      };
      
      const result = transformDiscountRequest(data);
      expect(result.discountPercentage).toBe(0);
    });

    it('should handle 100% discount', () => {
      const data = {
        ...mockDiscountData,
        discount_percentage: 100,
      };
      
      const result = transformDiscountRequest(data);
      expect(result.discountPercentage).toBe(100);
    });

    it('should handle percentage > 100 (invalid but store)', () => {
      const data = {
        ...mockDiscountData,
        discount_percentage: 150,
      };
      
      const result = transformDiscountRequest(data);
      expect(result.discountPercentage).toBe(150); // Validation elsewhere
    });

    it('should handle cancelled status', () => {
      const data = {
        ...mockDiscountData,
        status: 'CANCELLED',
      };
      
      const result = transformDiscountRequest(data);
      expect(result.status).toBe('cancelled');
    });
  });
});
