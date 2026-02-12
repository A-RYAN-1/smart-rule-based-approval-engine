import { useQuery } from '@tanstack/react-query';
import { balanceService } from '@/services/balance.service';
import { useAuth } from '@/contexts/AuthContext';

export function useBalances() {
    const { user } = useAuth();

    const { data, isLoading, error, refetch } = useQuery({
        queryKey: ['balances', 'unified'],
        queryFn: balanceService.getUnifiedBalances,
        staleTime: 1000 * 60 * 5, // 5 minutes
        enabled: !!user && user.role !== 'admin',
    });

    return {
        balances: data,
        isLoading,
        error,
        refetchBalances: refetch,
    };
}
