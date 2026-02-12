import { AppLayout } from '@/components/layout/AppLayout';
import { useAuth } from '@/contexts/AuthContext';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { 
  Table, 
  TableBody, 
  TableCell, 
  TableHead, 
  TableHeader, 
  TableRow 
} from '@/components/ui/table';
import { Users, Shield, User, Briefcase } from 'lucide-react';
import { Navigate } from 'react-router-dom';
import { cn } from '@/lib/utils';

const mockUsers = [
  { id: 1, name: 'Employee One', email: 'employee@test.com', role: 'employee', department: 'Engineering' },
  { id: 2, name: 'Lee Manager', email: 'manager@company.com', role: 'manager', department: 'Operations' },
  { id: 3, name: 'Sophia Admin', email: 'admin@company.com', role: 'admin', department: 'HR' },
  { id: 4, name: 'John Developer', email: 'john@company.com', role: 'employee', department: 'Engineering' },
  { id: 5, name: 'Sarah Designer', email: 'sarah@company.com', role: 'employee', department: 'Design' },
];

export default function UsersPage() {
  const { user } = useAuth();

  // Only admins can access
  if (user?.role !== 'admin') {
    return <Navigate to="/dashboard" replace />;
  }

  const getRoleBadgeStyle = (role: string) => {
    switch (role) {
      case 'admin': return 'bg-primary/10 text-primary border-primary/20';
      case 'manager': return 'bg-status-pending-bg text-status-pending border-status-pending/20';
      default: return 'bg-muted text-muted-foreground';
    }
  };

  const getRoleIcon = (role: string) => {
    switch (role) {
      case 'admin': return <Shield className="h-3 w-3" />;
      case 'manager': return <Briefcase className="h-3 w-3" />;
      default: return <User className="h-3 w-3" />;
    }
  };

  return (
    <AppLayout>
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">User Management</h1>
          <p className="text-muted-foreground mt-1">View and manage system users</p>
        </div>

        {/* Stats */}
        <div className="grid gap-4 md:grid-cols-3">
          <Card>
            <CardContent className="p-6">
              <div className="flex items-center gap-4">
                <div className="p-3 rounded-xl bg-primary/10">
                  <Users className="h-6 w-6 text-primary" />
                </div>
                <div>
                  <p className="text-2xl font-bold">{mockUsers.length}</p>
                  <p className="text-sm text-muted-foreground">Total Users</p>
                </div>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              <div className="flex items-center gap-4">
                <div className="p-3 rounded-xl bg-status-pending-bg">
                  <Briefcase className="h-6 w-6 text-status-pending" />
                </div>
                <div>
                  <p className="text-2xl font-bold">
                    {mockUsers.filter(u => u.role === 'manager').length}
                  </p>
                  <p className="text-sm text-muted-foreground">Managers</p>
                </div>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              <div className="flex items-center gap-4">
                <div className="p-3 rounded-xl bg-muted">
                  <User className="h-6 w-6 text-muted-foreground" />
                </div>
                <div>
                  <p className="text-2xl font-bold">
                    {mockUsers.filter(u => u.role === 'employee').length}
                  </p>
                  <p className="text-sm text-muted-foreground">Employees</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Users Table */}
        <Card>
          <CardHeader>
            <CardTitle>All Users</CardTitle>
            <CardDescription>System users and their roles</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>User</TableHead>
                  <TableHead>Email</TableHead>
                  <TableHead>Department</TableHead>
                  <TableHead>Role</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {mockUsers.map((u) => (
                  <TableRow key={u.id}>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <div className="w-10 h-10 rounded-full bg-muted flex items-center justify-center">
                          <span className="text-sm font-medium">{u.name.charAt(0)}</span>
                        </div>
                        <span className="font-medium">{u.name}</span>
                      </div>
                    </TableCell>
                    <TableCell className="text-muted-foreground">{u.email}</TableCell>
                    <TableCell>{u.department}</TableCell>
                    <TableCell>
                      <Badge 
                        variant="outline" 
                        className={cn('gap-1 capitalize', getRoleBadgeStyle(u.role))}
                      >
                        {getRoleIcon(u.role)}
                        {u.role}
                      </Badge>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        {/* Info Note */}
        <Card className="bg-muted/50">
          <CardContent className="p-4">
            <p className="text-sm text-muted-foreground">
              <strong>Note:</strong> User management is for viewing purposes in this demo. 
              In a production environment, you would be able to add, edit, and remove users, 
              as well as assign roles and departments.
            </p>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}
