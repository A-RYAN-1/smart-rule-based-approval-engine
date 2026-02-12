import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { format } from 'date-fns';
import { StatusBadge, RequestTypeBadge } from '@/components/ui/status-badge';
import { Separator } from '@/components/ui/separator';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import { CalendarDays, Receipt, Percent, Clock, User, MessageSquare, Info } from 'lucide-react';

interface RequestDetailDialogProps {
    isOpen: boolean;
    onClose: () => void;
    request: any;
}

export function RequestDetailDialog({ isOpen, onClose, request }: RequestDetailDialogProps) {
    if (!request) return null;

    const getIcon = () => {
        switch (request.type) {
            case 'leave': return <CalendarDays className="h-5 w-5 text-leave" />;
            case 'expense': return <Receipt className="h-5 w-5 text-expense" />;
            case 'discount': return <Percent className="h-5 w-5 text-discount" />;
            default: return null;
        }
    };

    return (
        <Dialog open={isOpen} onOpenChange={onClose}>
            <DialogContent className="sm:max-w-lg">
                <DialogHeader>
                    <div className="flex items-center gap-3 mb-2">
                        <div className={`p-2 rounded-lg ${request.type === 'leave' ? 'bg-leave-bg' :
                            request.type === 'expense' ? 'bg-expense-bg' : 'bg-discount-bg'
                            }`}>
                            {getIcon()}
                        </div>
                        <div>
                            <DialogTitle className="capitalize">{request.type} Request Details</DialogTitle>
                            <DialogDescription>
                                Submitted on {format(new Date(request.createdAt), 'PPP')}
                            </DialogDescription>
                        </div>
                    </div>
                </DialogHeader>

                <div className="space-y-6 py-4">
                    <div className="grid grid-cols-2 gap-4">
                        <div className={cn(
                            "p-4 rounded-2xl border flex flex-col items-center justify-center text-center transition-all",
                            request.status === 'approved' ? "bg-status-approved-bg/30 border-status-approved/20" :
                                request.status === 'rejected' ? "bg-status-rejected-bg/30 border-status-rejected/20" :
                                    request.status === 'cancelled' ? "bg-muted/30 border-border" : "bg-status-pending-bg/30 border-status-pending/20"
                        )}>
                            <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-2">Status</p>
                            <StatusBadge status={request.status} />
                        </div>
                        <div className="p-4 rounded-2xl border border-border/50 bg-muted/10 flex flex-col items-center justify-center text-center">
                            <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-2">Category</p>
                            <RequestTypeBadge type={request.type} />
                        </div>
                    </div>

                    <div className="bg-card border border-border/40 rounded-2xl overflow-hidden">
                        <div className="px-4 py-3 bg-muted/20 border-b border-border/40 flex items-center gap-2">
                            <Info className="h-3.5 w-3.5 text-muted-foreground" />
                            <span className="text-xs font-bold text-muted-foreground uppercase tracking-wider">Request Information</span>
                        </div>
                        <div className="p-4 space-y-4">
                            {request.type === 'leave' && (
                                <div className="flex flex-col gap-3">
                                    <div className="flex justify-between items-center bg-muted/10 p-2.5 rounded-xl border border-border/30">
                                        <span className="text-xs font-bold text-muted-foreground uppercase tracking-tight">Period</span>
                                        <span className="text-sm font-black">{format(new Date(request.fromDate), 'MMM d')} — {format(new Date(request.toDate), 'MMM d, yyyy')}</span>
                                    </div>
                                    <div className="flex justify-between items-center bg-muted/10 p-2.5 rounded-xl border border-border/30">
                                        <span className="text-xs font-bold text-muted-foreground uppercase tracking-tight">Leave Type</span>
                                        <Badge variant="outline" className="font-bold border-primary/20 text-primary">{request.leaveType}</Badge>
                                    </div>
                                </div>
                            )}

                            {request.type === 'expense' && (
                                <div className="flex flex-col gap-3">
                                    <div className="flex justify-between items-center bg-status-approved-bg/10 p-4 rounded-xl border border-status-approved/20">
                                        <span className="text-xs font-bold text-muted-foreground uppercase tracking-tight text-status-approved">Claim Amount</span>
                                        <span className="text-2xl font-black text-status-approved">${request.amount.toLocaleString()}</span>
                                    </div>
                                    <div className="flex justify-between items-center bg-muted/10 p-2.5 rounded-xl border border-border/30">
                                        <span className="text-xs font-bold text-muted-foreground uppercase tracking-tight">Category</span>
                                        <span className="text-sm font-bold">{request.category}</span>
                                    </div>
                                </div>
                            )}

                            {request.type === 'discount' && (
                                <div className="flex flex-col gap-3">
                                    <div className="flex justify-between items-center bg-discount-bg/10 p-4 rounded-xl border border-discount/20">
                                        <span className="text-xs font-bold text-discount uppercase tracking-tight">Requested Discount</span>
                                        <span className="text-3xl font-black text-discount">{request.discountPercentage}%</span>
                                    </div>
                                </div>
                            )}
                        </div>
                    </div>

                    <div className="space-y-2">
                        <div className="flex items-center gap-2 text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">
                            <MessageSquare className="h-3 w-3" />
                            Reason / Remarks
                        </div>
                        <div className="p-4 rounded-2xl bg-muted/20 border border-dashed border-border/60 text-sm leading-relaxed font-medium italic text-foreground/80">
                            "{request.reason && request.reason !== 'No reason found' ? request.reason : 'No additional notes provided.'}"
                        </div>
                    </div>

                    {request.statusReason && (
                        <div className="p-4 rounded-2xl bg-primary/5 border border-primary/10 space-y-2">
                            <div className="flex items-center gap-2 text-[10px] font-black text-primary uppercase tracking-widest">
                                <Clock className="h-3 w-3" />
                                Response from Admin
                            </div>
                            <p className="text-sm font-bold text-foreground/90">{request.statusReason}</p>
                        </div>
                    )}
                </div>
            </DialogContent>
        </Dialog>
    );
}
