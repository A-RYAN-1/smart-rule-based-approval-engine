import api from '@/lib/api';
import { UnifiedBalances } from '@/types';

export const balanceService = {
    async getUnifiedBalances(): Promise<UnifiedBalances> {
        const response = await api.get<any>('/balances');
        const data = response.data.data || response.data;

        // Map the backend data to UnifiedBalances schema
        return {
            leaves: data.leave ? [{
                leaveType: 'EARN', // Primary leave type as indicated by user
                balance: data.leave.remaining,
                used: data.leave.total - data.leave.remaining,
                total: data.leave.total
            }] : [],
            expenses: {
                total: data.expense?.total || 0,
                remaining: data.expense?.remaining || 0,
                used: (data.expense?.total || 0) - (data.expense?.remaining || 0)
            },
            discounts: {
                total: data.discount?.total || 0,
                remaining: data.discount?.remaining || 0,
                used: (data.discount?.total || 0) - (data.discount?.remaining || 0)
            }
        };
    }
};
