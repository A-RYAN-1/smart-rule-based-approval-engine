import {
  LayoutDashboard,
  CalendarDays,
  Receipt,
  Percent,
  Clock,
  Settings,
  Users,
  BarChart3,
  Shield,
  LogOut,
  ChevronLeft,
  ChevronRight
} from 'lucide-react';
import { NavLink, useLocation } from 'react-router-dom';
import { useAuth } from '@/contexts/AuthContext';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { useState } from 'react';

const employeeNav = [
  { title: 'Dashboard', url: '/dashboard', icon: LayoutDashboard },
  { title: 'Leave Requests', url: '/leaves', icon: CalendarDays },
  { title: 'Expense Claims', url: '/expenses', icon: Receipt },
  { title: 'Discount Requests', url: '/discounts', icon: Percent },
  { title: 'My Requests', url: '/my-requests', icon: Clock },
];

const approverNav = [
  { title: 'Pending Approvals', url: '/pending-approvals', icon: Clock },
];

const adminNav = [
  { title: 'Rules Management', url: '/admin/rules', icon: Shield },
  { title: 'Reports', url: '/admin/reports', icon: BarChart3 },
  { title: 'Holidays', url: '/admin/holidays', icon: CalendarDays },
];

export function AppSidebar() {
  const { user, logout } = useAuth();
  const location = useLocation();
  const [collapsed, setCollapsed] = useState(false);

  if (!user) return null;

  const isManager = user.role === 'manager';
  const isAdmin = user.role === 'admin';

  return (
    <aside
      className={cn(
        "h-full bg-sidebar text-sidebar-foreground flex flex-col transition-all duration-300 border-r border-sidebar-border",
        collapsed ? "w-16" : "w-64"
      )}
    >
      {/* Header */}
      <div className="p-4 border-b border-sidebar-border">
        <div className="flex items-center justify-between">
          {!collapsed && (
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-lg bg-sidebar-primary flex items-center justify-center">
                <Shield className="w-5 h-5 text-sidebar-primary-foreground" />
              </div>
              <span className="font-semibold text-sm">ApprovalFlow</span>
            </div>
          )}
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setCollapsed(!collapsed)}
            className="text-sidebar-foreground hover:bg-sidebar-accent h-8 w-8"
          >
            {collapsed ? <ChevronRight className="h-4 w-4" /> : <ChevronLeft className="h-4 w-4" />}
          </Button>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 py-4 overflow-y-auto">
        <div className="px-3 mb-2">
          {!collapsed && <span className="text-xs font-medium text-sidebar-foreground/60 uppercase tracking-wider">Main</span>}
        </div>
        <ul className="space-y-1 px-2">
          {employeeNav
            .filter((item) => !isAdmin || item.title === 'Dashboard')
            .map((item) => (
              <li key={item.url}>
                <NavLink
                  to={item.url}
                  className={({ isActive }) =>
                    cn(
                      "flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors",
                      isActive
                        ? "bg-sidebar-primary text-sidebar-primary-foreground"
                        : "text-sidebar-foreground/80 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
                      collapsed && "justify-center px-2"
                    )
                  }
                >
                  <item.icon className="h-5 w-5 flex-shrink-0" />
                  {!collapsed && <span>{item.title}</span>}
                </NavLink>
              </li>
            ))}
        </ul>

        {(isManager || isAdmin) && (
          <>
            <div className="px-3 mt-6 mb-2">
              {!collapsed && <span className="text-xs font-medium text-sidebar-foreground/60 uppercase tracking-wider">Approvals</span>}
            </div>
            <ul className="space-y-1 px-2">
              {approverNav.map((item) => (
                <li key={item.url}>
                  <NavLink
                    to={item.url}
                    className={({ isActive }) =>
                      cn(
                        "flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors",
                        isActive
                          ? "bg-sidebar-primary text-sidebar-primary-foreground"
                          : "text-sidebar-foreground/80 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
                        collapsed && "justify-center px-2"
                      )
                    }
                  >
                    <item.icon className="h-5 w-5 flex-shrink-0" />
                    {!collapsed && <span>{item.title}</span>}
                  </NavLink>
                </li>
              ))}
            </ul>
          </>
        )}

        {isAdmin && (
          <>
            <div className="px-3 mt-6 mb-2">
              {!collapsed && <span className="text-xs font-medium text-sidebar-foreground/60 uppercase tracking-wider">Admin</span>}
            </div>
            <ul className="space-y-1 px-2">
              {adminNav.map((item) => (
                <li key={item.url}>
                  <NavLink
                    to={item.url}
                    className={({ isActive }) =>
                      cn(
                        "flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors",
                        isActive
                          ? "bg-sidebar-primary text-sidebar-primary-foreground"
                          : "text-sidebar-foreground/80 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
                        collapsed && "justify-center px-2"
                      )
                    }
                  >
                    <item.icon className="h-5 w-5 flex-shrink-0" />
                    {!collapsed && <span>{item.title}</span>}
                  </NavLink>
                </li>
              ))}
            </ul>
          </>
        )}
      </nav>

      {/* User Profile & Logout */}
      <div className="p-4 border-t border-sidebar-border">
        {!collapsed && (
          <div className="flex items-center gap-3 mb-3">
            <div className="w-9 h-9 rounded-full bg-sidebar-accent flex items-center justify-center">
              <span className="text-sm font-medium">{user.name.charAt(0)}</span>
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium truncate">{user.name}</p>
              <p className="text-xs text-sidebar-foreground/60 capitalize">{user.role}</p>
            </div>
          </div>
        )}
        <Button
          variant="ghost"
          onClick={logout}
          className={cn(
            "w-full text-sidebar-foreground/80 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
            collapsed ? "px-2" : "justify-start"
          )}
        >
          <LogOut className="h-5 w-5" />
          {!collapsed && <span className="ml-3">Logout</span>}
        </Button>
      </div>
    </aside>
  );
}
