export type UserRole = 'employee' | 'manager' | 'admin';

export type RequestStatus = 'pending' | 'approved' | 'rejected' | 'cancelled' | 'auto_approved' | 'auto_rejected';

export type RequestType = 'leave' | 'expense' | 'discount';

export type LeaveType = 'EARN' | 'SICK' | 'CASUAL' | 'PERSONAL' | 'UNPAID';

export interface User {
  id: number;
  name: string;
  email: string;
  role: UserRole;
}

export interface LeaveRequest {
  id: number;
  userId: number;
  userName?: string;
  fromDate: string;
  toDate: string;
  leaveType: LeaveType;
  reason: string;
  status: RequestStatus;
  statusReason?: string;
  createdAt: string;
  updatedAt: string;
}

export interface ExpenseRequest {
  id: number;
  userId: number;
  userName?: string;
  amount: number;
  category: string;
  reason: string;
  status: RequestStatus;
  statusReason?: string;
  createdAt: string;
  updatedAt: string;
}

export interface DiscountRequest {
  id: number;
  userId: number;
  userName?: string;
  discountPercentage: number;
  reason: string;
  status: RequestStatus;
  statusReason?: string;
  createdAt: string;
  updatedAt: string;
}

export interface LeaveBalance {
  leaveType: LeaveType;
  balance: number;
  used: number;
  total: number;
}

export interface GeneralBalance {
  type: RequestType;
  category: string; // e.g. "Travel", "CASUAL", "Discount Pool"
  balance: number;
  used: number;
  total: number;
  unit: 'days' | 'amount' | 'percent';
}

export interface UnifiedBalances {
  leaves: LeaveBalance[];
  expenses: {
    total: number;
    remaining: number;
    used: number;
  };
  discounts: {
    total: number;
    remaining: number;
    used: number;
  };
}

export interface Holiday {
  id: number;
  date: string;
  description: string;
}

export interface ApprovalRule {
  id: number;
  requestType: RequestType;
  condition: any; // Can be object or stringified JSON
  action: 'auto_approve' | 'auto_reject' | 'assign_approver' | 'AUTO_APPROVE' | 'AUTO_REJECT';
  priority: number;
  isActive: boolean;
  gradeId?: number;
  createdAt?: string;
}

export interface StatusDistribution {
  approved: number;
  auto_rejected: number;
  cancelled: number;
  pending: number;
  rejected: number;
  auto_approved: number;
  total_requests: number;
}

export interface RequestTypeReport {
  type: string;
  total_requests: number;
  auto_approved: number;
  auto_approved_percentage: number;
}

export interface DashboardStats {
  totalRequests: number;
  pendingRequests: number;
  approvedRequests: number;
  rejectedRequests: number;
  autoApproved: number;
  autoRejected: number;
}
