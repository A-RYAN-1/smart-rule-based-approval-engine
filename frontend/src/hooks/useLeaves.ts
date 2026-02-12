import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { leaveService } from '@/services/leave.service';
import { useAuth } from '@/contexts/AuthContext';
import { toast } from 'sonner';

export function useLeaves() {
    const queryClient = useQueryClient();
    const { user } = useAuth();
    const isManagerOrAdmin = user?.role === 'manager' || user?.role === 'admin';

    const myLeavesQuery = useQuery({
        queryKey: ['leaves', 'my'],
        queryFn: leaveService.getMyLeaves,
    });

    const pendingLeavesQuery = useQuery({
        queryKey: ['leaves', 'pending'],
        queryFn: leaveService.getPendingLeaves,
        // Only fetch if we are manager/admin
        enabled: isManagerOrAdmin,
        retry: false,
    });

    const balancesQuery = useQuery({
        queryKey: ['balances'],
        queryFn: leaveService.getBalances,
    });

    const requestLeaveMutation = useMutation({
        mutationFn: leaveService.requestLeave,
        onSuccess: () => {
            toast.success('Leave requested successfully');
            queryClient.invalidateQueries({ queryKey: ['leaves', 'my'] });
            queryClient.invalidateQueries({ queryKey: ['balances'] });
        },
        onError: (error: any) => {
            const errorMessage = error.response?.data?.message || error.response?.data?.error || 'Failed to request leave';
            // Check for overlapping request error
            if (errorMessage.toLowerCase().includes('overlap') || errorMessage.toLowerCase().includes('conflict')) {
                toast.error('Overlapping Leave Request', {
                    description: errorMessage,
                });
            } else {
                toast.error(errorMessage);
            }
        },
    });

    const cancelLeaveMutation = useMutation({
        mutationFn: leaveService.cancelLeave,
        onSuccess: () => {
            toast.success('Leave cancelled');
            queryClient.invalidateQueries({ queryKey: ['leaves', 'my'] });
            queryClient.invalidateQueries({ queryKey: ['balances'] });
        },
        onError: (error: any) => {
            toast.error('Failed to cancel leave');
        }
    });

    const approveLeaveMutation = useMutation({
        mutationFn: ({ id, comment }: { id: number; comment?: string }) => leaveService.approveLeave(id, comment),
        onSuccess: () => {
            toast.success('Leave approved');
            queryClient.invalidateQueries({ queryKey: ['leaves', 'pending'] });
        },
        onError: () => toast.error('Failed to approve leave')
    });

    const rejectLeaveMutation = useMutation({
        mutationFn: ({ id, comment }: { id: number; comment?: string }) => leaveService.rejectLeave(id, comment),
        onSuccess: () => {
            toast.success('Leave rejected');
            queryClient.invalidateQueries({ queryKey: ['leaves', 'pending'] });
        },
        onError: () => toast.error('Failed to reject leave')
    });

    return {
        myLeaves: myLeavesQuery.data || [],
        isLoadingMyLeaves: myLeavesQuery.isLoading,
        pendingLeaves: pendingLeavesQuery.data || [],
        isLoadingPendingLeaves: pendingLeavesQuery.isLoading,
        balances: balancesQuery.data || [],
        requestLeave: requestLeaveMutation.mutateAsync,
        cancelLeave: cancelLeaveMutation.mutateAsync,
        approveLeave: approveLeaveMutation.mutateAsync,
        rejectLeave: rejectLeaveMutation.mutateAsync,
    };
}
