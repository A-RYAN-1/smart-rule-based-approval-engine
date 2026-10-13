import { useEffect, useState } from 'react';
import { authService, DemoLogin } from '@/services/auth.service';

const roleLabel: Record<string, string> = {
  ADMIN: 'Admin',
  MANAGER: 'Manager',
  EMPLOYEE: 'Employee',
};

interface DemoLoginsProps {
  /** Fills the sign-in form with the chosen account. */
  onPick: (email: string, password: string) => void;
}

/** "Try it live": the public demo accounts, shown only in demo mode. */
export function DemoLogins({ onPick }: DemoLoginsProps) {
  const [logins, setLogins] = useState<DemoLogin[]>([]);

  useEffect(() => {
    let active = true;
    authService.demoLogins().then((list) => {
      if (active) setLogins(list);
    });
    return () => {
      active = false;
    };
  }, []);

  if (logins.length === 0) return null;

  return (
    <section aria-labelledby="demo-logins-title" className="mt-6 rounded-lg border border-border bg-muted/40 p-4">
      <h3 id="demo-logins-title" className="text-sm font-semibold">Try it live</h3>
      <p className="mt-1 text-xs text-muted-foreground">
        Demo accounts with sample data, reset daily. Pick one to fill the form, then sign in.
      </p>
      <ul className="mt-3 space-y-2">
        {logins.map((l) => (
          <li key={l.email}>
            <button
              type="button"
              onClick={() => onPick(l.email, l.password)}
              className="w-full rounded-md border border-border bg-background px-3 py-2 text-left text-xs hover:border-primary focus:outline-none focus-visible:ring-2 focus-visible:ring-primary"
            >
              <span className="font-medium">{roleLabel[l.role] ?? l.role}</span>
              <span className="ml-2 text-muted-foreground">
                {l.email} / {l.password}
              </span>
            </button>
          </li>
        ))}
      </ul>
    </section>
  );
}
