import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { adminService, DryRunOutcome, RuleDraft, RuleDryRun as DryRunResult } from '@/services/admin.service';
import { apiErrorMessage } from '@/lib/api';
import { MAX_VALUES, parseValues } from '@/lib/dryRunValues';

const UNIT: Record<string, string> = { LEAVE: 'days', EXPENSE: 'rupees', DISCOUNT: 'percent' };

const OUTCOME_TEXT: Record<DryRunOutcome, string> = {
  AUTO_APPROVED: 'Approved by the system',
  NEEDS_APPROVAL: 'Goes to a person',
};

/**
 * Shows what a draft rule would decide for values the admin types and for the newest real
 * requests of that grade, next to the rule saved today. The server saves nothing.
 */
export function RuleDryRun({ rule }: { rule: RuleDraft['rule'] }) {
  const [text, setText] = useState('');
  const [result, setResult] = useState<DryRunResult | null>(null);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const unit = UNIT[rule.request_type] ?? '';

  const run = async () => {
    const values = parseValues(text);
    if (values === null) {
      setError(`Type up to ${MAX_VALUES} positive numbers separated by commas, or leave the box empty.`);
      return;
    }
    setBusy(true);
    setError('');
    try {
      setResult(await adminService.dryRunRule(rule, values));
    } catch (e) {
      setResult(null);
      setError(apiErrorMessage(e, 'Could not test the rule.'));
    } finally {
      setBusy(false);
    }
  };

  const summary = result?.summary;
  return (
    <div className="space-y-3 border-t pt-3">
      <p className="text-sm font-medium">Test before saving</p>
      <div className="flex flex-col gap-2 sm:flex-row sm:items-end">
        <div className="flex-1 space-y-2">
          <Label htmlFor="dry-run-values">Values to try ({unit}, optional)</Label>
          <Input
            id="dry-run-values"
            value={text}
            onChange={(e) => setText(e.target.value)}
            placeholder="for example 2, 4, 6"
          />
        </div>
        <Button variant="secondary" onClick={() => void run()} disabled={busy}>
          {busy ? 'Testing' : 'Test against recent requests'}
        </Button>
      </div>

      {error && (
        <p role="alert" className="text-sm text-destructive">
          {error}
        </p>
      )}

      {result && summary && (
        <div className="space-y-3" aria-live="polite">
          {summary.history_total === 0 ? (
            <p className="text-sm text-muted-foreground">No past requests of this type for this grade yet.</p>
          ) : (
            <p className="text-sm">
              Of the last {summary.history_total} {summary.history_total === 1 ? 'request' : 'requests'}, the rule saved today approves{' '}
              <strong>{summary.current_auto}</strong>; this draft would approve <strong>{summary.draft_auto}</strong>
              {' '}({summary.newly_auto} more, {summary.no_longer_auto} fewer).
            </p>
          )}
          {result.rows.length > 0 && (
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-muted-foreground">
                  <th className="py-1 pr-2 font-normal">Value</th>
                  <th className="py-1 pr-2 font-normal">Saved rule</th>
                  <th className="py-1 font-normal">This draft</th>
                </tr>
              </thead>
              <tbody>
                {result.rows.map((row, i) => (
                  <tr key={`${row.value}-${i}`} className="border-t">
                    <td className="py-1 pr-2">{row.value}</td>
                    <td className="py-1 pr-2">{OUTCOME_TEXT[row.current]}</td>
                    <td className="py-1">{OUTCOME_TEXT[row.draft]}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
          <p className="text-xs text-muted-foreground">Nothing was saved.</p>
        </div>
      )}
    </div>
  );
}
