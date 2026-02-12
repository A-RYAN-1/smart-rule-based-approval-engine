import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { expenseService } from '@/services/expense.service';
import { useAuth } from '@/contexts/AuthContext';
import { toast } from 'sonner';

export function useExpenses() {
    const queryClient = useQueryClient();
    const { user } = useAuth();
    const isManagerOrAdmin = user?.role === 'manager' || user?.role === 'admin';

    const myExpensesQuery = useQuery({
        queryKey: ['expenses', 'my'],
        queryFn: expenseService.getMyExpenses,
    });

    const pendingExpensesQuery = useQuery({
        queryKey: ['expenses', 'pending'],
        queryFn: expenseService.getPendingExpenses,
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
        myExpenses: myExpensesQuery.data || [],
        isLoadingMyExpenses: myExpensesQuery.isLoading,
        pendingExpenses: pendingExpensesQuery.data || [],
        isLoadingPendingExpenses: pendingExpensesQuery.isLoading,
        requestExpense: requestExpenseMutation.mutateAsync,
        cancelExpense: cancelExpenseMutation.mutateAsync,
        approveExpense: approveExpenseMutation.mutateAsync,
        rejectExpense: rejectExpenseMutation.mutateAsync,
    };
}
