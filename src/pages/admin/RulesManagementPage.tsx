import { useState } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { useAuth } from '@/contexts/AuthContext';
import { useAdmin } from '@/hooks/useAdmin';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Badge } from '@/components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
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
import { toast } from 'sonner';
import { Shield, Plus, Trash2, Zap, CheckCircle2, XCircle, Users, Edit } from 'lucide-react';
import { RequestType, ApprovalRule } from '@/types';
import { Navigate } from 'react-router-dom';
import { cn } from '@/lib/utils';

export default function RulesManagementPage() {
  const { user } = useAuth();
  const {
    rules,
    addRule,
    updateRule,
    toggleRule,
    deleteRule,
    isLoadingRules,
    runAutoReject,
    isRunningAutoReject
  } = useAdmin();
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
  const [isAddMode, setIsAddMode] = useState(true);
  const [editingRuleId, setEditingRuleId] = useState<number | null>(null);
  const [newRule, setNewRule] = useState({
    requestType: 'leave' as RequestType,
    condition: {} as any,
    action: 'AUTO_APPROVE' as ApprovalRule['action'],
    priority: 1,
    gradeId: 1,
  });

  // Only admins can access
  if (user?.role !== 'admin') {
    return <Navigate to="/dashboard" replace />;
  }

  if (isLoadingRules) {
    return (
      <AppLayout>
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
        </div>
      </AppLayout>
    );
  }

  const resetForm = () => {
    setIsAddMode(true);
    setEditingRuleId(null);
    setNewRule({
      requestType: 'leave',
      condition: {},
      action: 'AUTO_APPROVE',
      priority: 1,
      gradeId: 1,
    });
  };

  const handleEditRule = (rule: ApprovalRule) => {
    setIsAddMode(false);
    setEditingRuleId(rule.id);
    setNewRule({
      requestType: rule.requestType,
      condition: rule.condition,
      action: rule.action,
      gradeId: rule.gradeId,
      priority: rule.priority,
    });
    setIsEditDialogOpen(true);
  };

  const handleSaveRule = async () => {
    try {
      if (typeof newRule.condition !== 'object' || newRule.condition === null) {
        toast.error('Condition must be valid JSON');
        return;
      }

      if (isAddMode) {
        await addRule({
          ...newRule,
          isActive: true,
        });
      } else {
        await updateRule({
          id: editingRuleId!,
          rule: newRule,
        });
      }
      setIsAddDialogOpen(false);
      setIsEditDialogOpen(false);
      resetForm();
    } catch (e) {
      console.error(e);
    }
  };

  const handleAddRule = async () => {
    try {
      await addRule({
        ...newRule,
        isActive: true,
      });

      setIsAddDialogOpen(false);
      setNewRule({
        requestType: 'leave',
        condition: {},
        action: 'AUTO_APPROVE',
        priority: 1,
        gradeId: 1,
      });
    } catch (e) {
      console.error(e);
    }
  };

  const handleDeleteRule = async (id: number) => {
    try {
      await deleteRule(id);
    } catch (e) {
      console.error(e);
    }
  };

  const getActionIcon = (action: ApprovalRule['action']) => {
    const a = action.toLowerCase();
    switch (a) {
      case 'auto_approve': return <CheckCircle2 className="h-4 w-4 text-status-approved" />;
      case 'auto_reject': return <XCircle className="h-4 w-4 text-status-rejected" />;
      case 'assign_approver': return <Users className="h-4 w-4 text-status-pending" />;
      default: return null;
    }
  };

  const getActionLabel = (action: ApprovalRule['action']) => {
    const a = action.toLowerCase();
    switch (a) {
      case 'auto_approve': return 'Auto Approve';
      case 'auto_reject': return 'Auto Reject';
      case 'assign_approver': return 'Assign Approver';
      default: return action;
    }
  };

  const getActionBadgeVariant = (action: ApprovalRule['action']) => {
    const a = action.toLowerCase();
    switch (a) {
      case 'auto_approve': return 'bg-status-approved-bg text-status-approved';
      case 'auto_reject': return 'bg-status-rejected-bg text-status-rejected';
      case 'assign_approver': return 'bg-status-pending-bg text-status-pending';
      default: return 'bg-muted';
    }
  };

  const rulesByType = {
    leave: rules.filter(r => r.requestType === 'leave'),
    expense: rules.filter(r => r.requestType === 'expense'),
    discount: rules.filter(r => r.requestType === 'discount'),
  };

  return (
    <AppLayout>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Rules Management</h1>
            <p className="text-muted-foreground mt-1">Configure automatic approval rules</p>
          </div>
          <div className="flex items-center gap-3">
            <Button onClick={() => {
              setIsAddMode(true);
              setEditingRuleId(null);
              resetForm();
              setIsAddDialogOpen(true);
            }}>
              <Plus className="h-4 w-4 mr-2" />
              Add Rule
            </Button>
          </div>
        </div>

        {/* Stats Overview */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <Card>
            <CardContent className="p-4 text-center">
              <Shield className="h-6 w-6 mx-auto mb-2 text-primary" />
              <p className="text-2xl font-bold">{rules.length}</p>
              <p className="text-sm text-muted-foreground">Total Rules</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4 text-center">
              <Zap className="h-6 w-6 mx-auto mb-2 text-status-approved" />
              <p className="text-2xl font-bold">{rules.filter(r => r.isActive).length}</p>
              <p className="text-sm text-muted-foreground">Active</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4 text-center">
              <CheckCircle2 className="h-6 w-6 mx-auto mb-2 text-status-approved" />
              <p className="text-2xl font-bold">
                {rules.filter(r => r.action === 'auto_approve').length}
              </p>
              <p className="text-sm text-muted-foreground">Auto-Approve</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4 text-center">
              <XCircle className="h-6 w-6 mx-auto mb-2 text-status-rejected" />
              <p className="text-2xl font-bold">
                {rules.filter(r => r.action === 'auto_reject').length}
              </p>
              <p className="text-sm text-muted-foreground">Auto-Reject</p>
            </CardContent>
          </Card>
        </div>

        {/* Rules by Type */}
        {(['leave', 'expense', 'discount'] as RequestType[]).map((type) => (
          <Card key={type}>
            <CardHeader>
              <CardTitle className="capitalize flex items-center gap-2">
                <Badge variant="outline" className={cn(
                  type === 'leave' && 'border-leave text-leave',
                  type === 'expense' && 'border-expense text-expense',
                  type === 'discount' && 'border-discount text-discount',
                )}>
                  {type}
                </Badge>
                Rules
              </CardTitle>
              <CardDescription>
                {rulesByType[type].length} rule{rulesByType[type].length !== 1 ? 's' : ''} configured
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[80px]">Rank</TableHead>
                    <TableHead>Condition</TableHead>
                    <TableHead>Target Grade</TableHead>
                    <TableHead>System Decision</TableHead>
                    <TableHead>Active</TableHead>
                    <TableHead className="text-right">Manage</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {rulesByType[type].sort((a, b) => a.priority - b.priority).map((rule) => (
                    <TableRow key={rule.id}>
                      <TableCell>
                        <span className="text-sm text-muted-foreground font-medium">#{rule.priority}</span>
                      </TableCell>
                      <TableCell>
                        <code className="text-[11px] bg-muted px-2 py-1 rounded font-mono">
                          {typeof rule.condition === 'object' ? JSON.stringify(rule.condition) : rule.condition}
                        </code>
                      </TableCell>
                      <TableCell>
                        <Badge variant="secondary" className="font-normal">
                          {rule.gradeId === 1 ? 'Employee' : rule.gradeId === 2 ? 'Manager' : rule.gradeId === 3 ? 'Admin' : `Grade ${rule.gradeId}`}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Badge className={cn('gap-1', getActionBadgeVariant(rule.action))}>
                          {getActionIcon(rule.action)}
                          {getActionLabel(rule.action)}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Switch
                          checked={rule.isActive}
                          onCheckedChange={() => toggleRule(rule)}
                        />
                      </TableCell>
                      <TableCell className="text-right space-x-2">
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleEditRule(rule)}
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleDeleteRule(rule.id)}
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                  {rulesByType[type].length === 0 && (
                    <TableRow>
                      <TableCell colSpan={5} className="text-center py-8 text-muted-foreground">
                        No rules configured for {type} requests
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        ))}

        {/* Add/Edit Rule Dialog */}
        <Dialog open={isAddDialogOpen || isEditDialogOpen} onOpenChange={(open) => {
          if (!open) {
            setIsAddDialogOpen(false);
            setIsEditDialogOpen(false);
            resetForm();
          }
        }}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{isAddMode ? 'Add New Rule' : 'Edit Rule'}</DialogTitle>
              <DialogDescription>
                {isAddMode ? 'Create a new automatic approval rule' : 'Update the approval rule'}
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4">
              <div className="space-y-2">
                <Label>Request Type</Label>
                <Select
                  value={newRule.requestType}
                  onValueChange={(v) => setNewRule({ ...newRule, requestType: v as RequestType })}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="bg-popover">
                    <SelectItem value="leave">Leave</SelectItem>
                    <SelectItem value="expense">Expense</SelectItem>
                    <SelectItem value="discount">Discount</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label>Condition</Label>
                <Input
                  placeholder='e.g., {"max_amount": 10000}'
                  value={
                    typeof newRule.condition === 'object'
                      ? JSON.stringify(newRule.condition)
                      : newRule.condition
                  }
                  onChange={(e) => {
                    try {
                      const val = JSON.parse(e.target.value);
                      setNewRule({ ...newRule, condition: val });
                    } catch {
                      setNewRule({ ...newRule, condition: e.target.value });
                    }
                  }}
                />
                <p className="text-xs text-muted-foreground">
                  Enter condition as JSON: {"{ \"max_amount\": 5000 }"} or{" "}
                  {"{ \"max_days\": 3 }"}
                </p>
              </div>

              <div className="space-y-2">
                <Label>Action</Label>
                <Select
                  value={newRule.action}
                  onValueChange={(v) =>
                    setNewRule({ ...newRule, action: v as ApprovalRule['action'] })
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="bg-popover">
                    <SelectItem value="AUTO_APPROVE">Auto Approve</SelectItem>
                    <SelectItem value="AUTO_REJECT">Auto Reject</SelectItem>
                    <SelectItem value="ASSIGN_APPROVER">
                      Assign Approver
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label>Employee Grade</Label>
                <Select
                  value={String(newRule.gradeId)}
                  onValueChange={(v) =>
                    setNewRule({ ...newRule, gradeId: parseInt(v) })
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="bg-popover">
                    <SelectItem value="1">Grade 1 (Employee)</SelectItem>
                    <SelectItem value="2">Grade 2 (Manager)</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label>Priority</Label>
                <Input
                  type="number"
                  min={1}
                  value={newRule.priority}
                  onChange={(e) =>
                    setNewRule({
                      ...newRule,
                      priority: parseInt(e.target.value) || 1,
                    })
                  }
                />
                <p className="text-xs text-muted-foreground">
                  Lower numbers = higher priority.
                </p>
              </div>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => {
                setIsAddDialogOpen(false);
                setIsEditDialogOpen(false);
                resetForm();
              }}>
                Cancel
              </Button>
              <Button onClick={handleSaveRule}>
                {isAddMode ? 'Add Rule' : 'Save Changes'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </AppLayout>
  );
}
