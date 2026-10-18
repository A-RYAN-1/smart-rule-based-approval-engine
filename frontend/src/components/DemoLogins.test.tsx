import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { DemoLogins } from './DemoLogins';
import { authService } from '@/services/auth.service';

vi.mock('@/services/auth.service', () => ({
  authService: { demoLogins: vi.fn() },
}));

describe('DemoLogins', () => {
  beforeEach(() => vi.mocked(authService.demoLogins).mockReset());

  it('holds the panel space while loading, then renders nothing when demo mode is off', async () => {
    vi.mocked(authService.demoLogins).mockResolvedValue([]);
    const { container } = render(<DemoLogins onPick={() => {}} />);
    expect(container.firstElementChild).toHaveAttribute('aria-hidden', 'true');
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
    await waitFor(() => expect(container).toBeEmptyDOMElement());
  });

  it('fills the form with the picked account', async () => {
    vi.mocked(authService.demoLogins).mockResolvedValue([
      { role: 'MANAGER', email: 'manager@company.com', password: 'demo-pass' },
    ]);
    const onPick = vi.fn();
    render(<DemoLogins onPick={onPick} />);
    fireEvent.click(await screen.findByRole('button', { name: /Manager/ }));
    expect(onPick).toHaveBeenCalledWith('manager@company.com', 'demo-pass');
  });
});
