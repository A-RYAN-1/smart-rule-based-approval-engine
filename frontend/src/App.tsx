import { Toaster } from "@/components/ui/toaster";
import { Toaster as Sonner } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { AuthProvider } from "@/contexts/AuthContext";
import { ProtectedRoute } from "@/components/ProtectedRoute";

// Pages
import AuthPage from "./pages/AuthPage";
import Dashboard from "./pages/Dashboard";
import LeaveRequestPage from "./pages/LeaveRequestPage";
import ExpenseRequestPage from "./pages/ExpenseRequestPage";
import DiscountRequestPage from "./pages/DiscountRequestPage";
import MyRequestsPage from "./pages/MyRequestsPage";
import PendingApprovalsPage from "./pages/PendingApprovalsPage";
import RulesManagementPage from "./pages/admin/RulesManagementPage";
import ReportsPage from "./pages/admin/ReportsPage";
import HolidaysPage from "./pages/admin/HolidaysPage";
import UsersPage from "./pages/admin/UsersPage";
import NotFound from "./pages/NotFound";

const queryClient = new QueryClient();

const App = () => (
  <QueryClientProvider client={queryClient}>
    <TooltipProvider>
      <AuthProvider>
        <Toaster />
        <Sonner />
        <BrowserRouter>
          <Routes>
            <Route path="/" element={<Navigate to="/dashboard" replace />} />
            <Route path="/auth" element={<AuthPage />} />
            <Route path="/dashboard" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
            <Route path="/leaves" element={<ProtectedRoute><LeaveRequestPage /></ProtectedRoute>} />
            <Route path="/expenses" element={<ProtectedRoute><ExpenseRequestPage /></ProtectedRoute>} />
            <Route path="/discounts" element={<ProtectedRoute><DiscountRequestPage /></ProtectedRoute>} />
            <Route path="/my-requests" element={<ProtectedRoute><MyRequestsPage /></ProtectedRoute>} />
            <Route path="/pending-approvals" element={<ProtectedRoute><PendingApprovalsPage /></ProtectedRoute>} />
            <Route path="/admin/rules" element={<ProtectedRoute><RulesManagementPage /></ProtectedRoute>} />
            <Route path="/admin/reports" element={<ProtectedRoute><ReportsPage /></ProtectedRoute>} />
            <Route path="/admin/holidays" element={<ProtectedRoute><HolidaysPage /></ProtectedRoute>} />
            <Route path="/admin/users" element={<ProtectedRoute><UsersPage /></ProtectedRoute>} />
            <Route path="*" element={<NotFound />} />
          </Routes>
        </BrowserRouter>
      </AuthProvider>
    </TooltipProvider>
  </QueryClientProvider>
);

export default App;
