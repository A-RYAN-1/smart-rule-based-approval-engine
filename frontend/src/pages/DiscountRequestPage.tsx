import { useState } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { useAuth } from '@/contexts/AuthContext';
import { useDiscounts } from '@/hooks/useDiscounts';
import { useBalances } from '@/hooks/useBalances';
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
  const { toast } = useToast();

  const [percentage, setPercentage] = useState([5]);
  const [reason, setReason] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const getApprovalStatus = (pct: number) => {
    if (pct <= 10) return { text: 'Auto-approved', color: 'text-status-approved' };
    if (pct <= 20) return { text: 'Manager approval', color: 'text-status-pending' };
    return { text: 'Finance approval', color: 'text-status-rejected' };
  };

  const approvalStatus = getApprovalStatus(percentage[0]);

  const { balances: unifiedBalances } = useBalances();
  const remainingDiscount = unifiedBalances?.discounts.remaining ?? 100;

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

        {/* Discount Tiers */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base flex items-center gap-2">
              <Percent className="h-4 w-4 text-discount" />
              Discount Tiers
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-3 gap-4">
              <div className="p-3 rounded-lg bg-status-approved-bg text-center">
                <p className="text-xs font-medium text-muted-foreground">Standard</p>
                <p className="text-lg font-bold text-status-approved">≤ 10%</p>
                <p className="text-xs text-muted-foreground">Auto-approved</p>
              </div>
              <div className="p-3 rounded-lg bg-status-pending-bg text-center">
                <p className="text-xs font-medium text-muted-foreground">Special</p>
                <p className="text-lg font-bold text-status-pending">11-20%</p>
                <p className="text-xs text-muted-foreground">Manager review</p>
              </div>
              <div className="p-3 rounded-lg bg-status-rejected-bg text-center">
                <p className="text-xs font-medium text-muted-foreground">Premium</p>
                <p className="text-lg font-bold text-status-rejected">21-25%</p>
                <p className="text-xs text-muted-foreground">Finance review</p>
              </div>
            </div>
          </CardContent>
        </Card>

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
                  <span>10%</span>
                  <span>20%</span>
                  <span>25%</span>
                </div>
                {percentage[0] >= 21 && (
                  <p className="text-sm text-amber-600 flex items-center gap-1">
                    <Info className="h-3 w-3" />
                    Requests above 20% require Finance review and may take longer.
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

        {/* Info Card */}
        <Card className="bg-discount-bg border-discount/20">
          <CardContent className="p-4">
            <div className="flex gap-3">
              <Info className="h-5 w-5 text-discount flex-shrink-0 mt-0.5" />
              <div className="space-y-1">
                <p className="text-sm font-medium">Discount Policy</p>
                <p className="text-sm text-muted-foreground">
                  Employee discounts are available for company products and services.
                  Standard discounts up to 10% are automatically approved. Higher discounts
                  require valid justification and management approval.
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}
