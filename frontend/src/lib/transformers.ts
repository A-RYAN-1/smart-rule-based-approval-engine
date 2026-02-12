import { LeaveRequest, ExpenseRequest, DiscountRequest, Holiday, RequestStatus } from '@/types';

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

const safeExtractDate = (data: any, keys: string[]): string | null => {
    for (const key of keys) {
        const val = data[key];
        if (!val) continue;
        if (typeof val === 'string' && val.trim() !== '') return val;
        if (typeof val === 'number') return new Date(val).toISOString();
        if (val instanceof Date) return val.toISOString();
    }
    return null;
};

const safeExtractString = (data: any, keys: string[]): string | undefined => {
    for (const key of keys) {
        const val = data[key];
        if (val !== undefined && val !== null && String(val).trim() !== '') return String(val);
    }
    return undefined;
};

const REASON_KEYS = ['reason', 'description', 'remark', 'remarks', 'purpose', 'explanation', 'details'];
const COMMENT_KEYS = ['comment', 'comments', 'approval_comment', 'status_reason', 'manager_remark', 'manager_comment', 'decision_note', 'status_comment', 'feedback'];
const DATE_KEYS = ['created_at', 'createdAt', 'inserted_at', 'insertedAt', 'creation_date', 'registration_date', 'date', 'timestamp', 'submitted_at'];
const FROM_DATE_KEYS = ['from_date', 'fromDate', 'start_date', 'startDate'];
const TO_DATE_KEYS = ['to_date', 'toDate', 'end_date', 'endDate'];


export const transformLeaveRequest = (data: any): LeaveRequest => {
    console.log('[DEBUG] Raw Leave Request:', data);
    const created = safeExtractDate(data, DATE_KEYS);
    const reason = safeExtractString(data, REASON_KEYS);
    const comment = safeExtractString(data, COMMENT_KEYS);

    const result: LeaveRequest = {
        id: data.id,
        userId: data.user_id,
        userName: data.employee || data.user_name || data.name || data.staff_name,
        fromDate: safeExtractDate(data, FROM_DATE_KEYS) || FALLBACK_DATE,
        toDate: safeExtractDate(data, TO_DATE_KEYS) || FALLBACK_DATE,
        leaveType: data.leave_type || data.leaveType,
        reason: reason || 'No reason found',
        status: normalizeStatus(data.status),
        statusReason: comment,
        createdAt: created || FALLBACK_DATE,
        updatedAt: safeExtractDate(data, ['updated_at', 'updatedAt']) || created || FALLBACK_DATE,
    };
    console.log('[DEBUG] Transformed Leave Request:', result);
    return result;
};

export const transformExpenseRequest = (data: any): ExpenseRequest => {
    console.log('[DEBUG] Raw Expense Request:', data);
    const created = safeExtractDate(data, DATE_KEYS);
    const reason = safeExtractString(data, REASON_KEYS);
    const comment = safeExtractString(data, COMMENT_KEYS);

    const result: ExpenseRequest = {
        id: data.id,
        userId: data.user_id,
        userName: data.employee || data.user_name || data.name || data.staff_name,
        amount: data.amount,
        category: data.category,
        reason: reason || 'No reason found',
        status: normalizeStatus(data.status),
        statusReason: comment,
        createdAt: created || FALLBACK_DATE,
        updatedAt: safeExtractDate(data, ['updated_at', 'updatedAt']) || created || FALLBACK_DATE,
    };
    console.log('[DEBUG] Transformed Expense Request:', result);
    return result;
};

export const transformDiscountRequest = (data: any): DiscountRequest => {
    console.log('[DEBUG] Raw Discount Request:', data);
    const created = safeExtractDate(data, DATE_KEYS);
    const reason = safeExtractString(data, REASON_KEYS);
    const comment = safeExtractString(data, COMMENT_KEYS);

    const result: DiscountRequest = {
        id: data.id,
        userId: data.user_id,
        userName: data.employee || data.user_name || data.name || data.staff_name,
        discountPercentage: data.discount_percentage || data.discountPercentage || data.percentage || data.discount || data.percent || 0,
        reason: reason || 'No reason found',
        status: normalizeStatus(data.status),
        statusReason: comment,
        createdAt: created || FALLBACK_DATE,
        updatedAt: safeExtractDate(data, ['updated_at', 'updatedAt']) || created || FALLBACK_DATE,
    };
    console.log('[DEBUG] Transformed Discount Request:', result);
    return result;
};

export const transformHoliday = (data: any): Holiday => ({
    id: data.id,
    date: safeExtractDate(data, ['date']) || FALLBACK_DATE,
    description: data.description,
});

export const transformApprovalRule = (data: any): any => ({
    id: data.id,
    requestType: (data.request_type || data.requestType || 'leave').toLowerCase(),
    condition: data.condition,
    action: data.action,
    priority: data.priority || 1,
    isActive: data.active ?? data.isActive ?? true,
    gradeId: data.grade_id || data.gradeId,
    createdAt: data.timestamp || data.createdAt,
});
