import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AxiosError, AxiosHeaders } from 'axios';
import { AuthProvider, useAuth } from './AuthContext';
import { authService } from '@/services/auth.service';

vi.mock('@/services/auth.service', () => ({
  authService: { getUserInfo: vi.fn() },
}));

const USER_KEY = 'approval_engine_user';
const storedUser = { id: 7, name: 'Asha', email: 'asha@company.com', role: 'employee' };

function Probe() {
  const { user, isLoading } = useAuth();
  if (isLoading) return <p>loading</p>;
  return <p>{user ? `signed in as ${user.email}` : 'signed out'}</p>;
}

function renderProvider() {
  return render(
    <QueryClientProvider client={new QueryClient()}>
      <AuthProvider>
        <Probe />
      </AuthProvider>
    </QueryClientProvider>,
  );
}

function httpError(status: number) {
  return new AxiosError('request failed', String(status), undefined, undefined, {
    status,
    statusText: '',
    data: {},
    headers: {},
    config: { headers: new AxiosHeaders() },
  });
}

describe('AuthProvider session check', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.mocked(authService.getUserInfo).mockReset();
  });

  it('does not call /api/me when this browser never signed in', () => {
    renderProvider();
    expect(screen.getByText('signed out')).toBeInTheDocument();
    expect(authService.getUserInfo).not.toHaveBeenCalled();
  });

  it('restores the session the server confirms', async () => {
    localStorage.setItem(USER_KEY, JSON.stringify(storedUser));
    vi.mocked(authService.getUserInfo).mockResolvedValue({ ...storedUser, role: 'employee' });
    renderProvider();
    expect(screen.getByText('loading')).toBeInTheDocument();
    expect(await screen.findByText('signed in as asha@company.com')).toBeInTheDocument();
  });

  it('forgets the stored user when the session has expired', async () => {
    localStorage.setItem(USER_KEY, JSON.stringify(storedUser));
    vi.mocked(authService.getUserInfo).mockRejectedValue(httpError(401));
    renderProvider();
    expect(await screen.findByText('signed out')).toBeInTheDocument();
    expect(localStorage.getItem(USER_KEY)).toBeNull();
  });

  it('keeps the stored user when the API cannot be reached', async () => {
    localStorage.setItem(USER_KEY, JSON.stringify(storedUser));
    vi.mocked(authService.getUserInfo).mockRejectedValue(new Error('Network Error'));
    renderProvider();
    expect(await screen.findByText('signed in as asha@company.com')).toBeInTheDocument();
    await waitFor(() => expect(localStorage.getItem(USER_KEY)).not.toBeNull());
  });
});
