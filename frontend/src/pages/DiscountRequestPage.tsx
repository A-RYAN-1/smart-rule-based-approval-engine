import { useState, useMemo } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { useAuth } from '@/contexts/AuthContext';
import { useDiscounts } from '@/hooks/useDiscounts';
import { useBalances } from '@/hooks/useBalances';
import { useAdmin } from '@/hooks/useAdmin';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Slider } from '@/components/ui/slider';
import { useToast } from '@/hooks/use-toast';
import { Percent, Info } from 'lucide-react';
import { cn } from '@/lib/utils';
import { discountRequestSchema } from '@/lib/validations';

export default function DiscountRequestPage() {
  const { user } = useAuth();
  const { requestDiscount } = useDiscounts();
  const { rules } = useAdmin();
  const { toast } = useToast();

  const [percentage, setPercentage] = useState([5]);
  const [reason, setReason] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Extract discount thresholds from active rules
  const discountThresholds = useMemo(() => {
    // Default thresholds (fallback)
    const defaults = {
      autoApproveThreshold: 10,      // Auto-approve up to this %
      managerApprovalThreshold: 20,  // Manager approval for this range
      financeApprovalThreshold: 100  // Finance approval above this
    };

    if (!rules || rules.length === 0) return defaults;

    // Look for discount rules and extract conditions
    const discountRules = rules.filter(r => r.requestType.toUpperCase() === 'DISCOUNT' && r.isActive);
    
    // Parse conditions to find thresholds
    // Rules typically have conditions like: { "discount_percentage": { "<=": 10 } }
    for (const rule of discountRules) {
      if (rule.action === 'auto_approve' && rule.condition?.discount_percentage) {
        const condition = rule.condition.discount_percentage;
        if (condition['<=']) {
          defaults.autoApproveThreshold = Math.max(defaults.autoApproveThreshold, condition['<=']);
        }
      }
    }

    // For manager approval threshold, look for rules with manager approval action
    // or use a range between auto-approve and finance
    defaults.managerApprovalThreshold = defaults.autoApproveThreshold + 10;

    return defaults;
  }, [rules]);

  const getApprovalStatus = (pct: number) => {
    if (user?.role === 'employee') {
      if (pct <= discountThresholds.autoApproveThreshold) {
        return { text: 'Auto-approved', color: 'text-status-approved' };
      }
      if (pct <= discountThresholds.managerApprovalThreshold) {
        return { text: 'Manager approval', color: 'text-status-pending' };
      }
      return { text: 'Finance approval', color: 'text-status-rejected' };
    }
    // Manager/Admin can auto-approve all discounts they create
    return { text: 'Auto-approved', color: 'text-status-approved' };
  };

  const approvalStatus = getApprovalStatus(percentage[0]);

  const { balances: unifiedBalances, isLoading: isLoadingBalances } = useBalances();
  const remainingDiscount = unifiedBalances?.discounts.remaining;

  // Show error if balances failed to load
  const balancesError = !isLoadingBalances && remainingDiscount === undefined;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const result = discountRequestSchema.safeParse({
      discountPercentage: percentage[0],
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

    const { reason: sanitizedReason } = result.data;

    // Check if balances failed to load
    if (remainingDiscount === undefined) {
      toast({
        variant: "destructive",
        title: "Unable to load balances",
        description: "Please refresh the page and try again.",
      });
      return;
    }

    if (percentage[0] > remainingDiscount) {
      toast({
        variant: "destructive",
        title: "Insufficient balance",
        description: `You only have ${remainingDiscount}% discount remaining in your quota.`,
      });
      return;
    }

    setIsSubmitting(true);

    try {
      await requestDiscount({
        discountPercentage: percentage[0],
        reason: sanitizedReason,
      });

      toast({
        title: "Request submitted",
        description: "Your discount request has been sent for approval.",
      });

      // Reset form
      setPercentage([5]);
      setReason('');
    } catch (error: any) {
      const serverMessage = error.response?.data?.message || error.response?.data?.error || "Check your input and try again.";
      console.error('DISCOUNT ERROR:', error.response?.data || error);

      toast({
        variant: "destructive",
        title: "Submission failed",
        description: serverMessage,
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <AppLayout>
      <div className="max-w-2xl mx-auto space-y-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Request Discount</h1>
          <p className="text-muted-foreground mt-1">Apply for employee discounts and benefits</p>
        </div>
        {/* Your Discount Balance */}
        {isLoadingBalances ? (
          <Card>
            <CardContent className="p-4">
              <div className="flex items-center justify-center">
                <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary" />
              </div>
            </CardContent>
          </Card>
        ) : balancesError ? (
          <Card className="border-destructive/50 bg-destructive/10">
            <CardContent className="p-4">
              <p className="text-sm text-destructive">
                <Info className="h-4 w-4 inline mr-2" />
                Unable to load your discount balance. Please refresh the page.
              </p>
            </CardContent>
          </Card>
        ) : (
          <Card className="border-discount/20">
            <CardContent className="p-4">
              <p className="text-sm text-muted-foreground">Your Remaining Discount Quota</p>
              <p className="text-2xl font-bold text-discount">{remainingDiscount}% of {unifiedBalances?.discounts.total}%</p>
            </CardContent>
          </Card>
        )}

        {/* Discount Request Form */}
        <Card>
          <CardHeader>
            <CardTitle>Discount Details</CardTitle>
            <CardDescription>Specify the discount percentage you're requesting</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-6">
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <Label>Discount Percentage</Label>
                  <div className="flex items-center gap-2">
                    <span className="text-3xl font-bold">{percentage[0]}%</span>
                    <span className={cn('text-sm font-medium', approvalStatus.color)}>
                      ({approvalStatus.text})
                    </span>
                  </div>
                </div>
                <Slider
                  value={percentage}
                  onValueChange={setPercentage}
                  max={25}
                  min={1}
                  step={1}
                  className="py-4"
                />
                <div className="flex justify-between text-xs text-muted-foreground">
                  <span>1%</span>
                  <span>{discountThresholds.autoApproveThreshold}%</span>
                  <span>{discountThresholds.managerApprovalThreshold}%</span>
                  <span>25%</span>
                </div>
                {percentage[0] > discountThresholds.managerApprovalThreshold && (
                  <p className="text-sm text-amber-600 flex items-center gap-1">
                    <Info className="h-3 w-3" />
                    Requests above {discountThresholds.managerApprovalThreshold}% require Finance review and may take longer.
                  </p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="reason">Reason / Justification</Label>
                <Textarea
                  id="reason"
                  placeholder="Explain why you're requesting this discount..."
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
                  setPercentage([5]);
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
