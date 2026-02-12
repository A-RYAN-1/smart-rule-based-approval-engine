import { useQuery } from '@tanstack/react-query';
import { balanceService } from '@/services/balance.service';

export function useBalances() {
    const { data, isLoading, error, refetch } = useQuery({
        queryKey: ['balances', 'unified'],
        queryFn: balanceService.getUnifiedBalances,
        staleTime: 1000 * 60 * 5, // 5 minutes
    });

    return {
        balances: data,
        isLoading,
        error,
        refetchBalances: refetch,
    };
}
