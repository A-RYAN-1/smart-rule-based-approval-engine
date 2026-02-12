import { useState, useMemo } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { useAuth } from '@/contexts/AuthContext';
import { useLeaves } from '@/hooks/useLeaves';
import { useExpenses } from '@/hooks/useExpenses';
import { useDiscounts } from '@/hooks/useDiscounts';
import { StatusBadge, RequestTypeBadge } from '@/components/ui/status-badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogClose
} from '@/components/ui/dialog';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { useToast } from '@/hooks/use-toast';
import { format } from 'date-fns';
import { CheckCircle2, XCircle, Eye, Clock, Calendar, DollarSign, Percent, User } from 'lucide-react';
import { RequestType } from '@/types';
import { Navigate } from 'react-router-dom';

export default function PendingApprovalsPage() {
  const { user } = useAuth();
  const { pendingLeaves, approveLeave, rejectLeave } = useLeaves();
  const { pendingExpenses, approveExpense, rejectExpense } = useExpenses();
  const { pendingDiscounts, approveDiscount, rejectDiscount } = useDiscounts();

  const { toast } = useToast();
  const [selectedRequest, setSelectedRequest] = useState<any>(null);
  const [actionType, setActionType] = useState<'approve' | 'reject' | null>(null);
  const [reason, setReason] = useState('');
  const [activeTab, setActiveTab] = useState<'all' | RequestType>('all');

  // Only managers and admins can access
  if (user?.role === 'employee') {
    return <Navigate to="/dashboard" replace />;
  }

  const pendingRequests = useMemo(() => [
    ...pendingLeaves.map(r => ({ ...r, type: 'leave' as const })),
    ...pendingExpenses.map(r => ({ ...r, type: 'expense' as const })),
    ...pendingDiscounts.map(r => ({ ...r, type: 'discount' as const })),
  ].sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()), [pendingLeaves, pendingExpenses, pendingDiscounts]);

  const filteredRequests = activeTab === 'all'
    ? pendingRequests
    : pendingRequests.filter(r => r.type === activeTab);

  const handleAction = async () => {
    if (!selectedRequest || !actionType) return;

    // Note: The reason logic is "optional" in UI but maybe API ignores it if not provided or supported?
    // Postman doesn't explicitely show reason passing for Approve/Reject endpoints in the snippet provided.
    // "Approve Leave" -> POST /api/leaves/1/approve. Body is EMPTY?
    // "Reject Leave" -> POST /api/leaves/1/reject. Body is EMPTY?
    // User instruction says: "The manager can approve or reject... action updates request state..."
    // It doesn't explicitly mention providing a reason *to the backend*.
    // However, industry standard usually allows a reason, especially for rejection.
    // Given the Postman snippet only showed URL, I will assume NO BODY for now, or ignore the reason.
    // If API supports it, great. If not, we just call the endpoint.
    // But `useLeaves` hook calls `approveLeave(id)`. I didn't add `reason` argument to `approveLeave` in `leave.service.ts`.
    // So I will just call the function.

    try {
      if (selectedRequest.type === 'leave') {
        if (actionType === 'approve') await approveLeave({ id: selectedRequest.id, comment: reason });
        else await rejectLeave({ id: selectedRequest.id, comment: reason });
      } else if (selectedRequest.type === 'expense') {
        if (actionType === 'approve') await approveExpense({ id: selectedRequest.id, comment: reason });
        else await rejectExpense({ id: selectedRequest.id, comment: reason });
      } else if (selectedRequest.type === 'discount') {
        if (actionType === 'approve') await approveDiscount({ id: selectedRequest.id, comment: reason });
        else await rejectDiscount({ id: selectedRequest.id, comment: reason });
      }
      // Hook handles toast.
      setSelectedRequest(null);
      setActionType(null);
      setReason('');
    } catch (e) {
      console.error(e);
      // User notification is handled in hook usually, or we can add extra here.
    }
  };

  const getRequestDetails = (request: any) => {
    if (request.type === 'leave') {
      return `${request.leaveType} • ${format(new Date(request.fromDate), 'MMM d')} - ${format(new Date(request.toDate), 'MMM d')}`;
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

  return (
    <AppLayout>
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Pending Approvals</h1>
          <p className="text-muted-foreground mt-1">Review and process pending requests</p>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <Card className="bg-status-pending-bg border-status-pending/20">
            <CardContent className="p-4 text-center">
              <Clock className="h-6 w-6 mx-auto mb-2 text-status-pending" />
              <p className="text-2xl font-bold">{pendingRequests.length}</p>
              <p className="text-sm text-muted-foreground">Pending</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4 text-center">
              <Calendar className="h-6 w-6 mx-auto mb-2 text-leave" />
              <p className="text-2xl font-bold">
                {pendingRequests.filter(r => r.type === 'leave').length}
              </p>
              <p className="text-sm text-muted-foreground">Leave</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4 text-center">
              <DollarSign className="h-6 w-6 mx-auto mb-2 text-expense" />
              <p className="text-2xl font-bold">
                {pendingRequests.filter(r => r.type === 'expense').length}
              </p>
              <p className="text-sm text-muted-foreground">Expense</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4 text-center">
              <Percent className="h-6 w-6 mx-auto mb-2 text-discount" />
              <p className="text-2xl font-bold">
                {pendingRequests.filter(r => r.type === 'discount').length}
              </p>
              <p className="text-sm text-muted-foreground">Discount</p>
            </CardContent>
          </Card>
        </div>

        {/* Requests Table */}
        <Card>
          <CardHeader>
            <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
              <div>
                <CardTitle>Requests Awaiting Review</CardTitle>
                <CardDescription>Click on a request to view details and take action</CardDescription>
              </div>
              <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as any)}>
                <TabsList>
                  <TabsTrigger value="all">All</TabsTrigger>
                  <TabsTrigger value="leave">Leave</TabsTrigger>
                  <TabsTrigger value="expense">Expense</TabsTrigger>
                  <TabsTrigger value="discount">Discount</TabsTrigger>
                </TabsList>
              </Tabs>
            </div>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Type</TableHead>
                  <TableHead>Requester</TableHead>
                  <TableHead>Details</TableHead>
                  <TableHead>Reason</TableHead>
                  <TableHead>Date</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredRequests.map((request) => (
                  <TableRow key={`${request.type}-${request.id}`}>
                    <TableCell>
                      <RequestTypeBadge type={request.type} />
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <div className="w-8 h-8 rounded-full bg-muted flex items-center justify-center">
                          <User className="h-4 w-4 text-muted-foreground" />
                        </div>
                        <span className="font-medium">{request.userName}</span>
                      </div>
                    </TableCell>
                    <TableCell className="font-medium">
                      {getRequestDetails(request)}
                    </TableCell>
                    <TableCell className="max-w-[200px] truncate">
                      {request.reason}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {format(new Date(request.createdAt), 'MMM d, yyyy')}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => setSelectedRequest(request)}
                        >
                          <Eye className="h-4 w-4" />
                        </Button>
                        <Button
                          size="icon"
                          className="bg-status-approved hover:bg-status-approved/90"
                          onClick={() => {
                            setSelectedRequest(request);
                            setActionType('approve');
                          }}
                        >
                          <CheckCircle2 className="h-4 w-4" />
                        </Button>
                        <Button
                          size="icon"
                          variant="destructive"
                          onClick={() => {
                            setSelectedRequest(request);
                            setActionType('reject');
                          }}
                        >
                          <XCircle className="h-4 w-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
                {filteredRequests.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center py-12">
                      <Clock className="h-12 w-12 mx-auto mb-4 text-muted-foreground/50" />
                      <p className="text-muted-foreground">No pending requests</p>
                      <p className="text-sm text-muted-foreground">All caught up! 🎉</p>
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        {/* Action Dialog */}
        <Dialog open={!!actionType} onOpenChange={() => {
          setActionType(null);
          setReason('');
        }}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2">
                {actionType === 'approve' ? (
                  <CheckCircle2 className="h-5 w-5 text-status-approved" />
                ) : (
                  <XCircle className="h-5 w-5 text-destructive" />
                )}
                {actionType === 'approve' ? 'Approve Request' : 'Reject Request'}
              </DialogTitle>
              <DialogDescription>
                {selectedRequest && (
                  <>
                    {selectedRequest.type.charAt(0).toUpperCase() + selectedRequest.type.slice(1)} request from {selectedRequest.userName}
                  </>
                )}
              </DialogDescription>
            </DialogHeader>

            {selectedRequest && (
              <div className="space-y-4">
                <div className="p-3 rounded-lg bg-muted">
                  <p className="text-sm font-medium">{getRequestDetails(selectedRequest)}</p>
                  <p className="text-sm text-muted-foreground mt-1">{selectedRequest.reason}</p>
                </div>

                <div className="space-y-2">
                  <Label>Comment (optional)</Label>
                  <Textarea
                    placeholder={`Add a reason for ${actionType === 'approve' ? 'approving' : 'rejecting'} this request...`}
                    value={reason}
                    onChange={(e) => setReason(e.target.value)}
                  />
                </div>
              </div>
            )}

            <DialogFooter>
              <Button variant="outline" onClick={() => {
                setActionType(null);
                setReason('');
              }}>
                Cancel
              </Button>
              <Button
                onClick={handleAction}
                variant={actionType === 'approve' ? 'default' : 'destructive'}
              >
                {actionType === 'approve' ? 'Approve' : 'Reject'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* View Dialog */}
        <Dialog open={!!selectedRequest && !actionType} onOpenChange={() => setSelectedRequest(null)}>
          <DialogContent className="sm:max-w-md">
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2">
                {selectedRequest && getRequestIcon(selectedRequest.type)}
                Request Details
              </DialogTitle>
              <DialogDescription>
                Submitted on {selectedRequest && format(new Date(selectedRequest.createdAt), 'MMMM d, yyyy')}
              </DialogDescription>
            </DialogHeader>
            {selectedRequest && (
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <RequestTypeBadge type={selectedRequest.type} />
                  <StatusBadge status={selectedRequest.status} />
                </div>

                <div className="space-y-3">
                  <div className="flex items-center gap-2">
                    <User className="h-4 w-4 text-muted-foreground" />
                    <span className="font-medium">{selectedRequest.userName}</span>
                  </div>

                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Details</p>
                    <p className="font-medium">{getRequestDetails(selectedRequest)}</p>
                  </div>

                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Reason</p>
                    <p>{selectedRequest.reason}</p>
                  </div>
                </div>

                <div className="flex gap-2">
                  <Button
                    className="flex-1 bg-status-approved hover:bg-status-approved/90"
                    onClick={() => setActionType('approve')}
                  >
                    <CheckCircle2 className="h-4 w-4 mr-2" />
                    Approve
                  </Button>
                  <Button
                    className="flex-1"
                    variant="destructive"
                    onClick={() => setActionType('reject')}
                  >
                    <XCircle className="h-4 w-4 mr-2" />
                    Reject
                  </Button>
                </div>
              </div>
            )}
          </DialogContent>
        </Dialog>
      </div>
    </AppLayout>
  );
}
