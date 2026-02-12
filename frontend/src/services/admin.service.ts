import api from '@/lib/api';
import { Holiday, ApprovalRule, StatusDistribution, RequestTypeReport } from '@/types';
import { transformHoliday, transformApprovalRule } from '@/lib/transformers';

export const adminService = {
    // Holidays
    async getHolidays(): Promise<Holiday[]> {
        const response = await api.get<any>('/admin/holidays');
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
    async getRules(): Promise<ApprovalRule[]> {
        const response = await api.get<any>('/admin/rules');
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
    }
};
