
import api from '@/lib/api';
import { LeaveRequest, LeaveBalance } from '@/types';
import { transformLeaveRequest } from '@/lib/transformers';

// API request interface (snake_case)
interface CreateLeaveRequestPayload {
    from_date: string;
    to_date: string;
    leave_type: string;
    reason: string;
}

export const leaveService = {
    async getMyLeaves(): Promise<LeaveRequest[]> {
        const response = await api.get<any>('/leaves/my');
        const data = response.data.data || [];
        return data.map(transformLeaveRequest);
    },

    async requestLeave(payload: { fromDate: string; toDate: string; leaveType: string; reason: string }): Promise<LeaveRequest> {
        const apiPayload = {
            from_date: payload.fromDate,
            to_date: payload.toDate,
            leave_type: payload.leaveType,
            reason: payload.reason,
        };
        const response = await api.post<any>('/leaves/request', apiPayload);
        return transformLeaveRequest(response.data.data || response.data);
    },

    async cancelLeave(id: number): Promise<void> {
        await api.post(`/leaves/${id}/cancel`);
    },

    // Manager/Admin only
    async getPendingLeaves(): Promise<LeaveRequest[]> {
        const response = await api.get<any>('/leaves/pending');
        // Check if response is array or wrapped
        const data = Array.isArray(response.data) ? response.data : (response.data.data || []);
        return data.map(transformLeaveRequest);
    },

    async approveLeave(id: number, comment?: string): Promise<void> {
        await api.post(`/leaves/${id}/approve`, { comment });
    },

    async rejectLeave(id: number, comment?: string): Promise<void> {
        await api.post(`/leaves/${id}/reject`, { comment });
    },

    async getBalances(): Promise<LeaveBalance[]> {
        const response = await api.get<any>('/balances');
        const data = response.data.data || response.data; // { discount: {...}, expense: {...}, leave: {...} }

        // Transform object to array expected by UI
        // UI expects LeaveBalance[]: { leaveType: 'EARN', balance: 20, used: 10, total: 30 }
        // But backend sends consolidated totals by category, NOT by leave type (EARN, CASUAL etc) specifically for leaves?
        // Wait, response was: "leave": { "remaining": 20, "total": 30 }.
        // This implies aggregated leave balance? Or does backend not track separate types?
        // User said "types of leave are only EARN/CASUAL/PERSONAL".
        // If backend only gives total leave, we might need to mock breakdown or just show Total.
        // Let's map "leave" to a generic type or map keys if they exist in a more detailed object.
        // Based on log: "leave": { "remaining": 20, "total": 30 }.
        // We will create a generic "Leaves" entry for now.

        const balances: LeaveBalance[] = [];

        if (data.leave) {
            balances.push({
                leaveType: 'EARN', // Mapping generic 'leave' bucket to 'EARN' for display as primary
                balance: data.leave.remaining,
                used: data.leave.total - data.leave.remaining,
                total: data.leave.total
            });
        }

        // Note: The UI Dashboard shows "Leave Balances".
        // Expense and Discount "balances" (limits) are not standard "LeaveBalance", but we might want to show them?
        // The type `LeaveBalance` implies leaves.
        // We will return just the leave mapping for now.

        return balances;
    }
};

