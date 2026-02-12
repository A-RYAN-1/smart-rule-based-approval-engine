import api from '@/lib/api';
import { DiscountRequest } from '@/types';
import { transformDiscountRequest } from '@/lib/transformers';

export const discountService = {
    async getMyDiscounts(): Promise<DiscountRequest[]> {
        const response = await api.get<any>('/discounts/my');
        const data = response.data.data || [];
        return data.map(transformDiscountRequest);
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

    async getPendingDiscounts(): Promise<DiscountRequest[]> {
        const response = await api.get<any>('/discounts/pending');
        const data = Array.isArray(response.data) ? response.data : (response.data.data || []);
        return data.map(transformDiscountRequest);
    },

    async approveDiscount(id: number, comment?: string): Promise<void> {
        await api.post(`/discounts/${id}/approve`, { comment });
    },

    async rejectDiscount(id: number, comment?: string): Promise<void> {
        await api.post(`/discounts/${id}/reject`, { comment });
    }
};
