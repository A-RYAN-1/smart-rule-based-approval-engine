import { useState } from 'react';
import { 
  LeaveRequest, 
  ExpenseRequest, 
  DiscountRequest, 
  LeaveBalance, 
  Holiday,
  ApprovalRule,
  DashboardStats,
  RequestStatus
} from '@/types';

// Mock data
const mockLeaveRequests: LeaveRequest[] = [
  {
    id: 1,
    userId: 1,
    userName: 'Employee One',
    fromDate: '2026-02-01',
    toDate: '2026-02-03',
    leaveType: 'EARN',
    reason: 'Family function',
    status: 'pending',
    createdAt: '2026-01-20T10:00:00Z',
    updatedAt: '2026-01-20T10:00:00Z',
  },
  {
    id: 2,
    userId: 1,
    userName: 'Employee One',
    fromDate: '2026-01-15',
    toDate: '2026-01-16',
    leaveType: 'SICK',
    reason: 'Medical appointment',
    status: 'approved',
    statusReason: 'Auto-approved: Sick leave under 2 days',
    createdAt: '2026-01-14T09:00:00Z',
    updatedAt: '2026-01-14T09:30:00Z',
  },
  {
    id: 3,
    userId: 1,
    userName: 'Employee One',
    fromDate: '2026-03-10',
    toDate: '2026-03-20',
    leaveType: 'EARN',
    reason: 'Vacation',
    status: 'rejected',
    statusReason: 'Auto-rejected: Leave duration exceeds 7 days without prior approval',
    createdAt: '2026-01-10T14:00:00Z',
    updatedAt: '2026-01-10T14:05:00Z',
  },
];

const mockExpenseRequests: ExpenseRequest[] = [
  {
    id: 1,
    userId: 1,
    userName: 'Employee One',
    amount: 4000,
    category: 'Travel',
    reason: 'Client visit',
    status: 'pending',
    createdAt: '2026-01-22T11:00:00Z',
    updatedAt: '2026-01-22T11:00:00Z',
  },
  {
    id: 2,
    userId: 1,
    userName: 'Employee One',
    amount: 500,
    category: 'Office Supplies',
    reason: 'Keyboard and mouse',
    status: 'approved',
    statusReason: 'Auto-approved: Amount under $1000',
    createdAt: '2026-01-18T15:00:00Z',
    updatedAt: '2026-01-18T15:02:00Z',
  },
];

const mockDiscountRequests: DiscountRequest[] = [
  {
    id: 1,
    userId: 1,
    userName: 'Employee One',
    discountPercentage: 8,
    reason: 'Loan benefit',
    status: 'pending',
    createdAt: '2026-01-23T09:00:00Z',
    updatedAt: '2026-01-23T09:00:00Z',
  },
  {
    id: 2,
    userId: 1,
    userName: 'Employee One',
    discountPercentage: 5,
    reason: 'Employee discount',
    status: 'approved',
    statusReason: 'Auto-approved: Discount under 10%',
    createdAt: '2026-01-19T12:00:00Z',
    updatedAt: '2026-01-19T12:01:00Z',
  },
];

const mockLeaveBalances: LeaveBalance[] = [
  { leaveType: 'EARN', balance: 12, used: 3, total: 15 },
  { leaveType: 'SICK', balance: 8, used: 2, total: 10 },
  { leaveType: 'CASUAL', balance: 5, used: 1, total: 6 },
  { leaveType: 'UNPAID', balance: 10, used: 0, total: 10 },
];

const mockHolidays: Holiday[] = [
  { id: 1, date: '2026-01-01', description: 'New Year\'s Day' },
  { id: 2, date: '2026-01-26', description: 'Republic Day' },
  { id: 3, date: '2026-03-25', description: 'Company Foundation Day' },
  { id: 4, date: '2026-08-15', description: 'Independence Day' },
  { id: 5, date: '2026-12-25', description: 'Christmas' },
];

const mockRules: ApprovalRule[] = [
  {
    id: 1,
    requestType: 'leave',
    condition: 'leave_days <= 2 AND leave_type = "SICK"',
    action: 'auto_approve',
    priority: 1,
    isActive: true,
    createdAt: '2026-01-01T00:00:00Z',
  },
  {
    id: 2,
    requestType: 'leave',
    condition: 'leave_days > 7',
    action: 'auto_reject',
    priority: 2,
    isActive: true,
    createdAt: '2026-01-01T00:00:00Z',
  },
  {
    id: 3,
    requestType: 'expense',
    condition: 'amount < 1000',
    action: 'auto_approve',
    priority: 1,
    isActive: true,
    createdAt: '2026-01-01T00:00:00Z',
  },
  {
    id: 4,
    requestType: 'expense',
    condition: 'amount >= 5000',
    action: 'assign_approver',
    priority: 2,
    isActive: true,
    createdAt: '2026-01-01T00:00:00Z',
  },
  {
    id: 5,
    requestType: 'discount',
    condition: 'discount_percentage <= 10',
    action: 'auto_approve',
    priority: 1,
    isActive: true,
    createdAt: '2026-01-01T00:00:00Z',
  },
];

const mockStats: DashboardStats = {
  totalRequests: 24,
  pendingRequests: 5,
  approvedRequests: 15,
  rejectedRequests: 4,
  autoApproved: 10,
  autoRejected: 2,
};

export function useMockData() {
  const [leaveRequests, setLeaveRequests] = useState(mockLeaveRequests);
  const [expenseRequests, setExpenseRequests] = useState(mockExpenseRequests);
  const [discountRequests, setDiscountRequests] = useState(mockDiscountRequests);
  const [leaveBalances] = useState(mockLeaveBalances);
  const [holidays, setHolidays] = useState(mockHolidays);
  const [rules, setRules] = useState(mockRules);
  const [stats] = useState(mockStats);

  const updateRequestStatus = (
    type: 'leave' | 'expense' | 'discount',
    id: number,
    status: RequestStatus,
    reason?: string
  ) => {
    if (type === 'leave') {
      setLeaveRequests(prev => prev.map(r => 
        r.id === id ? { ...r, status, statusReason: reason, updatedAt: new Date().toISOString() } : r
      ));
    } else if (type === 'expense') {
      setExpenseRequests(prev => prev.map(r => 
        r.id === id ? { ...r, status, statusReason: reason, updatedAt: new Date().toISOString() } : r
      ));
    } else {
      setDiscountRequests(prev => prev.map(r => 
        r.id === id ? { ...r, status, statusReason: reason, updatedAt: new Date().toISOString() } : r
      ));
    }
  };

  const addLeaveRequest = (request: Omit<LeaveRequest, 'id' | 'createdAt' | 'updatedAt'>) => {
    const newRequest: LeaveRequest = {
      ...request,
      id: leaveRequests.length + 1,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
    setLeaveRequests(prev => [newRequest, ...prev]);
    return newRequest;
  };

  const addExpenseRequest = (request: Omit<ExpenseRequest, 'id' | 'createdAt' | 'updatedAt'>) => {
    const newRequest: ExpenseRequest = {
      ...request,
      id: expenseRequests.length + 1,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
    setExpenseRequests(prev => [newRequest, ...prev]);
    return newRequest;
  };

  const addDiscountRequest = (request: Omit<DiscountRequest, 'id' | 'createdAt' | 'updatedAt'>) => {
    const newRequest: DiscountRequest = {
      ...request,
      id: discountRequests.length + 1,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
    setDiscountRequests(prev => [newRequest, ...prev]);
    return newRequest;
  };

  const addHoliday = (holiday: Omit<Holiday, 'id'>) => {
    const newHoliday: Holiday = {
      ...holiday,
      id: holidays.length + 1,
    };
    setHolidays(prev => [...prev, newHoliday]);
    return newHoliday;
  };

  const deleteHoliday = (id: number) => {
    setHolidays(prev => prev.filter(h => h.id !== id));
  };

  const addRule = (rule: Omit<ApprovalRule, 'id' | 'createdAt'>) => {
    const newRule: ApprovalRule = {
      ...rule,
      id: rules.length + 1,
      createdAt: new Date().toISOString(),
    };
    setRules(prev => [...prev, newRule]);
    return newRule;
  };

  const toggleRule = (id: number) => {
    setRules(prev => prev.map(r => 
      r.id === id ? { ...r, isActive: !r.isActive } : r
    ));
  };

  const deleteRule = (id: number) => {
    setRules(prev => prev.filter(r => r.id !== id));
  };

  return {
    leaveRequests,
    expenseRequests,
    discountRequests,
    leaveBalances,
    holidays,
    rules,
    stats,
    updateRequestStatus,
    addLeaveRequest,
    addExpenseRequest,
    addDiscountRequest,
    addHoliday,
    deleteHoliday,
    addRule,
    toggleRule,
    deleteRule,
  };
}
