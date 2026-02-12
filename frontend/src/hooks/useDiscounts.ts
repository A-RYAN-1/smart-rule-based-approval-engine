import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { discountService } from '@/services/discount.service';
import { useAuth } from '@/contexts/AuthContext';
import { toast } from 'sonner';

export function useDiscounts(params?: { myLimit?: number; myOffset?: number; pendingLimit?: number; pendingOffset?: number }) {
    const queryClient = useQueryClient();
    const { user } = useAuth();
    const isManagerOrAdmin = user?.role === 'manager' || user?.role === 'admin';

    const myDiscountsQuery = useQuery({
        queryKey: ['discounts', 'my', params?.myLimit, params?.myOffset],
        queryFn: () => discountService.getMyDiscounts(params?.myLimit || 10, params?.myOffset || 0),
        enabled: !!user && user.role !== 'admin',
    });

    const pendingDiscountsQuery = useQuery({
        queryKey: ['discounts', 'pending', params?.pendingLimit, params?.pendingOffset],
        queryFn: () => discountService.getPendingDiscounts(params?.pendingLimit || 10, params?.pendingOffset || 0),
        enabled: isManagerOrAdmin,
        retry: false,
    });

    const requestDiscountMutation = useMutation({
        mutationFn: discountService.requestDiscount,
        onSuccess: () => {
            toast.success('Discount requested successfully');
            queryClient.invalidateQueries({ queryKey: ['discounts', 'my'] });
        },
        onError: (error: any) => {
            toast.error(error.response?.data?.error || 'Failed to request discount');
        },
    });

    const cancelDiscountMutation = useMutation({
        mutationFn: discountService.cancelDiscount,
        onSuccess: () => {
            toast.success('Discount cancelled');
            queryClient.invalidateQueries({ queryKey: ['discounts', 'my'] });
        },
        onError: () => toast.error('Failed to cancel discount')
    });

    const approveDiscountMutation = useMutation({
        mutationFn: ({ id, comment }: { id: number; comment?: string }) => discountService.approveDiscount(id, comment),
        onSuccess: () => {
            toast.success('Discount approved');
            queryClient.invalidateQueries({ queryKey: ['discounts', 'pending'] });
        },
        onError: () => toast.error('Failed to approve discount')
    });

    const rejectDiscountMutation = useMutation({
        mutationFn: ({ id, comment }: { id: number; comment?: string }) => discountService.rejectDiscount(id, comment),
        onSuccess: () => {
            toast.success('Discount rejected');
            queryClient.invalidateQueries({ queryKey: ['discounts', 'pending'] });
        },
        onError: () => toast.error('Failed to reject discount')
    });

    return {
        myDiscounts: myDiscountsQuery.data?.requests || [],
        myTotal: myDiscountsQuery.data?.total || 0,
        isLoadingMyDiscounts: myDiscountsQuery.isLoading,
        pendingDiscounts: pendingDiscountsQuery.data?.requests || [],
        pendingTotal: pendingDiscountsQuery.data?.total || 0,
        isLoadingPendingDiscounts: pendingDiscountsQuery.isLoading,
        requestDiscount: requestDiscountMutation.mutateAsync,
        cancelDiscount: cancelDiscountMutation.mutateAsync,
        approveDiscount: approveDiscountMutation.mutateAsync,
        rejectDiscount: rejectDiscountMutation.mutateAsync,
    };
}
