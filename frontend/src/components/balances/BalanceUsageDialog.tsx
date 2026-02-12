import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { format } from 'date-fns';
import { RequestTypeBadge } from '@/components/ui/status-badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { cn } from '@/lib/utils';
import { Clock } from 'lucide-react';

interface UsageItem {
    id: number;
    date: string;
    description: string;
    amount: string | number;
}

interface BalanceUsageDialogProps {
    isOpen: boolean;
    onClose: () => void;
    title: string;
    type: 'leave' | 'expense' | 'discount';
    usage: UsageItem[];
}

export function BalanceUsageDialog({ isOpen, onClose, title, type, usage }: BalanceUsageDialogProps) {
    return (
        <Dialog open={isOpen} onOpenChange={onClose}>
            <DialogContent className="sm:max-w-md">
                <DialogHeader>
                    <DialogTitle className="flex items-center gap-2">
                        Usage Breakdown: {title}
                    </DialogTitle>
                    <DialogDescription>
                        List of approved requests contributing to your used balance.
                    </DialogDescription>
                </DialogHeader>

                <ScrollArea className="max-h-[400px] pr-4 mt-4">
                    <div className="space-y-3">
                        {usage.length === 0 ? (
                            <div className="text-center py-12 rounded-2xl border-2 border-dashed border-muted bg-muted/10">
                                <Clock className="mx-auto h-8 w-8 text-muted-foreground opacity-20 mb-2" />
                                <p className="text-xs font-bold text-muted-foreground">No approved requests found.</p>
                            </div>
                        ) : (
                            usage.map((item, index) => (
                                <div key={item.id} className="p-4 rounded-2xl bg-muted/20 border border-border/40 hover:bg-muted/30 transition-all">
                                    <div className="flex items-center justify-between">
                                        <div className="space-y-1">
                                            <p className="text-xs font-black text-muted-foreground uppercase tracking-widest">
                                                {format(new Date(item.date), 'MMM d, yyyy')}
                                            </p>
                                            <p className="text-sm font-bold text-foreground leading-none">
                                                {item.description}
                                            </p>
                                        </div>
                                        <div className={cn(
                                            "text-sm font-black px-3 py-1 rounded-full",
                                            type === 'expense' ? "bg-status-approved-bg text-status-approved" :
                                                type === 'discount' ? "bg-discount-bg/20 text-discount" :
                                                    "bg-blue-50 text-blue-700"
                                        )}>
                                            {type === 'expense' ? `$${item.amount.toLocaleString()}` :
                                                type === 'discount' ? `${item.amount}%` :
                                                    `${item.amount} days`}
                                        </div>
                                    </div>
                                </div>
                            ))
                        )}
                    </div>
                </ScrollArea>
            </DialogContent>
        </Dialog>
    );
}
