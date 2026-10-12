import api, { unwrap, RawPage, RawRecord } from '@/lib/api';
import { ExpenseRequest } from '@/types';
import { transformExpenseRequest } from '@/lib/transformers';

export const expenseService = {
    async getMyExpenses(limit = 10, offset = 0): Promise<{ requests: ExpenseRequest[]; total: number }> {
        const response = await api.get<unknown>(`/expenses/my?limit=${limit}&offset=${offset}`);
        const result = unwrap<RawPage>(response.data);
        const requests = (result.requests || []).map(transformExpenseRequest);
        return { requests, total: result.total || 0 };
    },

    async requestExpense(payload: { amount: number; category: string; reason: string }): Promise<ExpenseRequest> {
        const response = await api.post<unknown>('/expenses/request', payload);
        return transformExpenseRequest(unwrap<RawRecord>(response.data));
    },

    async cancelExpense(id: number): Promise<void> {
        await api.post(`/expenses/${id}/cancel`);
    },

    async getPendingExpenses(limit = 10, offset = 0): Promise<{ requests: ExpenseRequest[]; total: number }> {
        const response = await api.get<unknown>(`/expenses/pending?limit=${limit}&offset=${offset}`);
        const result = unwrap<RawPage>(response.data);
        const requests = (result.requests || []).map(transformExpenseRequest);
        return { requests, total: result.total || 0 };
    },

    async approveExpense(id: number, comment?: string): Promise<void> {
        await api.post(`/expenses/${id}/approve`, { comment });
    },

    async rejectExpense(id: number, comment?: string): Promise<void> {
        await api.post(`/expenses/${id}/reject`, { comment });
    }
};
