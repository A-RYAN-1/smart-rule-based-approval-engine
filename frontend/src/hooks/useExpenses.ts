import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { expenseService } from '@/services/expense.service';
import { useAuth } from '@/contexts/AuthContext';
import { toast } from 'sonner';

export function useExpenses(params?: { myLimit?: number; myOffset?: number; pendingLimit?: number; pendingOffset?: number }) {
    const queryClient = useQueryClient();
    const { user } = useAuth();
    const isManagerOrAdmin = user?.role === 'manager' || user?.role === 'admin';

    const myExpensesQuery = useQuery({
        queryKey: ['expenses', 'my', params?.myLimit, params?.myOffset],
        queryFn: () => expenseService.getMyExpenses(params?.myLimit || 10, params?.myOffset || 0),
        enabled: !!user && user.role !== 'admin',
    });

    const pendingExpensesQuery = useQuery({
        queryKey: ['expenses', 'pending', params?.pendingLimit, params?.pendingOffset],
        queryFn: () => expenseService.getPendingExpenses(params?.pendingLimit || 10, params?.pendingOffset || 0),
        enabled: isManagerOrAdmin,
        retry: false,
    });

    const requestExpenseMutation = useMutation({
        mutationFn: expenseService.requestExpense,
        onSuccess: () => {
            toast.success('Expense requested successfully');
            queryClient.invalidateQueries({ queryKey: ['expenses', 'my'] });
        },
        onError: (error: any) => {
            toast.error(error.response?.data?.error || 'Failed to request expense');
        },
    });

    const cancelExpenseMutation = useMutation({
        mutationFn: expenseService.cancelExpense,
        onSuccess: () => {
            toast.success('Expense cancelled');
            queryClient.invalidateQueries({ queryKey: ['expenses', 'my'] });
        },
        onError: () => toast.error('Failed to cancel expense')
    });

    const approveExpenseMutation = useMutation({
        mutationFn: ({ id, comment }: { id: number; comment?: string }) => expenseService.approveExpense(id, comment),
        onSuccess: () => {
            toast.success('Expense approved');
            queryClient.invalidateQueries({ queryKey: ['expenses', 'pending'] });
        },
        onError: () => toast.error('Failed to approve expense')
    });

    const rejectExpenseMutation = useMutation({
        mutationFn: ({ id, comment }: { id: number; comment?: string }) => expenseService.rejectExpense(id, comment),
        onSuccess: () => {
            toast.success('Expense rejected');
            queryClient.invalidateQueries({ queryKey: ['expenses', 'pending'] });
        },
        onError: () => toast.error('Failed to reject expense')
    });

    return {
        myExpenses: myExpensesQuery.data?.requests || [],
        myTotal: myExpensesQuery.data?.total || 0,
        isLoadingMyExpenses: myExpensesQuery.isLoading,
        pendingExpenses: pendingExpensesQuery.data?.requests || [],
        pendingTotal: pendingExpensesQuery.data?.total || 0,
        isLoadingPendingExpenses: pendingExpensesQuery.isLoading,
        requestExpense: requestExpenseMutation.mutateAsync,
        cancelExpense: cancelExpenseMutation.mutateAsync,
        approveExpense: approveExpenseMutation.mutateAsync,
        rejectExpense: rejectExpenseMutation.mutateAsync,
    };
}
