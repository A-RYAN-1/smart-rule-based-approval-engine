import { useState } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { useAuth } from '@/contexts/AuthContext';
import { useExpenses } from '@/hooks/useExpenses';
import { useBalances } from '@/hooks/useBalances';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useToast } from '@/hooks/use-toast';
import { Receipt, Info, DollarSign } from 'lucide-react';
import { expenseRequestSchema } from '@/lib/validations';

const expenseCategories = [
  'Travel',
  'Office Supplies',
  'Equipment',
  'Meals & Entertainment',
  'Training',
  'Software & Subscriptions',
  'Other',
];

export default function ExpenseRequestPage() {
  const { user } = useAuth();
  const { requestExpense } = useExpenses();
  const { balances: unifiedBalances } = useBalances();
  const { toast } = useToast();

  const [amount, setAmount] = useState('');
  const [category, setCategory] = useState('');
  const [reason, setReason] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const result = expenseRequestSchema.safeParse({
      amount: parseFloat(amount),
      category,
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

    const { amount: sanitizedAmount, reason: sanitizedReason } = result.data;

    const remainingExpense = unifiedBalances?.expenses.remaining ?? 0;
    if (sanitizedAmount > remainingExpense) {
      toast({
        variant: "destructive",
        title: "Insufficient balance",
        description: `You only have $${remainingExpense.toLocaleString()} remaining in your expense limit.`,
      });
      return;
    }

    setIsSubmitting(true);

    try {
      await requestExpense({
        amount: sanitizedAmount,
        category,
        reason: sanitizedReason,
      });

      // Reset form
      setAmount('');
      setCategory('');
      setReason('');
    } catch (e) {
      console.error(e);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <AppLayout>
      <div className="max-w-2xl mx-auto space-y-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Claim Expense</h1>
          <p className="text-muted-foreground mt-1">Submit an expense reimbursement request</p>
        </div>

        {/* Expense Limits Info */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base flex items-center gap-2">
              <Receipt className="h-4 w-4 text-expense" />
              Expense Guidelines
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-3 gap-4">
              <div className="p-3 rounded-lg bg-status-approved-bg text-center">
                <p className="text-xs font-medium text-muted-foreground">Auto-Approved</p>
                <p className="text-lg font-bold text-status-approved">&lt; $1,000</p>
              </div>
              <div className="p-3 rounded-lg bg-status-pending-bg text-center">
                <p className="text-xs font-medium text-muted-foreground">Manager Review</p>
                <p className="text-lg font-bold text-status-pending">$1K - $5K</p>
              </div>
              <div className="p-3 rounded-lg bg-status-rejected-bg text-center">
                <p className="text-xs font-medium text-muted-foreground">Finance Review</p>
                <p className="text-lg font-bold text-status-rejected">&gt; $5,000</p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Expense Request Form */}
        <Card>
          <CardHeader>
            <CardTitle>Expense Details</CardTitle>
            <CardDescription>Provide details about your expense claim</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-6">
              <div className="space-y-2">
                <Label htmlFor="amount">Amount ($)</Label>
                <div className="relative">
                  <DollarSign className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input
                    id="amount"
                    type="number"
                    placeholder="0.00"
                    value={amount}
                    onChange={(e) => setAmount(e.target.value)}
                    className="pl-10"
                    min="0"
                    step="0.01"
                    required
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="category">Category</Label>
                <Select value={category} onValueChange={setCategory} required>
                  <SelectTrigger>
                    <SelectValue placeholder="Select expense category" />
                  </SelectTrigger>
                  <SelectContent className="bg-popover">
                    {expenseCategories.map((cat) => (
                      <SelectItem key={cat} value={cat}>{cat}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="reason">Description</Label>
                <Textarea
                  id="reason"
                  placeholder="Describe the expense and its business purpose..."
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                  rows={4}
                  required
                />
              </div>

              <div className="flex gap-3">
                <Button type="submit" className="flex-1" disabled={isSubmitting}>
                  {isSubmitting ? 'Submitting...' : 'Submit Claim'}
                </Button>
                <Button type="button" variant="outline" onClick={() => {
                  setAmount('');
                  setCategory('');
                  setReason('');
                }}>
                  Clear
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>

        {/* Info Card */}
        <Card className="bg-expense-bg border-expense/20">
          <CardContent className="p-4">
            <div className="flex gap-3">
              <Info className="h-5 w-5 text-expense flex-shrink-0 mt-0.5" />
              <div className="space-y-1">
                <p className="text-sm font-medium">Receipt Requirements</p>
                <p className="text-sm text-muted-foreground">
                  Please keep all receipts for expenses over $50. Receipts may be requested
                  during the review process. Expenses without proper documentation may be rejected.
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}
