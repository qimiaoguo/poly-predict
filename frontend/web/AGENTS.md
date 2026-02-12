# Web Frontend AGENTS.md

## Commands

```bash
bash -c 'cd frontend/web && npm run dev'    # Dev server :3000
bash -c 'cd frontend/web && npm run build'  # Production build
bash -c 'cd frontend/web && npm run lint'   # ESLint
```

## Stack

Next.js 16, React 19, TypeScript strict, Tailwind v4, shadcn/ui, SWR (data fetching), Zustand (state), Recharts (charts), Zod (validation), next-themes (dark/light mode)

## API Client

Frontend uses `apiGet<T>`, `apiPost<T>`, `apiPut<T>` helpers (in `lib/api/client.ts`) that auto-extract `.data` from the response envelope and attach the Supabase Bearer token. SWR hooks use these fetchers for data loading.

## Conventions

- Supabase Auth for authentication
- TypeScript: ESLint + Prettier defaults
