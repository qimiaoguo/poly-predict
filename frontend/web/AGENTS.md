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

Frontend uses `apiGet<T>`, `apiPost<T>`, `apiPatch<T>` helpers (in `lib/api/client.ts`) that auto-extract `.data` from the response envelope and attach the Supabase Bearer token. SWR hooks use these fetchers for data loading.

## Important Checks Before Writing Code

- **Before writing API calls**, read `docs/api-reference/user-api.yaml` to verify: field names, query param names (e.g. `search` not `q`), HTTP methods (e.g. `PATCH` not `PUT`)
- **All API response data must use optional chaining or defaults** — e.g. `(user.balance ?? 0).toLocaleString()`, `bet.outcome?.toUpperCase()`. API data may be undefined during loading or have unexpected null fields
- **shadcn Badge uses `<span>` not `<div>`** — never place Badge (or any block-level component) inside `<p>` tags to avoid hydration errors

## Testing

- **All frontend code changes must include corresponding test cases**
- Use Vitest + React Testing Library, test files next to source (`foo.tsx` → `foo.test.tsx`)
- Component tests: render with mock data, verify key elements are present, test user interactions
- Hook/utility tests: test pure logic, edge cases (null/undefined inputs, empty arrays)
- API integration: mock fetch/SWR responses, verify loading/error/success states
- Cover happy path and error/edge cases

## Conventions

- Supabase Auth for authentication
- TypeScript: ESLint + Prettier defaults
