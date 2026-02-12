import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminService } from '@/services/admin.service';
import { useAuth } from '@/contexts/AuthContext';
import { toast } from 'sonner';
import { ApprovalRule } from '@/types';

export function useAdmin() {
    const queryClient = useQueryClient();
    const { user } = useAuth();
    const isApprover = user?.role === 'admin' || user?.role === 'manager';
    const isAdmin = user?.role === 'admin';

    // Dashboard Summary
    const dashboardSummaryQuery = useQuery({
        queryKey: ['dashboard-summary'],
        queryFn: adminService.getDashboardSummary,
        enabled: isAdmin,
    });

    // Holidays
    const holidaysQuery = useQuery({
        queryKey: ['holidays'],
        queryFn: ({ queryKey }) => adminService.getHolidays(10, 0), // Default for now
        enabled: isApprover,
    });

    const addHolidayMutation = useMutation({
        mutationFn: adminService.addHoliday,
        onSuccess: () => {
            toast.success('Holiday added');
            queryClient.invalidateQueries({ queryKey: ['holidays'] });
        },
        onError: () => toast.error('Failed to add holiday')
    });

    const deleteHolidayMutation = useMutation({
        mutationFn: adminService.deleteHoliday,
        onSuccess: () => {
            toast.success('Holiday deleted');
            queryClient.invalidateQueries({ queryKey: ['holidays'] });
        },
        onError: () => toast.error('Failed to delete holiday')
    });

    // Rules
    const rulesQuery = useQuery({
        queryKey: ['admin-rules'],
        queryFn: ({ queryKey }) => adminService.getRules(20, 0), // Default higher limit for rules
        enabled: isApprover,
    });

    const addRuleMutation = useMutation({
        mutationFn: adminService.addRule,
        onSuccess: () => {
            toast.success('Rule added');
            queryClient.invalidateQueries({ queryKey: ['admin-rules'] });
        },
        onError: () => toast.error('Failed to add rule')
    });

    const updateRuleMutation = useMutation({
        mutationFn: ({ id, rule }: { id: number; rule: Partial<ApprovalRule> }) => adminService.updateRule(id, rule),
        onSuccess: () => {
            toast.success('Rule updated');
            queryClient.invalidateQueries({ queryKey: ['admin-rules'] });
        },
        onError: () => toast.error('Failed to update rule')
    });

    const deleteRuleMutation = useMutation({
        mutationFn: adminService.deleteRule,
        onSuccess: () => {
            toast.success('Rule deleted');
            queryClient.invalidateQueries({ queryKey: ['admin-rules'] });
        },
        onError: () => toast.error('Failed to delete rule')
    });

    const toggleRuleActive = async (rule: ApprovalRule) => {
        return updateRuleMutation.mutateAsync({
            id: rule.id,
            rule: { ...rule, isActive: !rule.isActive }
        });
    };

    // Reports - Keep for detail pages, but dashboard uses summary
    const statusDistributionQuery = useQuery({
        queryKey: ['report-status'],
        queryFn: adminService.getStatusDistribution,
        enabled: isApprover && !isAdmin, // Only if not admin, or if specific report page needs it
    });

    const requestsByTypeQuery = useQuery({
        queryKey: ['report-type'],
        queryFn: adminService.getRequestsByType,
        enabled: isApprover && !isAdmin,
    });

    const runAutoRejectMutation = useMutation({
        mutationFn: adminService.runAutoReject,
        onSuccess: () => {
            toast.success('Auto-reject system process triggered');
            queryClient.invalidateQueries({ queryKey: ['admin-rules'] });
            queryClient.invalidateQueries({ queryKey: ['report-status'] });
            queryClient.invalidateQueries({ queryKey: ['dashboard-summary'] });
        },
        onError: () => toast.error('Failed to trigger auto-reject process')
    });

    const usePendingAll = (limit: number, offset: number) => {
        return useQuery({
            queryKey: ['pending-all', limit, offset],
            queryFn: () => adminService.getPendingAllRequests(limit, offset),
            enabled: isApprover,
        });
    };

    const useMyAll = (limit: number, offset: number) => {
        return useQuery({
            queryKey: ['my-all', limit, offset],
            queryFn: () => adminService.getMyAllRequests(limit, offset),
            enabled: !!user,
        });
    };

    return {
        // Dashboard
        dashboardSummary: dashboardSummaryQuery.data,
        isLoadingDashboard: dashboardSummaryQuery.isLoading,

        // Holidays
        holidays: holidaysQuery.data || [],
        isLoadingHolidays: holidaysQuery.isLoading,
        addHoliday: addHolidayMutation.mutateAsync,
        deleteHoliday: deleteHolidayMutation.mutateAsync,

        // Rules
        rules: rulesQuery.data || [],
        isLoadingRules: rulesQuery.isLoading,
        addRule: addRuleMutation.mutateAsync,
        updateRule: updateRuleMutation.mutateAsync,
        deleteRule: deleteRuleMutation.mutateAsync,
        toggleRule: toggleRuleActive,

        // Reports (Legacy/Detailed)
        statusDistribution: dashboardSummaryQuery.data?.distribution || statusDistributionQuery.data,
        isLoadingStatusDistribution: dashboardSummaryQuery.isLoading || statusDistributionQuery.isLoading,
        requestsByType: dashboardSummaryQuery.data?.type_report || requestsByTypeQuery.data || [],
        isLoadingRequestsByType: dashboardSummaryQuery.isLoading || requestsByTypeQuery.isLoading,

        // System
        runAutoReject: runAutoRejectMutation.mutateAsync,
        isRunningAutoReject: runAutoRejectMutation.isPending,

        // Unified Pending
        usePendingAll,
        useMyAll,
    };
}
