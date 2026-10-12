import api, { unwrap } from '@/lib/api';
import { UnifiedBalances } from '@/types';

interface RawBucket {
    total: number;
    remaining: number;
}

/** GET /balances: remaining and total per request type. */
export interface RawBalances {
    leave?: RawBucket;
    expense?: RawBucket;
    discount?: RawBucket;
}

export const balanceService = {
    async getUnifiedBalances(): Promise<UnifiedBalances> {
        const response = await api.get<unknown>('/balances');
        const data = unwrap<RawBalances>(response.data);

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
