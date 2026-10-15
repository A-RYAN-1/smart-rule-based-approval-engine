# Frontend

The React single-page app of the Smart Rule-Based Approval Engine. The project overview, live demo,
architecture and full setup are in the [main README](../README.md); this file covers only working on
the frontend.

## Stack

React 18, TypeScript 5, Vite 5, React Router 6, TanStack Query 5, React Hook Form with Zod, Tailwind
CSS 3 with shadcn/ui components, Recharts, Axios. Tests: Vitest with Testing Library.

## Commands

Node is pinned in the repository's `mise.toml`; run each command as `mise exec -- <command>`, or
directly with Node 22.

| Command | What it does |
|---|---|
| `npm ci` | Install the exact locked dependencies |
| `npm run dev` | Dev server on http://localhost:5173; `/api` is proxied to the Go backend on port 8080 |
| `npm test` | Vitest, once |
| `npm run test:watch` | Vitest in watch mode |
| `npm run lint` | ESLint |
| `npm run build` | Production build into `dist/` |

## Structure

```
src/
  pages/        one component per route: dashboard, leave, expense and discount forms,
                my requests, pending approvals, admin rules, reports, holidays, users
  components/   layout, shared UI (shadcn/ui), request and approval widgets
  services/     typed API calls, one module per feature
  lib/api.ts    the Axios client: same-origin /api, cookie session, error mapping
  hooks/        TanStack Query hooks around the services
  contexts/     auth state (current user, login, logout)
  types/        API types shared by services and pages
  test/         Vitest setup
```

## Environment variables

`VITE_API_URL` (see `.env.example`): empty by default, so the app calls `/api` on its own origin. Set it
only when the API runs on another origin, and add that origin to the backend's `CORS_ORIGINS`.
