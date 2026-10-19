import { Toaster } from "@/components/ui/toaster";
import { Toaster as Sonner } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { lazy, Suspense } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { AuthProvider } from "@/contexts/AuthContext";
import { ProtectedRoute } from "@/components/ProtectedRoute";
import { PageLoader } from "@/components/PageLoader";

// The sign-in page is the first screen, so it ships in the main bundle;
// every other page is split into its own chunk and loaded on first visit.
import AuthPage from "./pages/AuthPage";
const Dashboard = lazy(() => import("./pages/Dashboard"));
const LeaveRequestPage = lazy(() => import("./pages/LeaveRequestPage"));
const ExpenseRequestPage = lazy(() => import("./pages/ExpenseRequestPage"));
const DiscountRequestPage = lazy(() => import("./pages/DiscountRequestPage"));
const MyRequestsPage = lazy(() => import("./pages/MyRequestsPage"));
const PendingApprovalsPage = lazy(() => import("./pages/PendingApprovalsPage"));
const RulesManagementPage = lazy(() => import("./pages/admin/RulesManagementPage"));
const ReportsPage = lazy(() => import("./pages/admin/ReportsPage"));
const HolidaysPage = lazy(() => import("./pages/admin/HolidaysPage"));
const UsersPage = lazy(() => import("./pages/admin/UsersPage"));
const NotFound = lazy(() => import("./pages/NotFound"));

const queryClient = new QueryClient();

const App = () => (
  <QueryClientProvider client={queryClient}>
    <TooltipProvider>
      <AuthProvider>
        <Toaster />
        <Sonner />
        {/* startTransition keeps the current page on screen while the next page's chunk loads. */}
        <BrowserRouter future={{ v7_startTransition: true }}>
          <Suspense fallback={<PageLoader />}>
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
          </Suspense>
        </BrowserRouter>
      </AuthProvider>
    </TooltipProvider>
  </QueryClientProvider>
);

export default App;
