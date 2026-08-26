# aDex-UI Development Guide

This guide covers the development workflow, architecture, and best practices for contributing to aDex-UI.

## Table of Contents

- [Project Structure](#project-structure)
- [Development Environment Setup](#development-environment-setup)
- [Building and Running](#building-and-running)
- [Testing](#testing)
- [Code Organization](#code-organization)
- [API Integration](#api-integration)
- [Frontend Development](#frontend-development)
- [Backend Development](#backend-development)
- [Contributing Guidelines](#contributing-guidelines)
- [Performance Considerations](#performance-considerations)
- [Debugging](#debugging)

## Project Structure

```
aDex-UI/
├── app.go                    # Wails application entry point
├── main.go                   # Main application file
├── go.mod                    # Go module dependencies
├── go.sum                    # Go dependency checksums
├── Taskfile.yml              # Task runner configuration
├── internal/                 # Go backend code
│   ├── models/              # Data models and structs
│   ├── services/            # Business logic services
│   ├── logger/              # Logging infrastructure
│   └── events/              # Event system
├── frontend/                 # Nuxt 4 frontend
│   ├── app/                 # Nuxt 4 app directory
│   │   ├── assets/          # Static assets
│   │   ├── components/      # Vue components
│   │   ├── composables/     # Vue composables
│   │   ├── layouts/         # Nuxt layouts
│   │   ├── pages/           # Nuxt pages
│   │   ├── plugins/         # Nuxt plugins
│   │   ├── stores/          # Pinia stores
│   │   └── types/           # TypeScript types
│   ├── nuxt.config.ts       # Nuxt configuration
│   └── package.json         # Node.js dependencies
├── build/                    # Build configuration
├── tests/                    # Test files
└── docs/                     # Documentation
```

## Development Environment Setup

### Prerequisites

- **Go 1.21+**: Backend development language
- **Node.js 18+**: Frontend development runtime
- **Bun**: Package manager (recommended) or npm/yarn
- **Task**: Task runner (install via `go install github.com/go-task/task/v3/cmd/task@latest`)
- **Git**: Version control

### Environment Setup

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd aDex-UI
   ```

2. **Install Go dependencies**:
   ```bash
   go mod download
   ```

3. **Install Node.js dependencies**:
   ```bash
   cd frontend
   bun install
   cd ..
   ```

4. **Install Task runner** (if not already installed):
   ```bash
   go install github.com/go-task/task/v3/cmd/task@latest
   ```

### IDE Configuration

#### VSCode Extensions
- **Go**: Go language support
- **Vue - Official**: Vue 3 support
- **TypeScript Vue**: TypeScript support for Vue
- **Volar**: Vue 3 IDE support
- **Tailwind CSS**: CSS framework support
- **ESLint**: Code linting
- **Prettier**: Code formatting

#### VSCode Settings (`.vscode/settings.json`)
```json
{
  "go.useLanguageServer": true,
  "go.formatTool": "goimports",
  "go.lintTool": "golangci-lint",
  "typescript.preferences.importModuleSpecifier": "relative",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": true
  }
}
```

## Building and Running

### Development Mode

1. **Start the development server**:
   ```bash
   task dev
   ```

   This starts both the Go backend and Nuxt frontend in development mode with hot reload.

2. **Start only the backend**:
   ```bash
   task dev:backend
   ```

3. **Start only the frontend**:
   ```bash
   task dev:frontend
   ```

### Building for Production

1. **Build for current platform**:
   ```bash
   task build
   ```

2. **Build for all platforms**:
   ```bash
   task build:all
   ```

3. **Build specific platform**:
   ```bash
   task build:darwin
   task build:windows
   task build:linux
   ```

### Running Tests

1. **Run all tests**:
   ```bash
   task test
   ```

2. **Run backend tests**:
   ```bash
   task test:backend
   ```

3. **Run frontend tests**:
   ```bash
   task test:frontend
   ```

4. **Run E2E tests**:
   ```bash
   task test:e2e
   ```

## Code Organization

### Backend (Go)

#### Models (`internal/models/`)
- Define data structures and business entities
- Include validation tags and JSON serialization
- Follow Go naming conventions

Example:
```go
type User struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

#### Services (`internal/services/`)
- Implement business logic
- Handle data processing and external integrations
- Follow dependency injection patterns

Example:
```go
type UserService struct {
    repo UserRepository
    logger *logger.Logger
}

func (s *UserService) CreateUser(user *User) error {
    if err := user.Validate(); err != nil {
        return err
    }
    return s.repo.Create(user)
}
```

#### Events (`internal/events/`)
- Define event contracts for frontend-backend communication
- Handle event routing and processing
- Use typed event payloads

Example:
```go
type Event struct {
    Type    string      `json:"type"`
    Payload interface{} `json:"payload"`
    Timestamp time.Time  `json:"timestamp"`
}
```

### Frontend (Nuxt 4)

#### Components (`frontend/app/components/`)
- Use Vue 3 Composition API
- Follow Single File Component structure
- Include TypeScript types

Example:
```vue
<template>
  <div class="component">
    <h1>{{ title }}</h1>
    <button @click="handleClick">Click me</button>
  </div>
</template>

<script setup lang="ts">
interface Props {
  title: string
}

const props = defineProps<Props>()
const emit = defineEmits<{
  click: []
}>()

const handleClick = () => {
  emit('click')
}
</script>

<style scoped>
.component {
  @apply p-4 bg-white rounded-lg shadow;
}
</style>
```

#### Composables (`frontend/app/composables/`)
- Reusable composition functions
- Use Vue 3 reactivity system
- Include TypeScript types

Example:
```typescript
import { ref, computed } from 'vue'

export const useCounter = (initial: number = 0) => {
  const count = ref(initial)

  const increment = () => count.value++
  const decrement = () => count.value--
  const isEven = computed(() => count.value % 2 === 0)

  return {
    count: readonly(count),
    increment,
    decrement,
    isEven
  }
}
```

#### Stores (`frontend/app/stores/`)
- Use Pinia for state management
- Follow store composition patterns
- Include TypeScript types

Example:
```typescript
import { defineStore } from 'pinia'

export const useUserStore = defineStore('user', () => {
  const user = ref<User | null>(null)

  const setUser = (newUser: User) => {
    user.value = newUser
  }

  const clearUser = () => {
    user.value = null
  }

  return {
    user: readonly(user),
    setUser,
    clearUser
  }
})
```

## API Integration

### Wails Bindings

1. **Backend function registration**:
```go
package main

import (
    "github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) GetUser(id string) (*User, error) {
    return a.userService.GetUser(id)
}
```

2. **Frontend usage**:
```typescript
import { useWails } from '~/composables/useWails'

export const useUser = () => {
  const wails = useWails()

  const getUser = async (id: string): Promise<User | null> => {
    try {
      return await wails.call('App.GetUser', id)
    } catch (error) {
      console.error('Failed to get user:', error)
      return null
    }
  }

  return { getUser }
}
```

### Error Handling

1. **Backend error handling**:
```go
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
```

2. **Frontend error handling**:
```typescript
try {
  const result = await wails.call('SomeOperation')
  return result
} catch (error) {
  console.error('Operation failed:', error)
  // Handle user-friendly error messages
  throw new Error('Operation failed. Please try again.')
}
```

## Frontend Development

### Nuxt 4 Structure

1. **App configuration** (`nuxt.config.ts`):
```typescript
export default defineNuxtConfig({
  app: {
    head: {
      title: 'aDex-UI',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' }
      ]
    }
  },
  modules: [
    '@nuxt/ui',
    '@vueuse/nuxt',
    '@pinia/nuxt'
  ],
  css: ['~/assets/css/main.css'],
  runtimeConfig: {
    public: {
      apiBase: '/api'
    }
  }
})
```

### Styling

1. **Tailwind CSS configuration**:
```typescript
// tailwind.config.ts
export default {
  content: [
    "./components/**/*.{js,vue,ts}",
    "./layouts/**/*.vue",
    "./pages/**/*.vue",
    "./plugins/**/*.{js,ts}",
    "./app.vue"
  ],
  theme: {
    extend: {
      colors: {
        primary: '#1976d2',
        secondary: '#424242'
      }
    }
  }
}
```

2. **CSS variables** (`assets/css/main.css`):
```css
:root {
  --primary-color: #1976d2;
  --secondary-color: #424242;
  --background-color: #121212;
  --text-color: #ffffff;
}

body {
  background-color: var(--background-color);
  color: var(--text-color);
}
```

### Component Development

1. **Component structure**:
```vue
<template>
  <div class="my-component" :class="{ 'is-active': isActive }">
    <slot name="header" />
    <div class="content">
      <slot />
    </div>
    <slot name="footer" />
  </div>
</template>

<script setup lang="ts">
interface Props {
  isActive?: boolean
  title?: string
}

const props = withDefaults(defineProps<Props>(), {
  isActive: false,
  title: 'Default Title'
})

const emit = defineEmits<{
  'update:isActive': [value: boolean]
  'title-changed': [title: string]
}>()
</script>

<style scoped>
.my-component {
  @apply p-4 border rounded-lg;
}

.my-component.is-active {
  @apply border-primary bg-primary/10;
}
</style>
```

## Backend Development

### Service Architecture

1. **Service interface**:
```go
type TerminalService interface {
    CreateSession(config *SessionConfig) (*Session, error)
    ExecuteCommand(sessionID, command string) (*CommandResult, error)
    CloseSession(sessionID string) error
}
```

2. **Service implementation**:
```go
type terminalService struct {
    sessions map[string]*Session
    logger   *logger.Logger
    events   *events.Bus
}

func NewTerminalService(logger *logger.Logger, events *events.Bus) TerminalService {
    return &terminalService{
        sessions: make(map[string]*Session),
        logger:   logger,
        events:   events,
    }
}

func (s *terminalService) CreateSession(config *SessionConfig) (*Session, error) {
    session := &Session{
        ID:        generateID(),
        Config:    config,
        CreatedAt: time.Now(),
    }

    s.sessions[session.ID] = session
    s.events.Publish("session:created", session)

    return session, nil
}
```

### Database Integration

1. **Repository pattern**:
```go
type UserRepository interface {
    Create(user *User) error
    GetByID(id string) (*User, error)
    Update(user *User) error
    Delete(id string) error
}

type userRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(user *User) error {
    query := `INSERT INTO users (id, name, email) VALUES (?, ?, ?)`
    _, err := r.db.Exec(query, user.ID, user.Name, user.Email)
    return err
}
```

## Testing

### Backend Testing

1. **Unit tests**:
```go
func TestTerminalService_CreateSession(t *testing.T) {
    logger := logger.NewNoOp()
    events := events.NewBus()
    service := NewTerminalService(logger, events)

    config := &SessionConfig{
        Shell: "/bin/bash",
        Cols:  80,
        Rows:  24,
    }

    session, err := service.CreateSession(config)

    assert.NoError(t, err)
    assert.NotEmpty(t, session.ID)
    assert.Equal(t, config.Shell, session.Config.Shell)
}
```

2. **Integration tests**:
```go
func TestTerminalIntegration(t *testing.T) {
    app := setupTestApp(t)
    defer app.Cleanup()

    session, err := app.TerminalService.CreateSession(&SessionConfig{
        Shell: "/bin/bash",
        Cols:  80,
        Rows:  24,
    })

    require.NoError(t, err)

    result, err := app.TerminalService.ExecuteCommand(session.ID, "echo 'test'")

    assert.NoError(t, err)
    assert.Equal(t, "test\n", result.Output)
}
```

### Frontend Testing

1. **Component tests**:
```typescript
import { mount } from '@vue/test-utils'
import { describe, it, expect } from 'vitest'
import MyComponent from '~/components/MyComponent.vue'

describe('MyComponent', () => {
  it('renders correctly', () => {
    const wrapper = mount(MyComponent, {
      props: {
        title: 'Test Title'
      }
    })

    expect(wrapper.text()).toContain('Test Title')
  })

  it('emits events correctly', async () => {
    const wrapper = mount(MyComponent)

    await wrapper.find('button').trigger('click')

    expect(wrapper.emitted('click')).toBeTruthy()
  })
})
```

2. **Composable tests**:
```typescript
import { describe, it, expect } from 'vitest'
import { useCounter } from '~/composables/useCounter'

describe('useCounter', () => {
  it('initializes with default value', () => {
    const { count } = useCounter()

    expect(count.value).toBe(0)
  })

  it('increments correctly', () => {
    const { count, increment } = useCounter()

    increment()

    expect(count.value).toBe(1)
  })
})
```

## Contributing Guidelines

### Code Style

1. **Go**: Follow `gofmt` and `golangci-lint` standards
2. **TypeScript**: Use ESLint and Prettier configurations
3. **Vue**: Use Vue 3 Composition API with `<script setup>`
4. **CSS**: Use Tailwind CSS utility classes

### Commit Messages

Follow conventional commits format:

```
type(scope): description

feat(terminal): add session management
fix(audio): resolve volume control issue
docs(readme): update installation instructions
```

### Pull Request Process

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make changes and add tests
4. Run tests: `task test`
5. Submit pull request with description

### Review Guidelines

- Code should be well-documented
- Tests should cover new functionality
- No breaking changes without discussion
- Follow existing patterns and conventions

## Performance Considerations

### Frontend

1. **Lazy loading**: Use `defineAsyncComponent` for heavy components
2. **Virtual scrolling**: For large lists
3. **Debouncing**: For input events and API calls
4. **Memoization**: Use `computed` for expensive calculations

### Backend

1. **Connection pooling**: For database connections
2. **Goroutine management**: Avoid leaks with proper cleanup
3. **Memory management**: Profile for memory leaks
4. **Caching**: Cache frequently accessed data

## Debugging

### Frontend

1. **Vue DevTools**: Browser extension for Vue debugging
2. **Console logging**: Use appropriate log levels
3. **Network tab**: Monitor API calls
4. **Performance profiling**: Use browser dev tools

### Backend

1. **Delve**: Go debugger
2. **Logging**: Use structured logging with context
3. **pprof**: Go profiling tools
4. **Memory profiling**: Check for leaks

### Common Issues

1. **Wails binding issues**: Check function signatures and JSON tags
2. **Hot reload problems**: Restart development server
3. **Type errors**: Run TypeScript compiler
4. **Import issues**: Check file paths and exports

## Wails v2 Workflow

This project targets **Wails v2.12.0** (`github.com/wailsapp/wails/v2`),
not v3. Use the v2 CLI:

```bash
# Install the CLI (one-time)
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0

# Live-reload dev (Go backend + Vite-proxied Nuxt frontend)
wails dev

# Production build for the current platform
wails build
```

`wails.json` (`frontend:dev:serverUrl`, `assetdir`, `wailsjsdir`) wires
the Nuxt build into the Wails embed. The frontend is served from
`frontend/.output/public` at build time and proxied from the Nuxt dev
server during `wails dev`.

### Binding regeneration

Bound Go methods (everything uppercase-leading on `App` and
`ServiceCoordinator`) are exposed to the frontend. After adding or
changing a bound method:

```bash
wails generate module
```

This regenerates `frontend/app/lib/wailsjs/`. A hand-written shim at
`frontend/app/lib/wailsjs/coordinator.ts` keeps the canonical import
surface stable between regens — keep its TypeScript signatures in sync
when you add coordinator methods so the frontend type-checks before the
next regen.

Gotchas (learned the hard way — see
`openspec/changes/edex-parity-and-uplift/tasks.md` "Launch crash fix"):

- Wails JSON-marshals every bound method's return value. Returning a
  live Go struct that holds channels / `sync` primitives /
  `context.CancelFunc` triggers a launch fatal
  (`json: unsupported type: func() error`). Add `json:"-"` tags to
  those fields or return a plain DTO.
- Unbind internal helpers by lowercasing the leading character
  (`getServiceInternal`, not `GetService`).
- Field names need explicit `json:"name"` tags or the frontend reads
  `undefined` (Go defaults to PascalCase, the TS side expects
  camelCase).

### Quit / shutdown

`runtime.Quit(ctx)` triggers teardown → `OnBeforeClose` →
`OnShutdown` → process exit. Call it **exactly once**: on Linux/GTK a
double call hits `gtk_main_quit: assertion 'main_loops != NULL'`. The
frontend's `doQuit()` has a single-fire latch for this reason. Do NOT
run `coordinator.Shutdown()` in `QuitApp` — `OnShutdown` already does
it on the correct thread; running it in parallel deadlocks `sc.mu`.

## Evidence Capture Protocol

Several spec items require visual evidence captured from a running
`wails dev` (screenshots/recordings). The agent workflow can't run
the GUI, so these are captured by a human and dropped into
`docs/evidence/<feature>/`.

Layout: `docs/evidence/<feature-id>/<theme-or-variant>.png`
(e.g. `docs/evidence/theme-engine/tron.png`).

Reference visual targets (the original eDEX-UI look) are noted in the
agent memory and the four reference screenshots
(`screenshot_default/disrupted/blade/horizon.png`). When capturing,
match the corresponding aDex layout to its reference and flag
divergences for iteration.

Features needing evidence: `desktop-shell-layout`, `theme-engine`,
`terminal`, `system-monitor`, `network-monitor`,
`globe-visualization`, `keyboard-layout-pack`, `settings-modal`,
`boot-sequence`. See section 8 of the OpenSpec tasks doc for the
exact capture list per feature.

## Additional Resources

- [Wails v2 Documentation](https://wails.io/docs/introduction)
- [Nuxt 4 Documentation](https://nuxt.com/)
- [Vue 3 Documentation](https://vuejs.org/)
- [Go Documentation](https://golang.org/)
- [Tailwind CSS](https://tailwindcss.com/)
- Project docs: [themes.md](./themes.md), [keyboards.md](./keyboards.md),
  [audio.md](./audio.md), [troubleshooting.md](./troubleshooting.md),
  [cross-platform-packaging.md](./cross-platform-packaging.md)

## Getting Help

1. Check existing documentation and issues
2. Search forums and Stack Overflow
3. Ask questions in appropriate channels
4. Create detailed bug reports with reproduction steps

---

Happy coding! 🚀