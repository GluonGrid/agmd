# Test Two Lists

## Rules

### typescript

# TypeScript Rules

- Always use strict typing
- Avoid using `any` type
- Enable `strictNullChecks` in tsconfig.json
- Prefer interfaces over type aliases for object shapes

### eslint

# ESLint Rules

- Run ESLint on all JavaScript/TypeScript files
- Use recommended rules from @typescript-eslint
- Fix all warnings before committing
- Configure pre-commit hooks

### security

# Security Rules

- Never commit secrets or API keys to repository
- Always sanitize user input
- Use parameterized queries to prevent SQL injection
- Implement proper CORS policies
- Keep dependencies up to date
- Use HTTPS for all external communication
- Implement rate limiting on public endpoints

### no-console

# No Console Statements

- Never use `console.log` in production code
- Use proper logging library (e.g., winston, pino)
- Remove debugging console statements before committing

