import { LeaveRequest, ExpenseRequest, DiscountRequest, Holiday, RequestStatus, ApprovalRule, LeaveType, RuleCondition } from '@/types';
import type { RawRecord } from '@/lib/api';

export const normalizeStatus = (status: string | undefined | null): RequestStatus => {
    if (!status) return 'pending'; // Default if missing
    const s = status.toUpperCase();
    if (s === 'AUTO_APPROVED') return 'auto_approved';
    if (s === 'APPROVED') return 'approved';
    if (s === 'AUTO_REJECTED') return 'auto_rejected';
    if (s === 'REJECTED') return 'rejected';
    if (s === 'CANCELLED') return 'cancelled';
    return 'pending';
};

const FALLBACK_DATE = '1970-01-01T00:00:00Z';

const safeExtractDate = (data: RawRecord, keys: string[]): string | null => {
    for (const key of keys) {
        const val = data[key];
        if (!val) continue;
        if (typeof val === 'string' && val.trim() !== '') return val;
        if (typeof val === 'number') return new Date(val).toISOString();
        if (val instanceof Date) return val.toISOString();
    }
    return null;
};

const safeExtractString = (data: RawRecord, keys: string[]): string | undefined => {
    for (const key of keys) {
        const val = data[key];
        if (val !== undefined && val !== null && String(val).trim() !== '') return String(val);
    }
    return undefined;
};

const num = (v: unknown): number => (typeof v === 'number' ? v : Number(v) || 0);
const str = (v: unknown): string | undefined =>
    v === undefined || v === null || String(v).trim() === '' ? undefined : String(v);
const firstString = (data: RawRecord, keys: string[]): string | undefined => {
    for (const key of keys) {
        const v = str(data[key]);
        if (v !== undefined) return v;
    }
    return undefined;
};
const USER_NAME_KEYS = ['employee', 'user_name', 'name', 'staff_name'];

const REASON_KEYS = ['reason', 'description', 'remark', 'remarks', 'purpose', 'explanation', 'details'];
const COMMENT_KEYS = ['comment', 'comments', 'approval_comment', 'status_reason', 'manager_remark', 'manager_comment', 'decision_note', 'status_comment', 'feedback'];
const DATE_KEYS = ['created_at', 'createdAt', 'inserted_at', 'insertedAt', 'creation_date', 'registration_date', 'date', 'timestamp', 'submitted_at'];
const FROM_DATE_KEYS = ['from_date', 'fromDate', 'start_date', 'startDate'];
const TO_DATE_KEYS = ['to_date', 'toDate', 'end_date', 'endDate'];


export const transformLeaveRequest = (data: RawRecord): LeaveRequest => {
    const created = safeExtractDate(data, DATE_KEYS);
    const reason = safeExtractString(data, REASON_KEYS);
    const comment = safeExtractString(data, COMMENT_KEYS);

    const result: LeaveRequest = {
        id: num(data.id),
        userId: num(data.user_id),
        userName: firstString(data, USER_NAME_KEYS),
        fromDate: safeExtractDate(data, FROM_DATE_KEYS) || FALLBACK_DATE,
        toDate: safeExtractDate(data, TO_DATE_KEYS) || FALLBACK_DATE,
        leaveType: (firstString(data, ['leave_type', 'leaveType']) ?? 'CASUAL') as LeaveType,
        reason: reason || 'No reason found',
        status: normalizeStatus(str(data.status)),
        statusReason: comment,
        createdAt: created || FALLBACK_DATE,
        updatedAt: safeExtractDate(data, ['updated_at', 'updatedAt']) || created || FALLBACK_DATE,
    };
    return result;
};

export const transformExpenseRequest = (data: RawRecord): ExpenseRequest => {
    const created = safeExtractDate(data, DATE_KEYS);
    const reason = safeExtractString(data, REASON_KEYS);
    const comment = safeExtractString(data, COMMENT_KEYS);

    const result: ExpenseRequest = {
        id: num(data.id),
        userId: num(data.user_id),
        userName: firstString(data, USER_NAME_KEYS),
        amount: num(data.amount),
        category: str(data.category) ?? '',
        reason: reason || 'No reason found',
        status: normalizeStatus(str(data.status)),
        statusReason: comment,
        createdAt: created || FALLBACK_DATE,
        updatedAt: safeExtractDate(data, ['updated_at', 'updatedAt']) || created || FALLBACK_DATE,
    };
    return result;
};

export const transformDiscountRequest = (data: RawRecord): DiscountRequest => {
    const created = safeExtractDate(data, DATE_KEYS);
    const reason = safeExtractString(data, REASON_KEYS);
    const comment = safeExtractString(data, COMMENT_KEYS);

    const result: DiscountRequest = {
        id: num(data.id),
        userId: num(data.user_id),
        userName: firstString(data, USER_NAME_KEYS),
        discountPercentage: num(firstString(data, ['discount_percentage', 'discountPercentage', 'percentage', 'discount', 'percent'])),
        reason: reason || 'No reason found',
        status: normalizeStatus(str(data.status)),
        statusReason: comment,
        createdAt: created || FALLBACK_DATE,
        updatedAt: safeExtractDate(data, ['updated_at', 'updatedAt']) || created || FALLBACK_DATE,
    };
    return result;
};

export const transformHoliday = (data: RawRecord): Holiday => ({
    id: num(data.id),
    date: safeExtractDate(data, ['date']) || FALLBACK_DATE,
    description: str(data.description) ?? '',
});

export const transformApprovalRule = (data: RawRecord): ApprovalRule => ({
    id: num(data.id),
    requestType: (firstString(data, ['request_type', 'requestType']) ?? 'leave').toLowerCase() as ApprovalRule['requestType'],
    condition: (data.condition ?? {}) as RuleCondition,
    action: (str(data.action) ?? 'AUTO_APPROVE') as ApprovalRule['action'],
    priority: num(data.priority) || 1,
    isActive: Boolean(data.active ?? data.isActive ?? true),
    gradeId: num(data.grade_id ?? data.gradeId) || undefined,
    createdAt: firstString(data, ['timestamp', 'createdAt']),
});
