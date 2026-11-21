import { Scale } from 'lucide-react';
import type { Decision } from '@/types';

const HEADINGS: Record<Decision['kind'], string> = {
  AUTO_APPROVED: 'Cleared by a rule',
  SENT_TO_PERSON: 'Sent to a person',
  AUTO_REJECTED: 'Closed by the system',
};

/** Explains why a request has its status, from the decision the server stored at filing time. */
export function DecisionNote({ decision }: { decision?: Decision }) {
  if (!decision) return null;
  return (
    <div className="p-4 rounded-xl bg-muted/30 border border-border/50" data-testid="decision-note">
      <div className="flex items-start gap-3">
        <Scale className="h-5 w-5 text-muted-foreground mt-0.5" aria-hidden="true" />
        <div className="space-y-1">
          <p className="text-xs font-bold uppercase tracking-tight text-muted-foreground">
            {HEADINGS[decision.kind]}
          </p>
          <p className="text-sm text-foreground/80 leading-relaxed">{decision.text}</p>
        </div>
      </div>
    </div>
  );
}
