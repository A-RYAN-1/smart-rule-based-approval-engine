import { AppLayout } from '@/components/layout/AppLayout';
import { useAuth } from '@/contexts/AuthContext';
import { useAdmin } from '@/hooks/useAdmin';
import { StatCard } from '@/components/ui/stat-card';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  Legend,
} from 'recharts';
import { FileText, Clock, CheckCircle2, XCircle, Zap, TrendingUp } from 'lucide-react';
import { Navigate } from 'react-router-dom';

const COLORS = {
  approvedManual: 'hsl(142, 76%, 70%)',
  approvedAuto: 'hsl(142, 76%, 36%)',
  rejectedManual: 'hsl(0, 84%, 75%)',
  rejectedAuto: 'hsl(0, 84%, 60%)',
  pending: 'hsl(38, 92%, 50%)',
  cancelled: 'hsl(215, 16%, 47%)',
};

export default function ReportsPage() {
  const { user } = useAuth();
  const {
    statusDistribution,
    requestsByType,
    rules,
    isLoadingStatusDistribution,
    isLoadingRequestsByType,
    isLoadingRules
  } = useAdmin();

  // Debug: Log all query states
  console.log('[ReportsPage] Loading states:', {
    isLoadingStatusDistribution,
    isLoadingRequestsByType,
    isLoadingRules,
    statusDistribution: statusDistribution ? 'exists' : 'null',
    requestsByType: requestsByType ? `array[${requestsByType.length}]` : 'null',
    rules: rules ? `array[${rules.length}]` : 'null'
  });
  console.log('[ReportsPage] statusDistribution data:', statusDistribution);
  console.log('[ReportsPage] requestsByType data:', requestsByType);
  console.log('[ReportsPage] rules data:', rules);

  // Only admins can access
  if (user?.role !== 'admin') {
    return <Navigate to="/dashboard" replace />;
  }

  const isLoading = isLoadingStatusDistribution || isLoadingRequestsByType || isLoadingRules;

  console.log('[ReportsPage] Is Loading:', isLoading);

  // Derived Stats Logic
  const autoApprovedTotal = requestsByType?.reduce((acc, curr) => acc + (curr.auto_approved || 0), 0) || 0;
  const totalApproved = statusDistribution?.approved || 0;
  const manualApproved = Math.max(0, totalApproved - autoApprovedTotal);

  const manualRejected = statusDistribution?.rejected || 0;
  const autoRejected = statusDistribution?.auto_rejected || 0;
  const totalRejected = manualRejected + autoRejected;

  // Unified Total calculation (Sum of all statuses in distribution)
  const processedTotal = totalApproved + totalRejected + (statusDistribution?.cancelled || 0);
  const pendingRequests = statusDistribution?.pending || 0;
  const totalRequests = processedTotal + pendingRequests;

  const totalAutoProcessed = autoApprovedTotal + autoRejected;
  const autoProcessingRate = processedTotal > 0
    ? Math.round((totalAutoProcessed / processedTotal) * 100)
    : 0;

  console.log('[ReportsPage] Calculated stats:', {
    autoApprovedTotal,
    totalApproved,
    manualApproved,
    totalRequests,
    pendingRequests,
    autoProcessingRate
  });

  // Status distribution data for Pie Chart
  const statusData = [
    { name: 'Approved (Manual)', value: manualApproved, color: COLORS.approvedManual },
    { name: 'Approved (Auto)', value: autoApprovedTotal, color: COLORS.approvedAuto },
    { name: 'Rejected (Manual)', value: manualRejected, color: COLORS.rejectedManual },
    { name: 'Rejected (Auto)', value: autoRejected, color: COLORS.rejectedAuto },
    { name: 'Pending', value: pendingRequests, color: COLORS.pending },
    { name: 'Cancelled', value: statusDistribution?.cancelled || 0, color: COLORS.cancelled },
  ].filter(s => s.value > 0);

  // Request type distribution data for Bar Chart
  const typeData = (requestsByType || []).map(item => ({
    name: item.type.charAt(0).toUpperCase() + item.type.slice(1).toLowerCase(),
    count: item.total_requests,
    auto: item.auto_approved
  }));

  const stats = {
    totalRequests,
    pendingRequests,
    autoApproved: autoApprovedTotal,
    autoRejected
  };

  const activeRules = rules?.filter(r => r.isActive).length || 0;

  if (isLoading) {
    console.log('[ReportsPage] Showing loading spinner');
    return (
      <AppLayout>
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
        </div>
      </AppLayout>
    );
  }

  console.log('[ReportsPage] Rendering content page');
  return (
    <AppLayout>
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Reports & Analytics</h1>
          <p className="text-muted-foreground mt-1">Insights into approval patterns and rule effectiveness</p>
        </div>

        {/* Key Metrics */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <StatCard
            title="Total Requests"
            value={stats.totalRequests}
            icon={FileText}
            variant="primary"
          />
          <StatCard
            title="Pending Review"
            value={stats.pendingRequests}
            icon={Clock}
            variant="warning"
          />
          <StatCard
            title="Auto-Processed"
            value={`${autoProcessingRate}%`}
            subtitle={`${totalAutoProcessed} requests`}
            icon={Zap}
            variant="success"
          />
          <StatCard
            title="Active Rules"
            value={activeRules}
            subtitle={`${rules.length} total`}
            icon={TrendingUp}
            variant="default"
          />
        </div>

        <div className="grid gap-6 lg:grid-cols-2">
          {/* Status Distribution Pie Chart */}
          <Card>
            <CardHeader>
              <CardTitle>Request Status Distribution</CardTitle>
              <CardDescription>Breakdown of all requests by current status</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="h-[300px]">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={statusData}
                      cx="50%"
                      cy="50%"
                      innerRadius={60}
                      outerRadius={100}
                      paddingAngle={2}
                      dataKey="value"
                    >
                      {statusData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={entry.color} />
                      ))}
                    </Pie>
                    <Tooltip
                      contentStyle={{
                        backgroundColor: 'hsl(var(--card))',
                        border: '1px solid hsl(var(--border))',
                        borderRadius: '8px',
                      }}
                    />
                    <Legend />
                  </PieChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>

          {/* Request Types Bar Chart */}
          <Card>
            <CardHeader>
              <CardTitle>Requests by Type</CardTitle>
              <CardDescription>Total vs. auto-processed requests</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="h-[300px]">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={typeData}>
                    <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                    <XAxis
                      dataKey="name"
                      className="text-muted-foreground"
                      tick={{ fill: 'hsl(var(--muted-foreground))' }}
                    />
                    <YAxis
                      className="text-muted-foreground"
                      tick={{ fill: 'hsl(var(--muted-foreground))' }}
                    />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: 'hsl(var(--card))',
                        border: '1px solid hsl(var(--border))',
                        borderRadius: '8px',
                      }}
                    />
                    <Bar dataKey="count" fill="hsl(var(--primary))" name="Total" radius={[4, 4, 0, 0]} />
                    <Bar dataKey="auto" fill="hsl(var(--status-approved))" name="Auto-processed" radius={[4, 4, 0, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Rule Performance */}
        <Card>
          <CardHeader>
            <CardTitle>Rule Performance Summary</CardTitle>
            <CardDescription>How your approval rules are performing</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-6 md:grid-cols-3">
              <div className="p-6 rounded-lg bg-status-approved-bg text-center">
                <CheckCircle2 className="h-8 w-8 mx-auto mb-3 text-status-approved" />
                <p className="text-3xl font-bold text-status-approved">{stats.autoApproved}</p>
                <p className="text-sm text-muted-foreground mt-1">Auto-Approved</p>
                <p className="text-xs text-muted-foreground">
                  by {rules.filter(r => r.action === 'auto_approve' && r.isActive).length} active rules
                </p>
              </div>
              <div className="p-6 rounded-lg bg-status-rejected-bg text-center">
                <XCircle className="h-8 w-8 mx-auto mb-3 text-status-rejected" />
                <p className="text-3xl font-bold text-status-rejected">{stats.autoRejected}</p>
                <p className="text-sm text-muted-foreground mt-1">Auto-Rejected</p>
                <p className="text-xs text-muted-foreground">
                  by {rules.filter(r => r.action === 'auto_reject' && r.isActive).length} active rules
                </p>
              </div>
              <div className="p-6 rounded-lg bg-status-pending-bg text-center">
                <Clock className="h-8 w-8 mx-auto mb-3 text-status-pending" />
                <p className="text-3xl font-bold text-status-pending">{stats.pendingRequests}</p>
                <p className="text-sm text-muted-foreground mt-1">Manual Review</p>
                <p className="text-xs text-muted-foreground">
                  assigned to approvers
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}
