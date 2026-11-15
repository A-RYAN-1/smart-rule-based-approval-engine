import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { RuleDryRun } from './RuleDryRun';
import { parseValues } from '@/lib/dryRunValues';
import { adminService } from '@/services/admin.service';

vi.mock('@/services/admin.service', () => ({
  adminService: { dryRunRule: vi.fn() },
}));

const rule = { request_type: 'EXPENSE', grade_id: 2, condition: { max_amount: 8000 }, action: 'AUTO_APPROVE' };

const result = {
  current_condition: { max_amount: 5000 },
  draft_condition: { max_amount: 8000 },
  rows: [
    { value: 6000, source: 'sample' as const, current: 'NEEDS_APPROVAL' as const, draft: 'AUTO_APPROVED' as const },
  ],
  summary: { history_total: 10, current_auto: 4, draft_auto: 7, newly_auto: 3, no_longer_auto: 0 },
};

describe('parseValues', () => {
  it('reads commas and spaces, and refuses anything that is not a positive number', () => {
    expect(parseValues('2, 4 6')).toEqual([2, 4, 6]);
    expect(parseValues('')).toEqual([]);
    expect(parseValues('0')).toBeNull();
    expect(parseValues('-3')).toBeNull();
    expect(parseValues('abc')).toBeNull();
    expect(parseValues(Array.from({ length: 21 }, () => '1').join(','))).toBeNull();
  });
});

describe('RuleDryRun', () => {
  beforeEach(() => vi.mocked(adminService.dryRunRule).mockReset());

  it('sends the draft and the typed values, then shows the counts and the sample row', async () => {
    vi.mocked(adminService.dryRunRule).mockResolvedValue(result);
    render(<RuleDryRun rule={rule} />);

    fireEvent.change(screen.getByLabelText(/Values to try/), { target: { value: '6000' } });
    fireEvent.click(screen.getByRole('button', { name: 'Test against recent requests' }));

    await waitFor(() => expect(adminService.dryRunRule).toHaveBeenCalledWith(rule, [6000]));
    expect(await screen.findByText(/Of the last 10 requests/)).toBeInTheDocument();
    expect(screen.getByText('Goes to a person')).toBeInTheDocument();
    expect(screen.getByText('Approved by the system')).toBeInTheDocument();
    expect(screen.getByText('Nothing was saved.')).toBeInTheDocument();
  });

  it('does not call the server for a bad value', async () => {
    render(<RuleDryRun rule={rule} />);
    fireEvent.change(screen.getByLabelText(/Values to try/), { target: { value: 'abc' } });
    fireEvent.click(screen.getByRole('button', { name: 'Test against recent requests' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('positive numbers');
    expect(adminService.dryRunRule).not.toHaveBeenCalled();
  });

  it('says so when there are no past requests', async () => {
    vi.mocked(adminService.dryRunRule).mockResolvedValue({
      ...result,
      rows: [],
      summary: { history_total: 0, current_auto: 0, draft_auto: 0, newly_auto: 0, no_longer_auto: 0 },
    });
    render(<RuleDryRun rule={rule} />);
    fireEvent.click(screen.getByRole('button', { name: 'Test against recent requests' }));
    expect(await screen.findByText(/No past requests of this type/)).toBeInTheDocument();
  });
});
