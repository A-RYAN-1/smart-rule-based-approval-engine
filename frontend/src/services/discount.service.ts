import api from '@/lib/api';
import { DiscountRequest } from '@/types';
import { transformDiscountRequest } from '@/lib/transformers';

export const discountService = {
    async getMyDiscounts(limit = 10, offset = 0): Promise<{ requests: DiscountRequest[]; total: number }> {
        const response = await api.get<any>(`/discounts/my?limit=${limit}&offset=${offset}`);
        const result = response.data.data || response.data;
        const requests = (result.requests || []).map(transformDiscountRequest);
        return { requests, total: result.total || 0 };
    },

    async requestDiscount(payload: { discountPercentage: number; reason: string }): Promise<DiscountRequest> {
        const apiPayload = {
            discount_percentage: payload.discountPercentage,
            reason: payload.reason,
        };
        const response = await api.post<any>('/discounts/request', apiPayload);
        return transformDiscountRequest(response.data.data || response.data);
    },

    async cancelDiscount(id: number): Promise<void> {
        await api.post(`/discounts/${id}/cancel`);
    },

    async getPendingDiscounts(limit = 10, offset = 0): Promise<{ requests: DiscountRequest[]; total: number }> {
        const response = await api.get<any>(`/discounts/pending?limit=${limit}&offset=${offset}`);
        const result = response.data.data || response.data;
        const requests = (result.requests || []).map(transformDiscountRequest);
        return { requests, total: result.total || 0 };
    },

    async approveDiscount(id: number, comment?: string): Promise<void> {
        await api.post(`/discounts/${id}/approve`, { comment });
    },

    async rejectDiscount(id: number, comment?: string): Promise<void> {
        await api.post(`/discounts/${id}/reject`, { comment });
    }
};
