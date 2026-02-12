import { useState, useMemo } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { cn } from '@/lib/utils';
import { useLeaves } from '@/hooks/useLeaves';
import { useExpenses } from '@/hooks/useExpenses';
import { useDiscounts } from '@/hooks/useDiscounts';
import { useAuth } from '@/contexts/AuthContext';
import { useAdmin } from '@/hooks/useAdmin';
import { Navigate } from 'react-router-dom';
import { StatusBadge, RequestTypeBadge } from '@/components/ui/status-badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { useToast } from '@/hooks/use-toast';
import { format } from 'date-fns';
import { Eye, X, Calendar, DollarSign, Percent, Info } from 'lucide-react';
import { RequestType, RequestStatus } from '@/types';

export default function MyRequestsPage() {
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState<'all' | RequestType>('all');
  const [statusFilter, setStatusFilter] = useState<'all' | RequestStatus>('all');
  const [currentPage, setCurrentPage] = useState(1);
  const ITEMS_PER_PAGE = 8;
  const offset = (currentPage - 1) * ITEMS_PER_PAGE;

  const { useMyAll } = useAdmin();

  const { myLeaves, myTotal: leaveTotal, cancelLeave } = useLeaves(
    activeTab === 'leave' || activeTab === 'all' ? { myLimit: ITEMS_PER_PAGE, myOffset: offset } : undefined
  );
  const { myExpenses, myTotal: expenseTotal, cancelExpense } = useExpenses(
    activeTab === 'expense' || activeTab === 'all' ? { myLimit: ITEMS_PER_PAGE, myOffset: offset } : undefined
  );
  const { myDiscounts, myTotal: discountTotal, cancelDiscount } = useDiscounts(
    activeTab === 'discount' || activeTab === 'all' ? { myLimit: ITEMS_PER_PAGE, myOffset: offset } : undefined
  );

  const myAllQuery = useMyAll(ITEMS_PER_PAGE, offset);
  const myAllData = myAllQuery.data;

  const { toast } = useToast();
  const [selectedRequest, setSelectedRequest] = useState<any>(null);

  const [confirmCancel, setConfirmCancel] = useState<{
    type: RequestType;
    id: number;
  } | null>(null);


  const filteredRequests = useMemo(() => {
    let baseRequests: any[] = [];
    if (activeTab === 'all') {
      baseRequests = [
        ...(myAllData?.leaves || []).map(r => ({ ...r, type: 'leave' as const })),
        ...(myAllData?.expenses || []).map(r => ({ ...r, type: 'expense' as const })),
        ...(myAllData?.discounts || []).map(r => ({ ...r, type: 'discount' as const })),
      ].sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
    } else if (activeTab === 'leave') {
      baseRequests = myLeaves.map(r => ({ ...r, type: 'leave' as const }));
    } else if (activeTab === 'expense') {
      baseRequests = myExpenses.map(r => ({ ...r, type: 'expense' as const }));
    } else if (activeTab === 'discount') {
      baseRequests = myDiscounts.map(r => ({ ...r, type: 'discount' as const }));
    }

    return baseRequests.filter(r => {
      // Group auto_approved with approved, auto_rejected with rejected
      let statusMatch = false;
      if (statusFilter === 'all') {
        statusMatch = true;
      } else if (statusFilter === 'approved') {
        statusMatch = r.status === 'approved' || r.status === 'auto_approved';
      } else if (statusFilter === 'rejected') {
        statusMatch = r.status === 'rejected' || r.status === 'auto_rejected';
      } else {
        statusMatch = r.status === statusFilter;
      }

      return statusMatch;
    });
  }, [myAllData, myLeaves, myExpenses, myDiscounts, activeTab, statusFilter]);

  const handleCancel = async (type: RequestType, id: number) => {
    try {
      if (type === 'leave') await cancelLeave(id);
      else if (type === 'expense') await cancelExpense(id);
      else if (type === 'discount') await cancelDiscount(id);
      setSelectedRequest(null);
    } catch (error) {
      console.error(error);
    }
  };

  const getRequestDetails = (request: any) => {
    if (request.type === 'leave') {
      return `${request.leaveType} • ${format(new Date(request.fromDate), 'MMM d')} - ${format(new Date(request.toDate), 'MMM d, yyyy')}`;
    } else if (request.type === 'expense') {
      return `${request.category} • $${request.amount.toLocaleString()}`;
    } else {
      return `${request.discountPercentage}% discount`;
    }
  };

  const getRequestIcon = (type: RequestType) => {
    switch (type) {
      case 'leave': return <Calendar className="h-4 w-4" />;
      case 'expense': return <DollarSign className="h-4 w-4" />;
      case 'discount': return <Percent className="h-4 w-4" />;
    }
  };

  const currentTotal = useMemo(() => {
    if (activeTab === 'all') return myAllData?.total || 0;
    if (activeTab === 'leave') return leaveTotal;
    if (activeTab === 'expense') return expenseTotal;
    if (activeTab === 'discount') return discountTotal;
    return 0;
  }, [activeTab, myAllData?.total, leaveTotal, expenseTotal, discountTotal]);

  const totalPages = Math.ceil(currentTotal / ITEMS_PER_PAGE);
  const paginatedRequests = filteredRequests; // Already paginated from hooks

  const { user: authUser } = useAuth(); // for usage if needed

  return (
    <AppLayout>
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">My Requests</h1>
          <p className="text-muted-foreground mt-1">View and manage all your submitted requests</p>
        </div>

        {/* Stats Summary */}
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4">
          <Card>
            <CardContent className="p-4 text-center">
              <p className="text-2xl font-bold">{leaveTotal + expenseTotal + discountTotal}</p>
              <p className="text-sm text-muted-foreground">Total Requests</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4 text-center">
              <p className="text-2xl font-bold text-status-pending">
                {leaveTotal + expenseTotal + discountTotal}
              </p>
              <p className="text-sm text-muted-foreground">Managed by API Pagination</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4 text-center">
              <p className="text-2xl font-bold text-status-approved">
                {leaveTotal}
              </p>
              <p className="text-sm text-muted-foreground">Leaves</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4 text-center">
              <p className="text-2xl font-bold text-status-rejected">
                {expenseTotal}
              </p>
              <p className="text-sm text-muted-foreground">Expenses</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4 text-center">
              <p className="text-2xl font-bold text-muted-foreground">
                {discountTotal}
              </p>
              <p className="text-sm text-muted-foreground">Discounts</p>
            </CardContent>
          </Card>
        </div>

        {/* Filters and Search Area */}
        <div className="flex flex-col gap-6">
          <Card className="border-none shadow-none bg-transparent">
            <CardContent className="p-0">
              <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-6">
                <Tabs value={activeTab} onValueChange={(v) => { setActiveTab(v as any); setStatusFilter('all'); setCurrentPage(1); }} className="w-full lg:w-auto">
                  <TabsList className="bg-muted/50 p-1 h-11 border border-border/50">
                    <TabsTrigger value="all" className="px-6 data-[state=active]:bg-background data-[state=active]:shadow-sm">All Types</TabsTrigger>
                    <TabsTrigger value="leave" className="px-6 data-[state=active]:bg-background data-[state=active]:shadow-sm">Leaves</TabsTrigger>
                    <TabsTrigger value="expense" className="px-6 data-[state=active]:bg-background data-[state=active]:shadow-sm">Expenses</TabsTrigger>
                    <TabsTrigger value="discount" className="px-6 data-[state=active]:bg-background data-[state=active]:shadow-sm">Discounts</TabsTrigger>
                  </TabsList>
                </Tabs>

                <div className="flex flex-col gap-2">
                  {/* <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground px-1">Status Filter</p> */}
                  <Tabs value={statusFilter} onValueChange={(v) => { setStatusFilter(v as any); setCurrentPage(1); }} className="w-full lg:w-auto">
                    <TabsList className="bg-muted/50 p-1 h-11 border border-border/50">
                      <TabsTrigger value="all" className="px-5 data-[state=active]:bg-background data-[state=active]:shadow-sm">All Status</TabsTrigger>
                      <TabsTrigger value="pending" className="px-5 data-[state=active]:bg-background data-[state=active]:shadow-sm">Pending</TabsTrigger>
                      <TabsTrigger value="approved" className="px-5 data-[state=active]:bg-background data-[state=active]:shadow-sm text-status-approved">Approved</TabsTrigger>
                      <TabsTrigger value="rejected" className="px-5 data-[state=active]:bg-background data-[state=active]:shadow-sm text-status-rejected">Rejected</TabsTrigger>
                      <TabsTrigger value="cancelled" className="px-5 data-[state=active]:bg-background data-[state=active]:shadow-sm">Cancelled</TabsTrigger>
                    </TabsList>
                  </Tabs>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Requests Table */}
          <Card className="overflow-hidden border-border/50 shadow-card bg-card/50 backdrop-blur-sm">
            <CardHeader className="border-b bg-muted/20 pb-4">
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle className="text-xl font-bold tracking-tight">Request History</CardTitle>
                  <CardDescription className="text-sm">
                    Showing {filteredRequests.length} {activeTab === 'all' ? 'total requests' : `${activeTab}s`}
                    {statusFilter !== 'all' && ` marked as ${statusFilter}`}
                  </CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent className="p-0">
              <Table>
                <TableHeader className="bg-muted/40">
                  <TableRow className="hover:bg-transparent">
                    <TableHead className="w-[140px] font-bold text-foreground/70 uppercase text-[10px] tracking-wider">Type</TableHead>
                    <TableHead className="font-bold text-foreground/70 uppercase text-[10px] tracking-wider">Summary & Details</TableHead>
                    <TableHead className="w-[180px] font-bold text-foreground/70 uppercase text-[10px] tracking-wider">Submission Date</TableHead>
                    <TableHead className="w-[120px] font-bold text-foreground/70 uppercase text-[10px] tracking-wider">Status</TableHead>
                    <TableHead className="text-right w-[100px] font-bold text-foreground/70 uppercase text-[10px] tracking-wider px-6">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {paginatedRequests.map((request) => (
                    <TableRow key={`${request.type}-${request.id}`} className="hover:bg-muted/30 transition-colors group">
                      <TableCell className="py-4">
                        <RequestTypeBadge type={request.type} />
                      </TableCell>
                      <TableCell className="py-4">
                        <div className="flex flex-col gap-1">
                          <span className="font-bold text-sm group-hover:text-primary transition-colors">
                            {getRequestDetails(request)}
                          </span>
                          <span className="text-xs text-muted-foreground/80 line-clamp-1">
                            {request.reason && request.reason !== 'No reason found'
                              ? request.reason
                              : `Note: ${request.type} description not provided by system`}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell className="py-4">
                        <div className="flex flex-col gap-0.5">
                          <span className="font-semibold text-sm">
                            {request.createdAt.startsWith('1970')
                              ? 'Date Unavailable'
                              : format(new Date(request.createdAt), 'MMM d, yyyy')}
                          </span>
                          <span className="text-[10px] text-muted-foreground/60 font-medium">
                            {request.createdAt.startsWith('1970') ? 'System auto-generated' : format(new Date(request.createdAt), 'hh:mm aa')}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell className="py-4">
                        <StatusBadge status={request.status} />
                      </TableCell>
                      <TableCell className="py-4 text-right px-6">
                        <div className="flex justify-end gap-2">
                          <Button
                            variant="secondary"
                            size="icon"
                            className="h-8 w-8 rounded-full shadow-sm hover:translate-y-[-1px] transition-all"
                            onClick={() => setSelectedRequest(request)}
                          >
                            <Eye className="h-4 w-4" />
                          </Button>
                          {request.status === 'pending' && (
                            <Button
                              variant="destructive"
                              size="icon"
                              className="h-8 w-8 rounded-full shadow-sm hover:translate-y-[-1px] transition-all"
                              onClick={() => setConfirmCancel({ type: request.type, id: request.id })}
                            >
                              <X className="h-4 w-4" />
                            </Button>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                  {paginatedRequests.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={5} className="text-center py-24">
                        <div className="flex flex-col items-center gap-4 text-muted-foreground animate-in fade-in zoom-in duration-300">
                          <div className="p-5 rounded-2xl bg-muted/20 border border-border/10">
                            <Info className="h-10 w-10 opacity-30" />
                          </div>
                          <div className="space-y-1">
                            <p className="font-bold text-xl text-foreground/80">No matches found</p>
                            <p className="text-sm max-w-[280px] mx-auto opacity-60">
                              Adjust your category or status filters to locate specific requests.
                            </p>
                          </div>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => { setActiveTab('all'); setStatusFilter('all'); }}
                            className="mt-2 rounded-xl shadow-sm hover:bg-background"
                          >
                            Reset all filters
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>

              {/* Pagination Controls - Moved inside CardContent bottom for clean UI */}
              {totalPages > 1 && (
                <div className="flex items-center justify-between px-6 py-6 border-t bg-muted/10">
                  <div className="text-xs font-medium text-muted-foreground tracking-tight">
                    Showing <span className="text-foreground">{(currentPage - 1) * ITEMS_PER_PAGE + 1}</span> to <span className="text-foreground">{Math.min(currentPage * ITEMS_PER_PAGE, filteredRequests.length)}</span> of <span className="text-foreground">{filteredRequests.length}</span> entries
                  </div>
                  <div className="flex items-center gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                      disabled={currentPage === 1}
                      className="h-8 px-4 rounded-lg shadow-sm"
                    >
                      Previous
                    </Button>
                    <div className="flex items-center gap-1.5 px-3">
                      {Array.from({ length: totalPages }, (_, i) => i + 1).map(page => (
                        <button
                          key={page}
                          onClick={() => setCurrentPage(page)}
                          className={cn(
                            "w-7 h-7 rounded-md text-[11px] font-bold transition-all",
                            currentPage === page
                              ? "bg-primary text-primary-foreground shadow-md"
                              : "hover:bg-muted text-muted-foreground"
                          )}
                        >
                          {page}
                        </button>
                      ))}
                    </div>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                      disabled={currentPage === totalPages}
                      className="h-8 px-4 rounded-lg shadow-sm"
                    >
                      Next
                    </Button>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Request Detail Dialog */}
        <Dialog open={!!selectedRequest} onOpenChange={() => setSelectedRequest(null)}>
          <DialogContent className="sm:max-w-md rounded-2xl">
            <DialogHeader>
              <DialogTitle className="flex items-center gap-3 text-2xl font-bold">
                <div className="p-2 rounded-xl bg-primary/10 text-primary">
                  {selectedRequest && getRequestIcon(selectedRequest.type)}
                </div>
                Request Overview
              </DialogTitle>
              <DialogDescription className="text-sm font-medium">
                Submission ID: #{selectedRequest?.id} • {selectedRequest && format(new Date(selectedRequest.createdAt), 'MMMM d, yyyy')}
              </DialogDescription>
            </DialogHeader>
            {selectedRequest && (
              <div className="space-y-6 pt-4">
                <div className="flex items-center justify-between p-4 rounded-2xl bg-muted/30 border border-border/50">
                  <div className="space-y-1">
                    <p className="text-[10px] font-bold uppercase tracking-widest text-muted-foreground">Category</p>
                    <RequestTypeBadge type={selectedRequest.type} />
                  </div>
                  <div className="text-right space-y-1">
                    <p className="text-[10px] font-bold uppercase tracking-widest text-muted-foreground">Current Status</p>
                    <StatusBadge status={selectedRequest.status} />
                  </div>
                </div>

                <div className="grid gap-6">
                  <div className="space-y-1">
                    <p className="text-[11px] font-bold text-muted-foreground uppercase tracking-wider">Reference Info</p>
                    <p className="text-lg font-bold leading-tight">{getRequestDetails(selectedRequest)}</p>
                  </div>

                  <div className="space-y-1">
                    <p className="text-[11px] font-bold text-muted-foreground uppercase tracking-wider">User Remarks</p>
                    <p className="text-sm text-foreground/80 italic leading-relaxed bg-muted/20 p-4 rounded-xl border border-dashed border-border/50">
                      {selectedRequest.reason && selectedRequest.reason !== 'No reason found' ? selectedRequest.reason : 'No additional context provided for this request.'}
                    </p>
                  </div>

                  {(selectedRequest.status === 'approved' || selectedRequest.status === 'rejected') && (
                    <div className="p-5 rounded-2xl bg-primary/5 border border-primary/10">
                      <div className="flex items-start gap-3">
                        <Info className="h-5 w-5 text-primary mt-0.5" />
                        <div className="space-y-1">
                          <p className="text-xs font-bold text-primary uppercase tracking-tight">Decision Note</p>
                          <p className="text-sm text-foreground/80 leading-relaxed font-medium">
                            {selectedRequest.statusReason || 'no comment found'}
                          </p>
                        </div>
                      </div>
                    </div>
                  )}
                </div>

                {selectedRequest.status === 'pending' && (
                  <Button
                    variant="destructive"
                    className="w-full h-11 rounded-xl font-bold shadow-lg shadow-destructive/20 hover:scale-[1.01] active:scale-[0.99] transition-all"
                    onClick={() => {
                      handleCancel(selectedRequest.type, selectedRequest.id);
                      setSelectedRequest(null);
                    }}
                  >
                    Cancel This Request
                  </Button>
                )}
              </div>
            )}
          </DialogContent>
        </Dialog>
        {/* Cancellation Confirmation Dialog */}
        <AlertDialog open={!!confirmCancel} onOpenChange={() => setConfirmCancel(null)}>
          <AlertDialogContent className="rounded-2xl">
            <AlertDialogHeader>
              <AlertDialogTitle className="text-2xl font-bold">Request Deletion</AlertDialogTitle>
              <AlertDialogDescription className="text-base">
                Are you sure you want to cancel this {confirmCancel?.type} request?
                This action cannot be undone and the request will be permanently removed from active processing.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter className="gap-2 pt-4">
              <AlertDialogCancel className="rounded-xl h-11 px-6 font-bold">Nevermind</AlertDialogCancel>
              <AlertDialogAction
                className="bg-destructive text-destructive-foreground hover:bg-destructive/90 rounded-xl h-11 px-6 font-bold shadow-lg shadow-destructive/20"
                onClick={() => {
                  if (confirmCancel) {
                    handleCancel(confirmCancel.type, confirmCancel.id);
                    setConfirmCancel(null);
                  }
                }}
              >
                Yes, Cancel Request
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>
    </AppLayout>
  );
}
