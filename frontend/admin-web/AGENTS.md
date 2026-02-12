# Admin Frontend AGENTS.md

## Commands

```bash
bash -c 'cd frontend/admin-web && npm run dev'    # Dev server :3001
bash -c 'cd frontend/admin-web && npm run build'  # Production build
bash -c 'cd frontend/admin-web && npm run lint'   # ESLint
```

## Stack

Next.js 16, React 19, TypeScript strict, Tailwind v4, shadcn/ui, SWR (data fetching), Zustand (state)

## Conventions

- Custom JWT auth (not Supabase)
- TypeScript: ESLint + Prettier defaults
