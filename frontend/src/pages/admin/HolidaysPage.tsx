import { useState } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { useAuth } from '@/contexts/AuthContext';
import { useAdmin } from '@/hooks/useAdmin';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Calendar } from '@/components/ui/calendar';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
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
} from '@/components/ui/dialog';
import { useToast } from '@/hooks/use-toast';
import { format } from 'date-fns';
import { CalendarIcon, Plus, Trash2, Calendar as CalendarIcon2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Navigate } from 'react-router-dom';

export default function HolidaysPage() {
  const { user } = useAuth();
  const { holidays, addHoliday, deleteHoliday, isLoadingHolidays } = useAdmin();
  const { toast } = useToast();
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [newDate, setNewDate] = useState<Date>();
  const [newDescription, setNewDescription] = useState('');

  // Only admins can access
  if (user?.role !== 'admin') {
    return <Navigate to="/dashboard" replace />;
  }

  if (isLoadingHolidays) {
    return (
      <AppLayout>
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
        </div>
      </AppLayout>
    );
  }

  const handleAddHoliday = async () => {
    if (!newDate || !newDescription.trim()) {
      toast({
        title: 'Invalid input',
        description: 'Please provide both date and description.',
        variant: 'destructive',
      });
      return;
    }

    try {
      await addHoliday({
        date: format(newDate, 'yyyy-MM-dd'),
        description: newDescription.trim(),
      });

      setIsAddDialogOpen(false);
      setNewDate(undefined);
      setNewDescription('');
    } catch (e) {
      console.error(e);
    }
  };

  const handleDeleteHoliday = async (id: number, description: string) => {
    try {
      await deleteHoliday(id);
      // Toast handled by hook
    } catch (e) {
      console.error(e);
    }
  };

  const sortedHolidays = [...holidays].sort(
    (a, b) => new Date(a.date).getTime() - new Date(b.date).getTime()
  );

  const upcomingHolidays = sortedHolidays.filter(
    h => new Date(h.date) >= new Date()
  );

  return (
    <AppLayout>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Holiday Management</h1>
            <p className="text-muted-foreground mt-1">Manage company holidays and non-working days</p>
          </div>
          <Button onClick={() => setIsAddDialogOpen(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Add Holiday
          </Button>
        </div>

        <div className="grid gap-6 lg:grid-cols-3">
          {/* Holiday Stats */}
          <Card>
            <CardContent className="p-6">
              <div className="flex items-center gap-4">
                <div className="p-3 rounded-xl bg-primary/10">
                  <CalendarIcon2 className="h-6 w-6 text-primary" />
                </div>
                <div>
                  <p className="text-2xl font-bold">{holidays.length}</p>
                  <p className="text-sm text-muted-foreground">Total Holidays</p>
                </div>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              <div className="flex items-center gap-4">
                <div className="p-3 rounded-xl bg-status-approved-bg">
                  <CalendarIcon2 className="h-6 w-6 text-status-approved" />
                </div>
                <div>
                  <p className="text-2xl font-bold">{upcomingHolidays.length}</p>
                  <p className="text-sm text-muted-foreground">Upcoming</p>
                </div>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              <div className="flex items-center gap-4">
                <div className="p-3 rounded-xl bg-muted">
                  <CalendarIcon2 className="h-6 w-6 text-muted-foreground" />
                </div>
                <div>
                  <p className="text-2xl font-bold">{holidays.length - upcomingHolidays.length}</p>
                  <p className="text-sm text-muted-foreground">Past</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Holidays Table */}
        <Card>
          <CardHeader>
            <CardTitle>All Holidays</CardTitle>
            <CardDescription>Holidays affect leave calculations and approvals</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Date</TableHead>
                  <TableHead>Day</TableHead>
                  <TableHead>Description</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {sortedHolidays.map((holiday) => {
                  const holidayDate = new Date(holiday.date);
                  const isPast = holidayDate < new Date();

                  return (
                    <TableRow key={holiday.id} className={isPast ? 'opacity-60' : ''}>
                      <TableCell className="font-medium">
                        {format(holidayDate, 'MMMM d, yyyy')}
                      </TableCell>
                      <TableCell>
                        {format(holidayDate, 'EEEE')}
                      </TableCell>
                      <TableCell>{holiday.description}</TableCell>
                      <TableCell>
                        <span className={cn(
                          'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                          isPast ? 'bg-muted text-muted-foreground' : 'bg-status-approved-bg text-status-approved'
                        )}>
                          {isPast ? 'Past' : 'Upcoming'}
                        </span>
                      </TableCell>
                      <TableCell className="text-right">
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleDeleteHoliday(holiday.id, holiday.description)}
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  );
                })}
                {holidays.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={5} className="text-center py-8 text-muted-foreground">
                      No holidays configured. Add some to get started!
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        {/* Add Holiday Dialog */}
        <Dialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Add New Holiday</DialogTitle>
              <DialogDescription>
                Add a company holiday or non-working day
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4">
              <div className="space-y-2">
                <Label>Date</Label>
                <Popover>
                  <PopoverTrigger asChild>
                    <Button
                      variant="outline"
                      className={cn(
                        'w-full justify-start text-left font-normal',
                        !newDate && 'text-muted-foreground'
                      )}
                    >
                      <CalendarIcon className="mr-2 h-4 w-4" />
                      {newDate ? format(newDate, 'PPP') : 'Pick a date'}
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent className="w-auto p-0 bg-popover" align="start">
                    <Calendar
                      mode="single"
                      selected={newDate}
                      onSelect={setNewDate}
                      initialFocus
                      className="pointer-events-auto"
                    />
                  </PopoverContent>
                </Popover>
              </div>

              <div className="space-y-2">
                <Label>Description</Label>
                <Input
                  placeholder="e.g., Christmas Day"
                  value={newDescription}
                  onChange={(e) => setNewDescription(e.target.value)}
                />
              </div>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setIsAddDialogOpen(false)}>
                Cancel
              </Button>
              <Button onClick={handleAddHoliday}>
                Add Holiday
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </AppLayout>
  );
}
