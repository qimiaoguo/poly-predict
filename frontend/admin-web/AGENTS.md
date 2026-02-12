# Admin Frontend AGENTS.md

## Commands

```bash
bash -c 'cd frontend/admin-web && npm run dev'    # Dev server :3001
bash -c 'cd frontend/admin-web && npm run build'  # Production build
bash -c 'cd frontend/admin-web && npm run lint'   # ESLint
```

## Stack

Next.js 16, React 19, TypeScript strict, Tailwind v4, shadcn/ui, SWR (data fetching), Zustand (state)

## Important Checks Before Writing Code

- **Before writing API calls**, read `docs/api-reference/admin-api.yaml` to verify field names, query params, and HTTP methods
- **All API response data must use optional chaining or defaults** — data may be undefined during loading
- **Never place block-level components (Badge, div) inside `<p>` tags** — causes hydration errors
- **Destructive buttons**: use outline variant with `border-destructive text-destructive` for readability, not `variant="destructive"` which has poor text contrast in dark mode

## Testing

- **All frontend code changes must include corresponding test cases**
- Use Vitest + React Testing Library, test files next to source (`foo.tsx` → `foo.test.tsx`)
- Component tests: render with mock data, verify key elements are present, test user interactions
- Hook/utility tests: test pure logic, edge cases (null/undefined inputs, empty arrays)
- API integration: mock fetch/SWR responses, verify loading/error/success states
- Cover happy path and error/edge cases

## Conventions

- Custom JWT auth (not Supabase)
- TypeScript: ESLint + Prettier defaults
