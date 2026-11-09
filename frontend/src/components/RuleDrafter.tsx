import { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { adminService, RuleDraft } from '@/services/admin.service';
import { apiErrorMessage } from '@/lib/api';
import { RequestType } from '@/types';

const FIELD_LABEL: Record<string, string> = {
  max_days: 'Longest leave (working days)',
  max_amount: 'Largest amount (rupees)',
  max_percent: 'Largest discount (percent)',
};

const MAX_SENTENCE = 300;

interface RuleDrafterProps {
  /** Saves the confirmed rule through the ordinary rules endpoint. */
  onConfirm: (rule: {
    requestType: RequestType;
    condition: Record<string, number>;
    action: 'AUTO_APPROVE';
    gradeId: number;
    isActive: boolean;
  }) => Promise<void>;
}

/**
 * Type a sentence, review the proposed rule as a before/after, then confirm.
 * A draft is never saved by the server; Confirm is the only thing that saves.
 */
export function RuleDrafter({ onConfirm }: RuleDrafterProps) {
  const [enabled, setEnabled] = useState(false);
  const [sentence, setSentence] = useState('');
  const [draft, setDraft] = useState<RuleDraft | null>(null);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    let live = true;
    adminService
      .ruleDraftingEnabled()
      .then((on) => live && setEnabled(on))
      .catch(() => live && setEnabled(false));
    return () => {
      live = false;
    };
  }, []);

  if (!enabled) return null;

  const requestDraft = async () => {
    setBusy(true);
    setError('');
    setDraft(null);
    try {
      setDraft(await adminService.draftRule(sentence));
    } catch (e) {
      setError(apiErrorMessage(e, 'Could not draft a rule.'));
    } finally {
      setBusy(false);
    }
  };

  const confirm = async () => {
    if (!draft) return;
    setBusy(true);
    try {
      await onConfirm({
        requestType: draft.rule.request_type.toLowerCase() as RequestType,
        condition: draft.rule.condition,
        action: 'AUTO_APPROVE',
        gradeId: draft.rule.grade_id,
        isActive: true,
      });
      setDraft(null);
      setSentence('');
    } finally {
      setBusy(false);
    }
  };

  const { change } = draft ?? {};

  return (
    <Card>
      <CardHeader>
        <CardTitle>Draft a rule from a sentence</CardTitle>
        <CardDescription>
          For example: auto-approve expenses under 5000 for grade 2. The draft is checked against the grade
          limits and saved only when you confirm.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <form
          className="flex flex-col gap-2 sm:flex-row sm:items-end"
          onSubmit={(e) => {
            e.preventDefault();
            if (sentence.trim() && !busy) void requestDraft();
          }}
        >
          <div className="flex-1 space-y-2">
            <Label htmlFor="rule-sentence">Describe the rule</Label>
            <Input
              id="rule-sentence"
              value={sentence}
              maxLength={MAX_SENTENCE}
              onChange={(e) => setSentence(e.target.value)}
              placeholder="Auto-approve leave up to 5 days for grade 2"
            />
          </div>
          <Button type="submit" disabled={busy || !sentence.trim()}>
            {busy && !draft ? 'Drafting' : 'Draft rule'}
          </Button>
        </form>

        {error && (
          <p role="alert" className="text-sm text-destructive">
            {error}
          </p>
        )}

        {draft && change && (
          <div className="space-y-3 rounded-md border p-4" aria-live="polite">
            <p className="text-sm font-medium">
              {draft.rule.request_type} rule for {draft.grade_name}
            </p>
            <p className="text-sm text-muted-foreground">{FIELD_LABEL[change.field] ?? change.field}</p>
            <p className="text-lg">
              {change.before === null ? (
                <>
                  <span className="text-muted-foreground">no rule yet</span> to <strong>{change.after}</strong>
                </>
              ) : (
                <>
                  <span className="text-muted-foreground">{change.before}</span> to <strong>{change.after}</strong>
                </>
              )}
            </p>
            <div className="flex gap-2">
              <Button onClick={() => void confirm()} disabled={busy}>
                Confirm and save
              </Button>
              <Button variant="outline" onClick={() => setDraft(null)} disabled={busy}>
                Discard
              </Button>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
