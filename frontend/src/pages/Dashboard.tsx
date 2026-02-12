import { useAuth } from '@/contexts/AuthContext';
import { useLeaves } from '@/hooks/useLeaves';
import { useExpenses } from '@/hooks/useExpenses';
import { useDiscounts } from '@/hooks/useDiscounts';
import { useBalances } from '@/hooks/useBalances';
import { AppLayout } from '@/components/layout/AppLayout';
import { StatCard } from '@/components/ui/stat-card';
import { StatusBadge } from '@/components/ui/status-badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import {
  FileText,
  Clock,
  CheckCircle2,
  XCircle,
  CalendarDays,
  Receipt,
  Percent,
  ArrowRight,
  Shield,
  BarChart3,
  Zap,
  Plus
} from 'lucide-react';
import { Link } from 'react-router-dom';
import { format, differenceInDays, isAfter } from 'date-fns';
import { useMemo, useState } from 'react';
import { useAdmin } from '@/hooks/useAdmin';
import { cn } from '@/lib/utils';
import {
  PieChart,
  Pie,
  Cell,
  Tooltip,
  Legend,
  ResponsiveContainer
} from 'recharts';
import { BalanceUsageDialog } from '@/components/balances/BalanceUsageDialog';
import { RequestDetailDialog } from '@/components/dashboard/RequestDetailDialog';

const COLORS = {
  approvedManual: 'hsl(142, 70%, 55%)',
  approvedAuto: 'hsl(142, 76%, 36%)',
  rejectedManual: 'hsl(0, 70%, 65%)',
  rejectedAuto: 'hsl(0, 84%, 60%)',
  pending: 'hsl(38, 92%, 50%)',
  cancelled: 'hsl(215, 16%, 47%)',
};

const getRequestDetails = (request: any) => {
  if (request.type === 'leave') {
    return `${request.leaveType} • ${format(new Date(request.fromDate), 'MMM d')} - ${format(new Date(request.toDate), 'MMM d, yyyy')}`;
  } else if (request.type === 'expense') {
    return `${request.category} • $${request.amount.toLocaleString()}`;
  } else {
    return `${request.discountPercentage}% discount request`;
  }
};

export default function Dashboard() {
  const { user } = useAuth();
  const { myLeaves, pendingLeaves } = useLeaves();
  const { myExpenses, pendingExpenses } = useExpenses();
  const { myDiscounts, pendingDiscounts } = useDiscounts();
  const { balances: unifiedBalances, isLoading: isLoadingBalances } = useBalances();
  const { rules, holidays, statusDistribution, requestsByType } = useAdmin();

  const isManager = user?.role === 'manager';
  const isAdmin = user?.role === 'admin';
  const isApprover = isManager || isAdmin;

  const [usageDialogOpen, setUsageDialogOpen] = useState(false);
  const [usageType, setUsageType] = useState<'leave' | 'expense' | 'discount'>('leave');
  const [usageTitle, setUsageTitle] = useState('');

  const [detailDialogOpen, setDetailDialogOpen] = useState(false);
  const [selectedRequest, setSelectedRequest] = useState<any>(null);

  const stats = useMemo(() => {
    const pendingCount = pendingLeaves.length + pendingExpenses.length + pendingDiscounts.length;
    const personalAll = [...myLeaves, ...myExpenses, ...myDiscounts];

    // Admin sees system-wide metrics
    if (isAdmin) {
      if (!statusDistribution) {
        return {
          totalRequests: pendingCount,
          pendingRequests: pendingCount,
          approvedRequests: 0,
          rejectedRequests: 0,
          cancelledRequests: 0,
          manualApproved: 0,
          autoApproved: 0,
          manualRejected: 0,
          autoRejected: 0,
          isLoading: true
        };
      }

      const totalApproved = statusDistribution.approved || 0;
      const autoApproved = requestsByType?.reduce((acc, curr) => acc + (curr.auto_approved || 0), 0) || 0;
      const manualApproved = Math.max(0, totalApproved - autoApproved);

      const manualRejected = statusDistribution.rejected || 0;
      const autoRejected = statusDistribution.auto_rejected || 0;
      const totalRejected = manualRejected + autoRejected;

      const cancelled = statusDistribution.cancelled || 0;
      const processedTotal = totalApproved + totalRejected + cancelled;

      return {
        totalRequests: processedTotal + pendingCount,
        pendingRequests: pendingCount,
        approvedRequests: totalApproved,
        rejectedRequests: totalRejected,
        cancelledRequests: cancelled,
        manualApproved,
        autoApproved,
        manualRejected,
        autoRejected,
        isLoading: false
      };
    }

    // Managers & Employees see THEIR OWN requests in top cards
    return {
      totalRequests: personalAll.length,
      pendingRequests: personalAll.filter(r => r.status === 'pending').length,
      approvedRequests: personalAll.filter(r => r.status === 'approved' || r.status === 'auto_approved').length,
      rejectedRequests: personalAll.filter(r => r.status === 'rejected' || r.status === 'auto_rejected').length,
      cancelledRequests: personalAll.filter(r => r.status === 'cancelled').length,
      manualApproved: personalAll.filter(r => r.status === 'approved').length,
      autoApproved: personalAll.filter(r => r.status === 'auto_approved').length,
      manualRejected: personalAll.filter(r => r.status === 'rejected').length,
      autoRejected: personalAll.filter(r => r.status === 'auto_rejected').length,
      isLoading: false
    };
  }, [myLeaves, myExpenses, myDiscounts, pendingLeaves, pendingExpenses, pendingDiscounts, isAdmin, isManager, statusDistribution, requestsByType]);

  const recentRequests = useMemo(() => {
    const list = isApprover
      ? [
        ...pendingLeaves.map(r => ({ ...r, type: 'leave' as const })),
        ...pendingExpenses.map(r => ({ ...r, type: 'expense' as const })),
        ...pendingDiscounts.map(r => ({ ...r, type: 'discount' as const })),
      ]
      : [
        ...myLeaves.map(r => ({ ...r, type: 'leave' as const })),
        ...myExpenses.map(r => ({ ...r, type: 'expense' as const })),
        ...myDiscounts.map(r => ({ ...r, type: 'discount' as const })),
      ];

    return list
      .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
      .slice(0, 5);
  }, [myLeaves, myExpenses, myDiscounts, pendingLeaves, pendingExpenses, pendingDiscounts, isApprover, isAdmin, statusDistribution]);

  const statusData = useMemo(() => {
    // Admin uses System Stats
    if (isAdmin && statusDistribution) {
      const data = [];
      if (stats.manualApproved > 0) data.push({ name: 'Approved (Manual)', value: stats.manualApproved, color: COLORS.approvedManual });
      if (stats.autoApproved > 0) data.push({ name: 'Approved (Auto)', value: stats.autoApproved, color: COLORS.approvedAuto });
      if (stats.manualRejected > 0) data.push({ name: 'Rejected (Manual)', value: stats.manualRejected, color: COLORS.rejectedManual });
      if (stats.autoRejected > 0) data.push({ name: 'Rejected (Auto)', value: stats.autoRejected, color: COLORS.rejectedAuto });
      if (stats.pendingRequests > 0) data.push({ name: 'Pending', value: stats.pendingRequests, color: COLORS.pending });
      if (stats.cancelledRequests > 0) data.push({ name: 'Cancelled', value: stats.cancelledRequests, color: COLORS.cancelled });
      return data;
    }

    // Manager/Employee uses Personal Stats
    const data = [];
    if (stats.manualApproved > 0) data.push({ name: 'Approved (Manual)', value: stats.manualApproved, color: COLORS.approvedManual });
    if (stats.autoApproved > 0) data.push({ name: 'Approved (Auto)', value: stats.autoApproved, color: COLORS.approvedAuto });
    if (stats.manualRejected > 0) data.push({ name: 'Rejected (Manual)', value: stats.manualRejected, color: COLORS.rejectedManual });
    if (stats.autoRejected > 0) data.push({ name: 'Rejected (Auto)', value: stats.autoRejected, color: COLORS.rejectedAuto });
    if (stats.pendingRequests > 0) data.push({ name: 'Pending', value: stats.pendingRequests, color: COLORS.pending });
    if (stats.cancelledRequests > 0) data.push({ name: 'Cancelled', value: stats.cancelledRequests, color: COLORS.cancelled });
    return data;
  }, [isAdmin, stats, statusDistribution]);

  const getUsageData = (type: 'leave' | 'expense' | 'discount') => {
    if (type === 'leave') {
      return myLeaves
        .filter(r => r.status === 'approved' || r.status === 'auto_approved')
        .map(r => ({
          id: r.id,
          date: r.createdAt,
          description: `${r.leaveType} • ${format(new Date(r.fromDate), 'MMM d')} - ${format(new Date(r.toDate), 'MMM d')}`,
          amount: differenceInDays(new Date(r.toDate), new Date(r.fromDate)) + 1
        }));
    } else if (type === 'expense') {
      return myExpenses
        .filter(r => r.status === 'approved' || r.status === 'auto_approved')
        .map(r => ({
          id: r.id,
          date: r.createdAt,
          description: r.category,
          amount: r.amount
        }));
    } else {
      return myDiscounts
        .filter(r => r.status === 'approved' || r.status === 'auto_approved')
        .map(r => ({
          id: r.id,
          date: r.createdAt,
          description: 'Benefit Discount',
          amount: r.discountPercentage
        }));
    }
  };

  return (
    <AppLayout>
      <div className="space-y-10">
        {/* Header Block */}
        <div className="flex flex-col gap-1">
          <h1 className="text-5xl font-black tracking-tight text-foreground letter-spacing-tight">
            {isManager ? 'Manager Overview' : `Welcome, ${user?.name || 'User'}!`}
          </h1>
          <p className="text-muted-foreground font-semibold text-lg">
            {isManager
              ? `There are ${stats.pendingRequests} team requests awaiting your approval.`
              : `You have ${stats.pendingRequests} pending requests requiring attention.`}
          </p>
        </div>

        <div className="space-y-8">
          {/* Top Stats */}
          <div className="grid gap-4 grid-cols-2 md:grid-cols-5">
            <StatCard title="Total" value={stats.totalRequests} icon={FileText} variant="primary" />
            <StatCard title="Pending" value={stats.pendingRequests} icon={Clock} variant="warning" />
            <StatCard title="Approved" value={stats.approvedRequests} icon={CheckCircle2} variant="success" />
            <StatCard title="Rejected" value={stats.rejectedRequests} icon={XCircle} variant="danger" />
            <StatCard title="Cancelled" value={stats.cancelledRequests} icon={XCircle} variant="default" />
          </div>

          {/* Quick Actions / Manager Tasks - Hidden for Admin */}
          {!isAdmin && (
            isManager ? (
              <div className="bg-primary/5 p-6 rounded-3xl border border-primary/20 backdrop-blur-sm">
                <div className="flex flex-col md:flex-row items-center justify-between gap-6">
                  <div className="flex flex-col">
                    <h3 className="text-lg font-black uppercase tracking-widest text-primary px-1">Manager Tasks</h3>
                    <p className="text-sm text-muted-foreground px-1 font-medium">Quickly review and process team requests</p>
                  </div>
                  <Button size="lg" asChild className="rounded-2xl bg-primary shadow-lg shadow-primary/20 hover:scale-[1.02] transition-all font-black px-12 h-14 text-lg">
                    <Link to="/pending-approvals" className="flex items-center gap-2">
                      Review Pending Approvals <ArrowRight className="h-5 w-5" />
                    </Link>
                  </Button>
                </div>
              </div>
            ) : (
              <div className="bg-muted/30 p-4 rounded-3xl border border-border/50 backdrop-blur-sm">
                <div className="flex flex-col md:flex-row items-center justify-between gap-4">
                  <div className="flex flex-col">
                    <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground px-1">Quick Actions</h3>
                    <p className="text-xs text-muted-foreground/60 px-1 font-medium">Submit a new request in seconds</p>
                  </div>
                  <div className="flex flex-wrap gap-3">
                    <Button size="lg" variant="outline" asChild className="rounded-2xl bg-background shadow-sm hover:bg-primary hover:text-primary-foreground hover:shadow-lg hover:shadow-primary/20 transition-all font-bold px-8">
                      <Link to="/leaves">Apply Leave</Link>
                    </Button>
                    <Button size="lg" variant="outline" asChild className="rounded-2xl bg-background shadow-sm hover:bg-primary hover:text-primary-foreground hover:shadow-lg hover:shadow-primary/20 transition-all font-bold px-8">
                      <Link to="/expenses">Claim Expense</Link>
                    </Button>
                    <Button size="lg" variant="outline" asChild className="rounded-2xl bg-background shadow-sm hover:bg-primary hover:text-primary-foreground hover:shadow-lg hover:shadow-primary/20 transition-all font-bold px-8">
                      <Link to="/discounts">Get Discount</Link>
                    </Button>
                  </div>
                </div>
              </div>
            )
          )}
        </div>

        <div className="grid gap-8 lg:grid-cols-12">
          {/* Recent Activity (LHS) */}
          <div className="lg:col-span-8 flex flex-col gap-6">
            <Card className="border-none shadow-none bg-transparent">
              <CardHeader className="px-0 pt-0">
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle className="text-xl font-bold">
                      {isApprover ? 'Recent Team Activity' : 'Recent Activity'}
                    </CardTitle>
                    <CardDescription>
                      {isApprover ? 'Lately submitted requests from your team' : 'Track your latest request statuses'}
                    </CardDescription>
                  </div>
                  <Button variant="ghost" size="sm" asChild className="text-primary hover:text-primary hover:bg-primary/10">
                    <Link to={isApprover ? "/pending-approvals" : "/my-requests"} className="flex items-center gap-1 font-bold">
                      {isApprover ? 'View All' : 'History'} <ArrowRight className="h-4 w-4" />
                    </Link>
                  </Button>
                </div>
              </CardHeader>
              <CardContent className="px-0">
                <div className="grid gap-3">
                  {recentRequests.map((request) => (
                    <button
                      key={`${request.type}-${request.id}`}
                      onClick={() => {
                        setSelectedRequest(request);
                        setDetailDialogOpen(true);
                      }}
                      className="flex items-center justify-between p-4 rounded-2xl bg-card border border-border/40 hover:border-primary/40 hover:shadow-md transition-all group text-left w-full"
                    >
                      <div className="flex items-center gap-4">
                        <div className={`p-3 rounded-xl transition-colors ${request.type === 'leave' ? 'bg-leave-bg text-leave' :
                          request.type === 'expense' ? 'bg-expense-bg text-expense' : 'bg-discount-bg text-discount'
                          }`}>
                          {request.type === 'leave' && <CalendarDays className="h-5 w-5" />}
                          {request.type === 'expense' && <Receipt className="h-5 w-5" />}
                          {request.type === 'discount' && <Percent className="h-5 w-5" />}
                        </div>
                        <div className="space-y-0.5">
                          <p className="font-bold text-sm capitalize group-hover:text-primary transition-colors">
                            {request.type} Request
                          </p>
                          <p className="text-xs text-muted-foreground flex items-center gap-1.5 font-medium">
                            {getRequestDetails(request)}
                          </p>
                        </div>
                      </div>
                      <div className="flex items-center gap-3">
                        <StatusBadge status={request.status} />
                        <ArrowRight className="h-4 w-4 opacity-0 group-hover:opacity-100 -translate-x-2 group-hover:translate-x-0 transition-all text-muted-foreground" />
                      </div>
                    </button>
                  ))}
                  {recentRequests.length === 0 && (
                    <div className="text-center py-12 rounded-2xl border-2 border-dashed border-muted">
                      <Clock className="mx-auto h-12 w-12 text-muted-foreground opacity-20 mb-3" />
                      <p className="text-muted-foreground font-medium">No activity yet</p>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>

            {isAdmin && (
              <div className="grid grid-cols-1 gap-6">
                <Card className="rounded-3xl border-border/40 shadow-sm transition-all hover:shadow-md cursor-default text-black">
                  <CardHeader>
                    <div className="flex items-center justify-between">
                      <CardTitle className="text-lg font-bold">System Optimization</CardTitle>
                      <Zap className="h-5 w-5 text-yellow-500 fill-yellow-500" />
                    </div>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-4">
                      <div className="flex justify-between items-center text-sm">
                        <span className="text-muted-foreground">Auto-Processing Efficiency</span>
                        <span className="font-bold text-green-600">
                          {stats.totalRequests > 0 ? Math.round((stats.autoApproved + stats.autoRejected) / (stats.totalRequests - stats.pendingRequests) * 100) : 0}%
                        </span>
                      </div>
                      <div className="w-full bg-muted rounded-full h-2">
                        <div
                          className="bg-green-500 h-2 rounded-full transition-all duration-1000"
                          style={{ width: `${stats.totalRequests > 0 ? ((stats.autoApproved + stats.autoRejected) / (stats.totalRequests - stats.pendingRequests)) * 100 : 0}%` }}
                        />
                      </div>
                      <p className="text-xs text-muted-foreground italic">Percentage of requests handled automatically</p>
                    </div>
                  </CardContent>
                </Card>
              </div>
            )}
            {isManager && (
              <div className="space-y-6">
              </div>
            )}
          </div>

          {/* Sidebar (RHS) */}
          <div className="lg:col-span-4 space-y-6">
            <Card className="rounded-3xl border-border/40 overflow-hidden shadow-sm">
              <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle className="text-lg font-bold">Insights</CardTitle>
                    <CardDescription>
                      {isApprover ? (statusDistribution ? 'System-wide activity distribution' : 'My Requests Status') : 'Request Distribution'}
                    </CardDescription>
                  </div>
                  {statusDistribution && (
                    <div className="bg-primary/10 text-primary text-[10px] font-bold px-2 py-1 rounded-full uppercase tracking-wider">
                      Live
                    </div>
                  )}
                </div>
              </CardHeader>
              <CardContent>
                <div className="h-[320px] -mt-2">
                  {stats.totalRequests > 0 ? (
                    <ResponsiveContainer width="100%" height="100%">
                      <PieChart>
                        <Pie
                          data={statusData}
                          cx="50%"
                          cy="45%"
                          innerRadius={65}
                          outerRadius={90}
                          paddingAngle={8}
                          dataKey="value"
                          stroke="none"
                          cornerRadius={6}
                        >
                          {statusData.map((entry, index) => (
                            <Cell key={`cell-${index}`} fill={entry.color} />
                          ))}
                        </Pie>
                        <Tooltip
                          content={({ active, payload }) => {
                            if (active && payload && payload.length) {
                              const data = payload[0].payload;
                              const percentage = Math.round((data.value / stats.totalRequests) * 100);
                              return (
                                <div className="bg-background/95 backdrop-blur-sm border border-border px-3 py-2 rounded-xl shadow-xl">
                                  <p className="text-xs font-bold text-foreground mb-1">{data.name}</p>
                                  <div className="flex items-center gap-2">
                                    <div className="h-2 w-2 rounded-full" style={{ backgroundColor: data.color }} />
                                    <p className="text-sm font-black">{data.value} <span className="text-[10px] text-muted-foreground font-medium">({percentage}%)</span></p>
                                  </div>
                                </div>
                              );
                            }
                            return null;
                          }}
                        />
                        <Legend
                          verticalAlign="bottom"
                          height={80}
                          content={({ payload }) => (
                            <div className="grid grid-cols-2 gap-x-4 gap-y-2 mt-4">
                              {payload?.map((entry: any, index) => (
                                <div key={index} className="flex items-center gap-2">
                                  <div className="h-2 w-2 rounded-full flex-shrink-0" style={{ backgroundColor: entry.color }} />
                                  <span className="text-[11px] font-medium text-muted-foreground truncate">{entry.value}</span>
                                  <span className="text-[11px] font-bold ml-auto">{statusData[index]?.value}</span>
                                </div>
                              ))}
                            </div>
                          )}
                        />
                      </PieChart>
                    </ResponsiveContainer>
                  ) : (
                    <div className="h-full flex flex-col items-center justify-center text-muted-foreground">
                      <Clock className="h-10 w-10 opacity-20 mb-3" />
                      <p className="text-sm font-bold">Awaiting Activity Data...</p>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>

            {/* Show My Quotas for everyone with a session EXCEPT Admin */}
            {(unifiedBalances || isLoadingBalances) && !isAdmin && (
              <Card className="rounded-3xl border-border/40 shadow-sm overflow-hidden text-black">
                <CardHeader className="pb-4">
                  <CardTitle className="text-lg font-bold">My Quotas</CardTitle>
                  <CardDescription>Click a bar for history</CardDescription>
                </CardHeader>
                <CardContent className="space-y-6 pb-6">
                  {unifiedBalances?.leaves.map((balance) => (
                    <button
                      key={balance.leaveType}
                      className="w-full text-left group space-y-2.5"
                      onClick={() => {
                        setUsageType('leave');
                        setUsageTitle(`Leaves (${balance.leaveType})`);
                        setUsageDialogOpen(true);
                      }}
                    >
                      <div className="flex items-center justify-between">
                        <span className="text-[10px] font-black uppercase tracking-widest text-muted-foreground flex items-center gap-1.5">
                          <div className="w-1.5 h-1.5 rounded-full bg-blue-500" />
                          Leave ({balance.leaveType.toLowerCase()})
                        </span>
                        <span className="text-xs font-black text-foreground bg-blue-50 px-2 py-0.5 rounded text-blue-700">
                          {balance.balance} / {balance.total} days
                        </span>
                      </div>
                      <div className="h-2 rounded-full bg-muted/60 overflow-hidden">
                        <div
                          className="h-full bg-blue-500 rounded-full transition-all duration-700"
                          style={{ width: `${(balance.balance / balance.total) * 100}%` }}
                        />
                      </div>
                    </button>
                  ))}

                  {unifiedBalances?.expenses && (
                    <button
                      className="w-full text-left group space-y-2.5"
                      onClick={() => {
                        setUsageType('expense');
                        setUsageTitle('Expenses Pool');
                        setUsageDialogOpen(true);
                      }}
                    >
                      <div className="flex items-center justify-between">
                        <span className="text-[10px] font-black uppercase tracking-widest text-muted-foreground flex items-center gap-1.5">
                          <div className="w-1.5 h-1.5 rounded-full bg-green-500" />
                          Expenses
                        </span>
                        <span className="text-xs font-black text-foreground bg-green-50 px-2 py-0.5 rounded text-green-700">
                          {unifiedBalances.expenses.remaining.toLocaleString()} / {unifiedBalances.expenses.total.toLocaleString()} $
                        </span>

                      </div>
                      <div className="h-2 rounded-full bg-muted/60 overflow-hidden">
                        <div
                          className="h-full bg-green-500 rounded-full transition-all duration-700"
                          style={{ width: `${(unifiedBalances.expenses.remaining / unifiedBalances.expenses.total) * 100}%` }}
                        />
                      </div>
                    </button>
                  )}

                  {unifiedBalances?.discounts && (
                    <button
                      className="w-full text-left group space-y-2.5"
                      onClick={() => {
                        setUsageType('discount');
                        setUsageTitle('Discount Quota');
                        setUsageDialogOpen(true);
                      }}
                    >
                      <div className="flex items-center justify-between">
                        <span className="text-[10px] font-black uppercase tracking-widest text-muted-foreground flex items-center gap-1.5">
                          <div className="w-1.5 h-1.5 rounded-full bg-purple-500" />
                          Discount
                        </span>
                        <span className="text-xs font-black text-foreground bg-purple-50 px-2 py-0.5 rounded text-purple-700">
                          {unifiedBalances.discounts.remaining} / {unifiedBalances.discounts.total} %
                        </span>

                      </div>
                      <div className="h-2 rounded-full bg-muted/60 overflow-hidden">
                        <div
                          className="h-full bg-purple-500 rounded-full transition-all duration-700"
                          style={{ width: `${(unifiedBalances.discounts.remaining / unifiedBalances.discounts.total) * 100}%` }}
                        />
                      </div>
                    </button>
                  )}

                  {!unifiedBalances && isLoadingBalances && (
                    <div className="space-y-6">
                      {[1, 2, 3].map(i => (
                        <div key={i} className="animate-pulse space-y-2">
                          <div className="h-3 bg-muted rounded w-1/3" />
                          <div className="h-2.5 bg-muted rounded full" />
                        </div>
                      ))}
                    </div>
                  )}
                </CardContent>
              </Card>
            )}

            {/* Admin Fast-Access Modules */}
            {isAdmin && (
              <div className="space-y-6">
                <Card className="rounded-3xl border-primary/20 bg-primary/5 shadow-sm overflow-hidden">
                  <CardHeader className="pb-3 px-6 pt-6">
                    <div className="flex items-center gap-2">
                      <Shield className="h-5 w-5 text-primary" />
                      <CardTitle className="text-lg font-bold">Rules Engine</CardTitle>
                    </div>
                  </CardHeader>
                  <CardContent className="px-6 pb-6">
                    <p className="text-sm text-muted-foreground mb-4">
                      You have <span className="font-bold text-foreground">{rules.length} active rules</span> controlling automation.
                    </p>
                    <Button variant="outline" size="sm" asChild className="w-full rounded-xl font-bold bg-background">
                      <Link to="/admin/rules">Manage Rules</Link>
                    </Button>
                  </CardContent>
                </Card>

                <Card className="rounded-3xl border-border/40 shadow-sm overflow-hidden">
                  <CardHeader className="pb-3 px-6 pt-6">
                    <div className="flex items-center gap-2">
                      <CalendarDays className="h-5 w-5 text-sidebar-primary" />
                      <CardTitle className="text-lg font-bold">Next Holiday</CardTitle>
                    </div>
                  </CardHeader>
                  <CardContent className="px-6 pb-6 text-black">
                    {holidays.filter(h => isAfter(new Date(h.date), new Date())).sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime()).slice(0, 1).map(h => (
                      <div key={h.id} className="space-y-1">
                        <p className="font-bold text-base">{h.description}</p>
                        <p className="text-xs text-muted-foreground">{format(new Date(h.date), 'EEEE, MMMM do')}</p>
                      </div>
                    ))}
                    {holidays.length === 0 && <p className="text-xs text-muted-foreground italic">No upcoming holidays</p>}
                    <Button variant="ghost" size="sm" asChild className="w-full mt-4 rounded-xl font-bold hover:bg-sidebar-accent">
                      <Link to="/admin/holidays">View Calendar</Link>
                    </Button>
                  </CardContent>
                </Card>
              </div>
            )}
          </div>
        </div>
      </div>

      <BalanceUsageDialog
        isOpen={usageDialogOpen}
        onClose={() => setUsageDialogOpen(false)}
        title={usageTitle}
        type={usageType}
        usage={getUsageData(usageType)}
      />

      <RequestDetailDialog
        isOpen={detailDialogOpen}
        onClose={() => setDetailDialogOpen(false)}
        request={selectedRequest}
      />
    </AppLayout>
  );
}
