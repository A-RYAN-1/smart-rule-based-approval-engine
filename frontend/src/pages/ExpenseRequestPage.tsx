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
  const { balances: unifiedBalances, isLoading: isLoadingBalances } = useBalances();
  const { toast } = useToast();

  const [amount, setAmount] = useState('');
  const [category, setCategory] = useState('');
  const [reason, setReason] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const remainingExpense = unifiedBalances?.expenses.remaining;
  const balancesError = !isLoadingBalances && remainingExpense === undefined;

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

    // Check if balances failed to load
    if (remainingExpense === undefined) {
      toast({
        variant: "destructive",
        title: "Unable to load balances",
        description: "Please refresh the page and try again.",
      });
      return;
    }

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

        {/* Your Expense Balance */}
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
                Unable to load your expense balance. Please refresh the page.
              </p>
            </CardContent>
          </Card>
        ) : (
          <Card className="border-expense/20">
            <CardContent className="p-4">
              <div className="flex justify-between items-center">
                <div>
                  <p className="text-sm text-muted-foreground">Your Remaining Expense Limit</p>
                  <p className="text-2xl font-bold text-expense">${remainingExpense?.toLocaleString()}</p>
                </div>
                <div className="text-right">
                  <p className="text-sm text-muted-foreground">Total Allocation</p>
                  <p className="text-lg font-semibold">${unifiedBalances?.expenses.total?.toLocaleString()}</p>
                </div>
              </div>
            </CardContent>
          </Card>
        )}        

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
      </div>
    </AppLayout>
  );
}
