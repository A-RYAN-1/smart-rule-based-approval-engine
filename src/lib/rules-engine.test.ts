import { describe, it, expect, beforeEach } from 'vitest';
import { ApprovalRule, RequestStatus } from '@/types';

/**
 * Rules Engine Test Suite
 * Tests the logic for how rules are evaluated and applied to requests
 */

describe('Rules Engine Logic', () => {
  
  /**
   * Mock rule evaluation function
   * Simulates what backend does when evaluating rules
   */
  const evaluateRuleCondition = (condition: any, request: any): boolean => {
    if (!condition) return false;
    
    // Simple condition evaluation
    if (condition.max_amount !== undefined) {
      return request.amount <= condition.max_amount;
    }
    if (condition.min_amount !== undefined) {
      return request.amount >= condition.min_amount;
    }
    if (condition.max_days !== undefined) {
      return request.days <= condition.max_days;
    }
    if (condition.min_days !== undefined) {
      return request.days >= condition.min_days;
    }
    if (condition.leave_type !== undefined) {
      return request.leaveType === condition.leave_type;
    }
    if (condition.discount_percentage !== undefined) {
      return request.discountPercentage <= condition.discount_percentage;
    }
    
    return false;
  };

  /**
   * Apply rules to request - finds first matching rule
   */
  const applyRules = (
    request: any,
    rules: ApprovalRule[]
  ): { action?: ApprovalRule['action']; ruleId?: number } => {
    // Sort by priority (lower = higher priority)
    const sortedRules = [...rules].sort((a, b) => a.priority - b.priority);

    for (const rule of sortedRules) {
      // Skip inactive rules
      if (!rule.isActive) continue;

      // Skip if grade doesn't match
      if (rule.gradeId && rule.gradeId !== request.gradeId) continue;

      // Check if condition matches
      if (evaluateRuleCondition(rule.condition, request)) {
        return {
          action: rule.action,
          ruleId: rule.id,
        };
      }
    }

    // No rule matched
    return {};
  };

  describe('Priority-based Execution', () => {
    it('should execute highest priority rule first', () => {
      const rules: ApprovalRule[] = [
        {
          id: 1,
          requestType: 'expense',
          condition: { max_amount: 5000 },
          action: 'AUTO_APPROVE',
          priority: 2,
          isActive: true,
        },
        {
          id: 2,
          requestType: 'expense',
          condition: { max_amount: 1000 },
          action: 'AUTO_APPROVE',
          priority: 1,
          isActive: true,
        },
      ];

      const request = { amount: 800, gradeId: 1 };
      const result = applyRules(request, rules);

      // Priority 1 should match before priority 2
      expect(result.ruleId).toBe(2);
    });

    it('should stop at first matching rule', () => {
      const rules: ApprovalRule[] = [
        {
          id: 1,
          requestType: 'expense',
          condition: { max_amount: 10000 },
          action: 'AUTO_APPROVE',
          priority: 1,
          isActive: true,
        },
        {
          id: 2,
          requestType: 'expense',
          condition: { max_amount: 5000 },
          action: 'AUTO_REJECT',
          priority: 2,
          isActive: true,
        },
      ];

      const request = { amount: 3000, gradeId: 1 };
      const result = applyRules(request, rules);

      // Should return first matching (rule 1), not rule 2
      expect(result.ruleId).toBe(1);
      expect(result.action).toBe('AUTO_APPROVE');
    });

    it('should skip disabled rules', () => {
      const rules: ApprovalRule[] = [
        {
          id: 1,
          requestType: 'expense',
          condition: { max_amount: 5000 },
          action: 'AUTO_APPROVE',
          priority: 1,
          isActive: false, // DISABLED
        },
        {
          id: 2,
          requestType: 'expense',
          condition: { max_amount: 10000 },
          action: 'AUTO_APPROVE',
          priority: 2,
          isActive: true,
        },
      ];

      const request = { amount: 3000, gradeId: 1 };
      const result = applyRules(request, rules);

      // Should skip disabled rule 1, use rule 2
      expect(result.ruleId).toBe(2);
    });

    it('should return no action if no rule matches', () => {
      const rules: ApprovalRule[] = [
        {
          id: 1,
          requestType: 'expense',
          condition: { max_amount: 1000 },
          action: 'AUTO_APPROVE',
          priority: 1,
          isActive: true,
        },
      ];

      const request = { amount: 5000, gradeId: 1 };
      const result = applyRules(request, rules);

      // No matching rule
      expect(result.action).toBeUndefined();
      expect(result.ruleId).toBeUndefined();
    });
  });

  describe('Condition Evaluation - Leave', () => {
    it('should match leave with max days condition', () => {
      const rules: ApprovalRule[] = [
        {
          id: 1,
          requestType: 'leave',
          condition: { max_days: 2, leave_type: 'SICK' },
          action: 'AUTO_APPROVE',
          priority: 1,
          isActive: true,
        },
      ];

      const request = { days: 2, leaveType: 'SICK', gradeId: 1 };
      const result = applyRules(request, rules);

      expect(result.ruleId).toBe(1);
    });

    it('should not match leave exceeding max days', () => {
      const rules: ApprovalRule[] = [
        {
          id: 1,
          requestType: 'leave',
          condition: { max_days: 2 },
          action: 'AUTO_APPROVE',
          priority: 1,
          isActive: true,
        },
      ];

      const request = { days: 3, gradeId: 1 };
      const result = applyRules(request, rules);

      expect(result.action).toBeUndefined();
    });

    it('should auto-reject long leaves', () => {
      const rules: ApprovalRule[] = [
        {
          id: 1,
          requestType: 'leave',
          condition: { min_days: 8 },
          action: 'AUTO_REJECT',
          priority: 1,
          isActive: true,
        },
      ];

      const request = { days: 10, gradeId: 1 };
      const result = applyRules(request, rules);

      expect(result.action).toBe('AUTO_REJECT');
    });
  });

  describe('Condition Evaluation - Expense', () => {
    it('should auto-approve small expenses', () => {
      const rules: ApprovalRule[] = [
        {
          id: 1,
          requestType: 'expense',
          condition: { max_amount: 1000 },
          action: 'AUTO_APPROVE',
          priority: 1,
          isActive: true,
        },
      ];

      const request = { amount: 500, gradeId: 1 };
      const result = applyRules(request, rules);

      expect(result.action).toBe('AUTO_APPROVE');
    });

    it('should assign large expenses to approver', () => {
      const rules: ApprovalRule[] = [
        {
          id: 1,
          requestType: 'expense',
          condition: { min_amount: 5000 },
          action: 'assign_approver',
          priority: 1,
          isActive: true,
        },
      ];

      const request = { amount: 10000, gradeId: 1 };
      const result = applyRules(request, rules);

      expect(result.action).toBe('assign_approver');
    });

    it('should auto-reject expenses exceeding limit', () => {
      const rules: ApprovalRule[] = [
        {
          id: 1,
          requestType: 'expense',
          condition: { min_amount: 50000 },
          action: 'AUTO_REJECT',
          priority: 1,
          isActive: true,
        },
      ];

      const request = { amount: 75000, gradeId: 1 };
      const result = applyRules(request, rules);

      expect(result.action).toBe('AUTO_REJECT');
    });
  });

  describe('Grade-based Filtering', () => {
    it('should only apply rule to matching grade', () => {
      const rules: ApprovalRule[] = [
        {
          id: 1,
          requestType: 'expense',
          condition: { max_amount: 5000 },
          action: 'AUTO_APPROVE',
          priority: 1,
          isActive: true,
          gradeId: 2, // Manager only
        },
      ];

      // Employee request (grade 1)
      const employeeRequest = { amount: 3000, gradeId: 1 };
      const result1 = applyRules(employeeRequest, rules);
      expect(result1.action).toBeUndefined(); // Rule doesn't apply

      // Manager request (grade 2)
      const managerRequest = { amount: 3000, gradeId: 2 };
      const result2 = applyRules(managerRequest, rules);
      expect(result2.action).toBe('AUTO_APPROVE'); // Rule applies
    });

    it('should apply rule to all grades if gradeId is not set', () => {
      const rules: ApprovalRule[] = [
        {
          id: 1,
          requestType: 'expense',
          condition: { max_amount: 5000 },
          action: 'AUTO_APPROVE',
          priority: 1,
          isActive: true,
          // gradeId not set = applies to all
        },
      ];

      const request1 = { amount: 3000, gradeId: 1 };
      const result1 = applyRules(request1, rules);
      expect(result1.action).toBe('AUTO_APPROVE');

      const request2 = { amount: 3000, gradeId: 2 };
      const result2 = applyRules(request2, rules);
      expect(result2.action).toBe('AUTO_APPROVE');
    });
  });

  describe('Edge Cases', () => {
    it('should handle empty rules array', () => {
      const request = { amount: 5000, gradeId: 1 };
      const result = applyRules(request, []);

      expect(result.action).toBeUndefined();
    });

    it('should handle all rules disabled', () => {
      const rules: ApprovalRule[] = [
        {
          id: 1,
          requestType: 'expense',
          condition: { max_amount: 1000 },
          action: 'AUTO_APPROVE',
          priority: 1,
          isActive: false,
        },
        {
          id: 2,
          requestType: 'expense',
          condition: { max_amount: 5000 },
          action: 'AUTO_APPROVE',
          priority: 2,
          isActive: false,
        },
      ];

      const request = { amount: 3000, gradeId: 1 };
      const result = applyRules(request, rules);

      expect(result.action).toBeUndefined();
    });

    it('should handle conflicting conditions with priority', () => {
      const rules: ApprovalRule[] = [
        {
          id: 1,
          requestType: 'expense',
          condition: { min_amount: 4000 },
          action: 'AUTO_REJECT', // Higher priority
          priority: 1,
          isActive: true,
        },
        {
          id: 2,
          requestType: 'expense',
          condition: { max_amount: 5000 },
          action: 'AUTO_APPROVE', // Lower priority
          priority: 2,
          isActive: true,
        },
      ];

      const request = { amount: 4500, gradeId: 1 };
      const result = applyRules(request, rules);

      // Both match, but priority 1 executed first
      expect(result.action).toBe('AUTO_REJECT');
      expect(result.ruleId).toBe(1);
    });

    it('should handle same priority rules (order matters)', () => {
      const rules: ApprovalRule[] = [
        {
          id: 1,
          requestType: 'expense',
          condition: { max_amount: 1000 },
          action: 'AUTO_APPROVE',
          priority: 1,
          isActive: true,
        },
        {
          id: 2,
          requestType: 'expense',
          condition: { max_amount: 5000 },
          action: 'AUTO_APPROVE',
          priority: 1, // Same priority!
          isActive: true,
        },
      ];

      const request = { amount: 500, gradeId: 1 };
      const result = applyRules(request, rules);

      // Undefined which one - but should be one of them
      expect([1, 2]).toContain(result.ruleId);
    });

    it('should handle zero/negative conditions', () => {
      const rules: ApprovalRule[] = [
        {
          id: 1,
          requestType: 'expense',
          condition: { max_amount: 0 },
          action: 'AUTO_REJECT',
          priority: 1,
          isActive: true,
        },
      ];

      const request = { amount: 0, gradeId: 1 };
      const result = applyRules(request, rules);

      expect(result.action).toBe('AUTO_REJECT');
    });

    it('should handle invalid condition format gracefully', () => {
      const rules: ApprovalRule[] = [
        {
          id: 1,
          requestType: 'expense',
          condition: {}, // Empty condition
          action: 'AUTO_APPROVE',
          priority: 1,
          isActive: true,
        },
      ];

      const request = { amount: 5000, gradeId: 1 };
      const result = applyRules(request, rules);

      // Empty condition should not match
      expect(result.action).toBeUndefined();
    });
  });

  describe('Multiple Rule Types Combined', () => {
    it('should handle mixed rule types with different priorities', () => {
      const rules: ApprovalRule[] = [
        {
          id: 1,
          requestType: 'expense',
          condition: { min_amount: 5000 },
          action: 'assign_approver',
          priority: 1,
          isActive: true,
        },
        {
          id: 2,
          requestType: 'expense',
          condition: { max_amount: 1000 },
          action: 'AUTO_APPROVE',
          priority: 2,
          isActive: true,
        },
        {
          id: 3,
          requestType: 'expense',
          condition: { min_amount: 50000 },
          action: 'AUTO_REJECT',
          priority: 3,
          isActive: true,
        },
      ];

      // Test $500
      expect(applyRules({ amount: 500, gradeId: 1 }, rules).action).toBe(
        'AUTO_APPROVE'
      ); // Matches rule 2

      // Test $7000
      expect(applyRules({ amount: 7000, gradeId: 1 }, rules).action).toBe(
        'assign_approver'
      ); // Matches rule 1

      // Test $75000
      expect(applyRules({ amount: 75000, gradeId: 1 }, rules).action).toBe(
        'assign_approver'
      ); // Matches rule 1 (higher priority than rule 3)
    });
  });
});
