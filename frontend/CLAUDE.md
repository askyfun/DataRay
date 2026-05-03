# DataRay Frontend Coding Standards

This document defines coding standards and conventions for AI assistants working on the DataRay frontend codebase.

## Tech Stack

- React 18 + TypeScript
- Ant Design 5.x
- ECharts 5.x
- Zustand 4.x
- @dnd-kit (drag and drop)
- Vite 6
- Biome (linting + formatting)
- Vitest (testing)

## Code Style

### TypeScript

- **Always use strict TypeScript**: Enable `strict: true` in tsconfig
- **Never use `any`**: Use `unknown` instead when type is uncertain
- **Use explicit types**: Prefer explicit type annotations over type inference for function parameters and return values
- **Use snake_case for API**: All API communication uses snake_case (backend convention)

```typescript
// Correct
interface Datasource {
  id: number;
  name: string;
  host: string;
  port: number;
  database_name: string;
  created_at: string;
}

// Wrong
interface Datasource {
  id: any;
  name: string;
}
```

### React Components

- **Use functional components**: Never use class components
- **Use FC with explicit props**: Define prop types explicitly

```typescript
// Correct
interface ButtonProps {
  label: string;
  onClick: () => void;
}

export const Button: React.FC<ButtonProps> = ({ label, onClick }) => {
  return <button onClick={onClick}>{label}</button>
}

// Wrong
export const Button = ({ label, onClick }) => {
  return <button onClick={onClick}>{label}</button>
}
```

### Import Order

1. React/library imports
2. Project internal imports
3. Type imports

```typescript
// 1. React/Library
import React, { useState, useEffect } from 'react'
import { Button, Table, Modal } from 'antd'
import { useNavigate } from 'react-router-dom'

// 2. Project internal
import { datasourcesApi } from '@/api'
import { useStore } from '@/store'

// 3. Types
import type { Datasource, Dataset } from '@/api'
```

### Naming Conventions

- Components: PascalCase (`ChartBuilder.tsx`)
- Functions/variables: camelCase
- Constants: UPPER_SNAKE_CASE
- Files: kebab-case (except components)

### State Management (Zustand)

```typescript
// Correct - Use Zustand with TypeScript
interface StoreState {
  datasources: Datasource[]
  setDatasources: (datasources: Datasource[]) => void
}

export const useStore = create<StoreState>((set) => ({
  datasources: [],
  setDatasources: (datasources) => set({ datasources }),
}))
```

### Ant Design Usage

- Use Ant Design components consistently
- Follow Ant Design patterns for forms and tables
- Use `message` from antd for notifications

### API Calls

- Use the centralized API client in `@/lib/api/client`
- Always handle errors appropriately
- Use snake_case for API request/response

```typescript
// Correct
import { get, post } from '@/lib/api/client'

const result = await get<Datasource[]>('/api/datasources')
if (result.code === 20000) {
  const datasources = result.data
}
```

### Testing (Vitest)

- Write tests for all business logic
- Use @testing-library/react for component tests
- Follow AAA pattern: Arrange, Act, Assert

```typescript
describe('ComponentName', () => {
  it('should render correctly', () => {
    // Arrange
    const props = { /* ... */ }

    // Act
    render(<ComponentName {...props} />)

    // Assert
    expect(screen.getByText('Expected')).toBeInTheDocument()
  })
})
```

### Formatting

- Use Biome for formatting (2 spaces indent, 100 char line width)
- Run `npm run format` before committing
- Run `npm run check` to verify code quality

### Accessibility

- Use semantic HTML elements
- Add aria-label to icon buttons
- Ensure keyboard navigation works

## Project Structure

```
frontend/src/
├── api/              # API clients
├── components/       # Reusable components
├── idls/            # API type definitions
├── lib/             # Utilities
├── pages/           # Page components
├── store/           # Zustand stores
├── styles/          # Global styles
└── __tests__/       # Test files
```

## Commands

| Command | Description |
|---------|-------------|
| `npm run dev` | Start development server |
| `npm run build` | Build for production |
| `npm run format` | Format code with Biome |
| `npm run lint` | Lint with Biome |
| `npm run check` | Run Biome check |
| `npm run test` | Run tests |
| `npm run build:check` | Run full check (lint + test) |

## Pre-commit

Before committing, ensure:
1. `npm run build:check` passes
2. All tests pass
3. No TypeScript errors
