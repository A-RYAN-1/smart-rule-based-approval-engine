import { describe, it, expect } from 'vitest';
import { transformDecision, transformLeaveRequest, transformExpenseRequest } from '@/lib/transformers';

const raw = {
  kind: 'AUTO_APPROVED',
  rule_id: 7,
  limit: 3,
  text: 'Approved by the system: 2 days is within the limit of 3 days set by rule #7.',
};

describe('transformDecision', () => {
  it('maps the server decision', () => {
    expect(transformDecision(raw)).toEqual({ kind: 'AUTO_APPROVED', ruleId: 7, limit: 3, text: raw.text });
  });

  it('keeps a missing rule or limit undefined instead of zero', () => {
    const d = transformDecision({ kind: 'SENT_TO_PERSON', rule_id: null, limit: null, text: 'Sent to a person for approval.' });
    expect(d?.ruleId).toBeUndefined();
    expect(d?.limit).toBeUndefined();
  });

  it('returns undefined for null, an unknown kind or no text', () => {
    expect(transformDecision(null)).toBeUndefined();
    expect(transformDecision(undefined)).toBeUndefined();
    expect(transformDecision({ ...raw, kind: 'SOMETHING' })).toBeUndefined();
    expect(transformDecision({ ...raw, text: '' })).toBeUndefined();
  });

  it('is carried onto a transformed request', () => {
    const leave = transformLeaveRequest({ id: 1, status: 'AUTO_APPROVED', decision: raw });
    expect(leave.decision?.ruleId).toBe(7);
    const expense = transformExpenseRequest({ id: 2, status: 'CANCELLED', decision: null });
    expect(expense.decision).toBeUndefined();
  });
});
