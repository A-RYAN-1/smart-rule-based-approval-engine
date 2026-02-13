import { useState } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { useAuth } from '@/contexts/AuthContext';
import { useLeaves } from '@/hooks/useLeaves';
import { useBalances } from '@/hooks/useBalances';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Calendar } from '@/components/ui/calendar';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { useToast } from '@/hooks/use-toast';
import { CalendarIcon, CalendarDays, Info } from 'lucide-react';
import { format, differenceInDays } from 'date-fns';
import { cn } from '@/lib/utils';
import { LeaveType } from '@/types';
import { leaveRequestSchema } from '@/lib/validations';

export default function LeaveRequestPage() {
  const { user } = useAuth();
  const { requestLeave } = useLeaves();
  const { balances: unifiedBalances } = useBalances();
  const balances = unifiedBalances?.leaves || [];
  const { toast } = useToast();

  const [fromDate, setFromDate] = useState<Date>();
  const [toDate, setToDate] = useState<Date>();
  const [leaveType, setLeaveType] = useState<LeaveType>('EARN');
  const [reason, setReason] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const selectedBalance = balances.find(b => b.leaveType === leaveType);
  const leaveDays = fromDate && toDate ? differenceInDays(toDate, fromDate) + 1 : 0;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const result = leaveRequestSchema.safeParse({
      fromDate: fromDate ? format(fromDate, 'yyyy-MM-dd') : '',
      toDate: toDate ? format(toDate, 'yyyy-MM-dd') : '',
      leaveType,
      reason,
    });

    if (!result.success) {
      toast({
        title: 'Validation Error',
        description: result.error.errors[0].message,
        variant: 'destructive',
      });
      return;
    }

    const { fromDate: sanitizedFrom, toDate: sanitizedTo, reason: sanitizedReason } = result.data;

    const earnBalance = balances.find(b => b.leaveType === 'EARN')?.balance ?? 0;
    // Allow applying for leave if either the specific type has balance or EARN has balance
    const effectiveBalance = Math.max(selectedBalance?.balance ?? 0, earnBalance);

    if (leaveDays > effectiveBalance) {
      toast({
        variant: "destructive",
        title: "Insufficient balance",
        description: `You only have ${effectiveBalance} days available (including Earned Leave balance).`,
      });
      return;
    }

    setIsSubmitting(true);

    try {
      await requestLeave({
        fromDate: sanitizedFrom,
        toDate: sanitizedTo,
        leaveType,
        reason: sanitizedReason,
      });

      // Reset form
      setFromDate(undefined);
      setToDate(undefined);
      setReason('');
    } catch (error) {
      console.error(error);
      // Hook handles error toast
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <AppLayout>
      <div className="max-w-2xl mx-auto space-y-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Apply for Leave</h1>
          <p className="text-muted-foreground mt-1">Submit a new leave request</p>
        </div>

        {/* Leave Balances Summary */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base flex items-center gap-2">
              <CalendarDays className="h-4 w-4 text-primary" />
              Your Leave Balances
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              {balances.map((balance) => (
                <div
                  key={balance.leaveType}
                  className={cn(
                    'p-3 rounded-lg text-center transition-colors',
                    balance.leaveType === leaveType
                      ? 'bg-primary/10 border border-primary/20'
                      : 'bg-muted/50'
                  )}
                >
                  <p className="text-xs font-medium text-muted-foreground">{balance.leaveType}</p>
                  <p className="text-2xl font-bold">{balance.balance}</p>
                  <p className="text-xs text-muted-foreground">of {balance.total}</p>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Leave Request Form */}
        <Card>
          <CardHeader>
            <CardTitle>Leave Details</CardTitle>
            <CardDescription>Fill in the details for your leave request</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-6">
              <div className="space-y-2">
                <Label htmlFor="leaveType">Leave Type</Label>
                <Select value={leaveType} onValueChange={(v) => setLeaveType(v as LeaveType)}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select leave type" />
                  </SelectTrigger>
                  <SelectContent className="bg-popover">
                    <SelectItem value="EARN">Earned Leave</SelectItem>
                    <SelectItem value="SICK">Sick Leave</SelectItem>
                    <SelectItem value="CASUAL">Casual Leave</SelectItem>
                    <SelectItem value="UNPAID">Unpaid Leave</SelectItem>
                  </SelectContent>
                </Select>
                {selectedBalance && selectedBalance.balance === 0 && (
                  <p className="text-sm text-destructive flex items-center gap-1">
                    <Info className="h-3 w-3" />
                    No balance available for this leave type
                  </p>
                )}
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label>From Date</Label>
                  <Popover>
                    <PopoverTrigger asChild>
                      <Button
                        variant="outline"
                        className={cn(
                          'w-full justify-start text-left font-normal',
                          !fromDate && 'text-muted-foreground'
                        )}
                      >
                        <CalendarIcon className="mr-2 h-4 w-4" />
                        {fromDate ? format(fromDate, 'PPP') : 'Pick a date'}
                      </Button>
                    </PopoverTrigger>
                    <PopoverContent className="w-auto p-0 bg-popover" align="start">
                      <Calendar
                        mode="single"
                        selected={fromDate}
                        onSelect={setFromDate}
                        disabled={(date) => date < new Date()}
                        initialFocus
                        className="pointer-events-auto"
                      />
                    </PopoverContent>
                  </Popover>
                </div>

                <div className="space-y-2">
                  <Label>To Date</Label>
                  <Popover>
                    <PopoverTrigger asChild>
                      <Button
                        variant="outline"
                        className={cn(
                          'w-full justify-start text-left font-normal',
                          !toDate && 'text-muted-foreground'
                        )}
                      >
                        <CalendarIcon className="mr-2 h-4 w-4" />
                        {toDate ? format(toDate, 'PPP') : 'Pick a date'}
                      </Button>
                    </PopoverTrigger>
                    <PopoverContent className="w-auto p-0 bg-popover" align="start">
                      <Calendar
                        mode="single"
                        selected={toDate}
                        onSelect={setToDate}
                        disabled={(date) => date < (fromDate || new Date())}
                        initialFocus
                        className="pointer-events-auto"
                      />
                    </PopoverContent>
                  </Popover>
                </div>
              </div>

              {leaveDays > 0 && (
                <div className="p-3 rounded-lg bg-muted/50 text-sm">
                  <span className="font-medium">Duration:</span> {leaveDays} day{leaveDays > 1 ? 's' : ''}
                  {leaveDays > 7 && (
                    <span className="text-destructive ml-2">(May require additional approval)</span>
                  )}
                </div>
              )}

              <div className="space-y-2">
                <Label htmlFor="reason">Reason</Label>
                <Textarea
                  id="reason"
                  placeholder="Please provide a reason for your leave request..."
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                  rows={4}
                  required
                />
              </div>

              <div className="flex gap-3">
                <Button type="submit" className="flex-1" disabled={isSubmitting}>
                  {isSubmitting ? 'Submitting...' : 'Submit Request'}
                </Button>
                <Button type="button" variant="outline" onClick={() => {
                  setFromDate(undefined);
                  setToDate(undefined);
                  setReason('');
                }}>
                  Clear
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}
