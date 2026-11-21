import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { DecisionNote } from './DecisionNote';

describe('DecisionNote', () => {
  it('shows the heading and the explanation for a rule decision', () => {
    render(<DecisionNote decision={{ kind: 'AUTO_APPROVED', ruleId: 7, limit: 3, text: 'Approved by the system: 2 days is within the limit of 3 days set by rule #7.' }} />);
    expect(screen.getByText('Cleared by a rule')).toBeInTheDocument();
    expect(screen.getByText(/within the limit of 3 days set by rule #7/)).toBeInTheDocument();
  });

  it('says a request went to a person', () => {
    render(<DecisionNote decision={{ kind: 'SENT_TO_PERSON', ruleId: 8, limit: 5000, text: 'Sent to a person: 7500 is over the limit of 5000 set by rule #8.' }} />);
    expect(screen.getByText('Sent to a person')).toBeInTheDocument();
  });

  it('renders nothing without a decision', () => {
    const { container } = render(<DecisionNote />);
    expect(container).toBeEmptyDOMElement();
  });
});
