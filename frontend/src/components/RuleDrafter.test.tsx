import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { RuleDrafter } from './RuleDrafter';
import { adminService } from '@/services/admin.service';

vi.mock('@/services/admin.service', () => ({
  adminService: { ruleDraftingEnabled: vi.fn(), draftRule: vi.fn() },
}));

const draft = {
  rule: { request_type: 'EXPENSE', grade_id: 2, condition: { max_amount: 8000 }, action: 'AUTO_APPROVE' },
  grade_name: 'Grade 2',
  change: { field: 'max_amount', before: 5000, after: 8000 },
  prompt_version: 'rule-draft-v1',
};

describe('RuleDrafter', () => {
  beforeEach(() => {
    vi.mocked(adminService.ruleDraftingEnabled).mockReset();
    vi.mocked(adminService.draftRule).mockReset();
  });

  it('renders nothing when drafting is not configured', async () => {
    vi.mocked(adminService.ruleDraftingEnabled).mockResolvedValue(false);
    const { container } = render(<RuleDrafter onConfirm={vi.fn()} />);
    await waitFor(() => expect(adminService.ruleDraftingEnabled).toHaveBeenCalled());
    expect(container).toBeEmptyDOMElement();
  });

  it('shows the before and after, and saves only on confirm', async () => {
    vi.mocked(adminService.ruleDraftingEnabled).mockResolvedValue(true);
    vi.mocked(adminService.draftRule).mockResolvedValue(draft);
    const onConfirm = vi.fn().mockResolvedValue(undefined);
    render(<RuleDrafter onConfirm={onConfirm} />);

    fireEvent.change(await screen.findByLabelText('Describe the rule'), {
      target: { value: 'expenses up to 8k for grade 2' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Draft rule' }));

    expect(await screen.findByText('EXPENSE rule for Grade 2')).toBeInTheDocument();
    expect(screen.getByText('5000')).toBeInTheDocument();
    expect(screen.getByText('8000')).toBeInTheDocument();
    expect(onConfirm).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole('button', { name: 'Confirm and save' }));
    await waitFor(() =>
      expect(onConfirm).toHaveBeenCalledWith({
        requestType: 'expense',
        condition: { max_amount: 8000 },
        action: 'AUTO_APPROVE',
        gradeId: 2,
        isActive: true,
      }),
    );
  });

  it('discarding a draft saves nothing', async () => {
    vi.mocked(adminService.ruleDraftingEnabled).mockResolvedValue(true);
    vi.mocked(adminService.draftRule).mockResolvedValue(draft);
    const onConfirm = vi.fn();
    render(<RuleDrafter onConfirm={onConfirm} />);
    fireEvent.change(await screen.findByLabelText('Describe the rule'), { target: { value: 'x' } });
    fireEvent.click(screen.getByRole('button', { name: 'Draft rule' }));
    fireEvent.click(await screen.findByRole('button', { name: 'Discard' }));
    expect(screen.queryByText('EXPENSE rule for Grade 2')).not.toBeInTheDocument();
    expect(onConfirm).not.toHaveBeenCalled();
  });

  it('shows the server message when the draft fails', async () => {
    vi.mocked(adminService.ruleDraftingEnabled).mockResolvedValue(true);
    vi.mocked(adminService.draftRule).mockRejectedValue(new Error('boom'));
    render(<RuleDrafter onConfirm={vi.fn()} />);
    fireEvent.change(await screen.findByLabelText('Describe the rule'), { target: { value: 'x' } });
    fireEvent.click(screen.getByRole('button', { name: 'Draft rule' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('Could not draft a rule.');
  });
});
