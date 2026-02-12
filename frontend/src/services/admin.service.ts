import api from '@/lib/api';
import { Holiday, ApprovalRule, StatusDistribution, RequestTypeReport, LeaveRequest, ExpenseRequest, DiscountRequest } from '@/types';
import { transformHoliday, transformApprovalRule, transformLeaveRequest, transformExpenseRequest, transformDiscountRequest } from '@/lib/transformers';

export const adminService = {
    // Holidays
    async getHolidays(limit = 10, offset = 0): Promise<Holiday[]> {
        const response = await api.get<any>(`/admin/holidays?limit=${limit}&offset=${offset}`);
        const data = response.data.data || response.data;
        return Array.isArray(data) ? data.map(transformHoliday) : [];
    },

    async addHoliday(payload: { date: string; description: string }): Promise<void> {
        await api.post<any>('/admin/holidays', payload);
    },

    async deleteHoliday(id: number): Promise<void> {
        await api.delete(`/admin/holidays/${id}`);
    },

    // Rules
    async getRules(limit = 10, offset = 0): Promise<ApprovalRule[]> {
        const response = await api.get<any>(`/admin/rules?limit=${limit}&offset=${offset}`);
        const data = response.data.data || response.data;
        return Array.isArray(data) ? data.map(transformApprovalRule) : [];
    },

    async addRule(rule: Partial<ApprovalRule>): Promise<void> {
        const apiPayload = {
            request_type: (rule.requestType || 'LEAVE').toUpperCase(),
            condition: rule.condition,
            action: rule.action,
            grade_id: rule.gradeId || 1,
            active: rule.isActive ?? true
        };
        await api.post('/admin/rules', apiPayload);
    },

    async updateRule(id: number, rule: Partial<ApprovalRule>): Promise<void> {
        const apiPayload = {
            request_type: (rule.requestType || 'LEAVE').toUpperCase(),
            condition: rule.condition,
            action: rule.action,
            grade_id: rule.gradeId || 1,
            active: rule.isActive ?? true
        };
        await api.put(`/admin/rules/${id}`, apiPayload);
    },

    async deleteRule(id: number): Promise<void> {
        await api.delete(`/admin/rules/${id}`);
    },

    // Reports
    async getDashboardSummary(): Promise<any> {
        const response = await api.get<any>('/reports/dashboard');
        return response.data.data || response.data;
    },

    async getStatusDistribution(): Promise<StatusDistribution> {
        const response = await api.get<any>('/admin/reports/request-status-distribution');
        return response.data.data || response.data;
    },

    async getRequestsByType(): Promise<RequestTypeReport[]> {
        const response = await api.get<any>('/admin/reports/requests-by-type');
        const data = response.data.data || response.data;
        return Array.isArray(data) ? data : [];
    },

    // System
    async runAutoReject(): Promise<void> {
        await api.post('/system/run-auto-reject');
    },

    async getPendingAllRequests(limit = 10, offset = 0): Promise<{
        leaves: LeaveRequest[];
        expenses: ExpenseRequest[];
        discounts: DiscountRequest[];
        total: number;
    }> {
        const response = await api.get<any>(`/pending/all?limit=${limit}&offset=${offset}`);
        const data = response.data.data || response.data;
        return {
            leaves: (data.leave_request || []).map(transformLeaveRequest),
            expenses: (data.expense_request || []).map(transformExpenseRequest),
            discounts: (data.discount_request || []).map(transformDiscountRequest),
            total: data.total || 0
        };
    },

    async getMyAllRequests(limit = 10, offset = 0): Promise<{
        leaves: LeaveRequest[];
        expenses: ExpenseRequest[];
        discounts: DiscountRequest[];
        total: number;
    }> {
        const response = await api.get<any>(`/my-requests/all?limit=${limit}&offset=${offset}`);
        const data = response.data.data || response.data;
        return {
            leaves: (data.leave_request || []).map(transformLeaveRequest),
            expenses: (data.expense_request || []).map(transformExpenseRequest),
            discounts: (data.discount_request || []).map(transformDiscountRequest),
            total: data.total || 0
        };
    }
};
