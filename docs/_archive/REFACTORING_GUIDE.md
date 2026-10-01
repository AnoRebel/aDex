# Refactoring Guide for aDex-UI

## Overview

This guide provides comprehensive refactoring guidelines for the aDex-UI project to maintain code quality, performance, and maintainability.

## Principles

### 1. SOLID Principles
- **S**ingle Responsibility: Each component/service should have one reason to change
- **O**pen/Closed: Open for extension, closed for modification
- **L**iskov Substitution: Subtypes must be substitutable for base types
- **I**nterface Segregation: Many specific interfaces are better than one general interface
- **D**ependency Inversion: Depend on abstractions, not concretions

### 2. Clean Code Principles
- Meaningful names
- Small functions and methods
- Comments should explain why, not what
- DRY (Don't Repeat Yourself)
- KISS (Keep It Simple, Stupid)

## Go Backend Refactoring

### Service Layer

#### Before (Problematic)
```go
type TerminalService struct {
    // Large service handling multiple concerns
    sessions map[string]*Session
    config   Config
    logger   *logger.Logger
    // Also handling system monitoring
    systemMetrics *SystemMetrics
    // Also handling file operations
    fileManager *FileManager
}

func (s *TerminalService) ExecuteCommand(sessionID, cmd string) (string, error) {
    // Terminal logic
    // System monitoring logic
    // File operation logic
    // Error handling logic
    // All mixed together
}
```

#### After (Refactored)
```go
type TerminalService struct {
    sessions map[string]*Session
    config   Config
    logger   *logger.Logger
    systemMonitor system.Monitor  // Dependency injection
    fileOps      file.Operator   // Dependency injection
}

// Separate concerns into focused methods
func (s *TerminalService) ExecuteCommand(sessionID, cmd string) (string, error) {
    return s.executeTerminalCommand(sessionID, cmd)
}

func (s *TerminalService) executeTerminalCommand(sessionID, cmd string) (string, error) {
    // Only terminal logic here
}

func (s *TerminalService) monitorResourceUsage() {
    // Delegate to system monitor
}
```

### Interface Segregation

#### Before
```go
type FileSystem interface {
    ReadFile(path string) ([]byte, error)
    WriteFile(path string, data []byte) error
    CreateDir(path string) error
    DeleteFile(path string) error
    ListFiles(path string) ([]string, error)
    GetFileInfo(path string) (*FileInfo, error)
    WatchFile(path string) (Watcher, error)
}
```

#### After
```go
type FileReader interface {
    ReadFile(path string) ([]byte, error)
}

type FileWriter interface {
    WriteFile(path string, data []byte) error
}

type DirectoryManager interface {
    CreateDir(path string) error
    ListFiles(path string) ([]string, error)
}

type FileDeleter interface {
    DeleteFile(path string) error
}

type FileInfoProvider interface {
    GetFileInfo(path string) (*FileInfo, error)
}
```

### Error Handling

#### Before
```go
func (s *Service) ProcessData(data string) error {
    if data == "" {
        return errors.New("data is empty")
    }

    file, err := os.Open("config.txt")
    if err != nil {
        return err
    }

    // Process data
    return nil
}
```

#### After
```go
func (s *Service) ProcessData(data string) error {
    if data == "" {
        return NewValidationError("data cannot be empty")
    }

    file, err := s.configFileOpener.Open("config.txt")
    if err != nil {
        return NewConfigurationError("failed to open config", err)
    }
    defer file.Close()

    return s.processDataWithConfig(data, file)
}

// Custom error types with context
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}
```

## Frontend Refactoring

### Component Structure

#### Before (Large Component)
```vue
<template>
  <div class="terminal-manager">
    <!-- Terminal tabs -->
    <div class="tabs">
      <div v-for="session in sessions" :key="session.id">
        <!-- Complex tab UI -->
      </div>
    </div>

    <!-- Terminal content -->
    <div class="terminal">
      <!-- Complex terminal UI -->
    </div>

    <!-- Settings panel -->
    <div class="settings">
      <!-- Complex settings UI -->
    </div>

    <!-- File browser -->
    <div class="file-browser">
      <!-- Complex file browser UI -->
    </div>
  </div>
</template>

<script setup lang="ts">
// All logic mixed together (300+ lines)
const sessions = ref([])
const settings = ref({})
const files = ref([])

// Terminal logic, settings logic, file browser logic all mixed
</script>
```

#### After (Separated Components)
```vue
<!-- TerminalManager.vue -->
<template>
  <div class="terminal-manager">
    <TerminalTabs
      :sessions="sessions"
      @session-select="handleSessionSelect"
    />
    <TerminalContent
      :session="activeSession"
      @command-execute="handleCommandExecute"
    />
    <TerminalSettings v-model="settings" />
    <FileBrowser @file-select="handleFileSelect" />
  </div>
</template>

<script setup lang="ts">
// Only orchestrator logic (50-80 lines)
const sessions = useTerminalSessions()
const activeSession = computed(() => sessions.value.find(s => s.active))
const settings = useTerminalSettings()

const handleSessionSelect = (sessionId: string) => {
  // Delegated to composables
  useTerminalSessionActions().selectSession(sessionId)
}
</script>
```

### Composable Pattern

#### Before (Mixed Logic in Component)
```vue
<script setup lang="ts">
const sessions = ref([])
const activeSession = ref(null)

const createSession = async () => {
  // Session creation logic
}

const executeCommand = async (command: string) => {
  // Command execution logic
}

const closeSession = async (sessionId: string) => {
  // Session closing logic
}

// Settings logic mixed in
const settings = ref({})
const updateSettings = async (newSettings: any) => {
  // Settings update logic
}
</script>
```

#### After (Composable Separation)
```typescript
// composables/useTerminalSessions.ts
export function useTerminalSessions() {
  const sessions = ref<TerminalSession[]>([])
  const activeSession = ref<TerminalSession | null>(null)

  const createSession = async () => {
    const session = await terminalService.createSession()
    sessions.value.push(session)
    return session
  }

  const executeCommand = async (command: string) => {
    if (!activeSession.value) return
    return await activeSession.value.execute(command)
  }

  const closeSession = async (sessionId: string) => {
    await terminalService.closeSession(sessionId)
    sessions.value = sessions.value.filter(s => s.id !== sessionId)
  }

  return {
    sessions: readonly(sessions),
    activeSession: readonly(activeSession),
    createSession,
    executeCommand,
    closeSession
  }
}

// composables/useTerminalSettings.ts
export function useTerminalSettings() {
  const settings = ref<TerminalSettings>(defaultSettings)

  const updateSettings = async (newSettings: Partial<TerminalSettings>) => {
    const updated = await settingsService.update(newSettings)
    settings.value = updated
    return updated
  }

  return {
    settings: readonly(settings),
    updateSettings
  }
}
```

## Performance Refactoring

### Memory Optimization

#### Before (Memory Leaks)
```go
func (s *Service) ProcessStream() {
    buffer := make([]byte, 0, 1024*1024) // 1MB buffer

    for {
        data := <-s.inputChannel
        buffer = append(buffer, data...)

        // Buffer never cleared - memory leak
        if len(buffer) > 1024*1024 {
            // Process but keep growing
            processData(buffer)
        }
    }
}
```

#### After (Memory Efficient)
```go
func (s *Service) ProcessStream() {
    buffer := make([]byte, 0, 4096) // Smaller buffer

    for {
        data := <-s.inputChannel

        // Process in chunks, don't accumulate
        if err := s.processChunk(data); err != nil {
            s.logger.Error("Processing failed", err)
            continue
        }

        // Reset buffer regularly
        if len(buffer) > cap(buffer)/2 {
            buffer = buffer[:0]
        }
    }
}

func (s *Service) processChunk(data []byte) error {
    // Process data without accumulation
    return nil
}
```

### Frontend Performance

#### Before (Inefficient Reactivity)
```vue
<script setup lang="ts">
const largeArray = ref([]) // Large reactive array
const computedValue = computed(() => {
  return largeArray.value.filter(item => item.active)
})

function addItems(items: any[]) {
  largeArray.value.push(...items) // Triggers many updates
}
</script>
```

#### After (Performance Optimized)
```vue
<script setup lang="ts">
const largeArray = shallowRef([]) // Shallow reactivity for large arrays
const activeItems = computed(() => {
  return largeArray.value.filter(item => item.active)
})

function addItems(items: any[]) {
  // Batch updates
  const newValue = [...largeArray.value, ...items]
  largeArray.value = newValue
}

// Use VueUse utilities for performance
const debouncedUpdate = useDebounceFn(updateItems, 100)
</script>
```

## Testing Refactoring

### Test Structure

#### Before (Monolithic Test)
```go
func TestTerminalService(t *testing.T) {
    service := NewService()

    // Test setup
    session1, _ := service.CreateSession("test1")
    session2, _ := service.CreateSession("test2")

    // Test execution
    result1, _ := session1.ExecuteCommand("echo hello")
    result2, _ := session2.ExecuteCommand("echo world")

    // Test cleanup
    _ = service.CloseSession(session1.ID)
    _ = service.CloseSession(session2.ID)

    // All functionality tested in one large test
}
```

#### After (Focused Tests)
```go
func TestTerminalService_CreateSession(t *testing.T) {
    tests := []struct {
        name     string
        sessionID string
        wantErr  bool
    }{
        {"valid session", "test-session", false},
        {"empty session", "", true},
        {"duplicate session", "test-session", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service := NewTestService(t)

            session, err := service.CreateSession(tt.sessionID)

            if tt.wantErr {
                assert.Error(t, err)
                assert.Nil(t, session)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, session)
                assert.Equal(t, tt.sessionID, session.ID)
            }
        })
    }
}

func TestTerminalService_ExecuteCommand(t *testing.T) {
    service := NewTestService(t)
    session := createTestSession(t, service, "test")

    tests := []struct {
        name     string
        command  string
        wantErr  bool
        expected string
    }{
        {"echo command", "echo hello", false, "hello"},
        {"invalid command", "invalid-command", true, ""},
        {"empty command", "", false, ""},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := session.ExecuteCommand(tt.command)

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                if tt.expected != "" {
                    assert.Contains(t, result, tt.expected)
                }
            }
        })
    }
}
```

## Database/Storage Refactoring

### Repository Pattern

#### Before (Direct Database Access)
```go
type SettingsService struct {
    db *sql.DB
}

func (s *SettingsService) GetSettings(id string) (*Settings, error) {
    row := s.db.QueryRow("SELECT * FROM settings WHERE id = ?", id)
    // Direct SQL scattered throughout service
}

func (s *SettingsService) UpdateSettings(settings *Settings) error {
    _, err := s.db.Exec("UPDATE settings SET ... WHERE id = ?", settings.ID)
    // More direct SQL
}
```

#### After (Repository Pattern)
```go
type SettingsRepository interface {
    GetByID(id string) (*Settings, error)
    Save(settings *Settings) error
    Delete(id string) error
    List() ([]*Settings, error)
}

type SettingsService struct {
    repo SettingsRepository
}

func (s *SettingsService) GetSettings(id string) (*Settings, error) {
    return s.repo.GetByID(id)
}

func (s *SettingsService) UpdateSettings(settings *Settings) error {
    return s.repo.Save(settings)
}
```

## Configuration Refactoring

### Configuration Management

#### Before (Hardcoded Values)
```go
type Service struct {
    port        = 8080
    timeout     = 30 * time.Second
    maxSessions = 100
}
```

#### After (Configuration Injection)
```go
type Config struct {
    Port        int           `yaml:"port"`
    Timeout     time.Duration `yaml:"timeout"`
    MaxSessions int           `yaml:"max_sessions"`
}

type Service struct {
    config Config
}

func NewService(config Config) *Service {
    return &Service{config: config}
}
```

## Refactoring Checklist

### Before Refactoring
- [ ] Understand current codebase
- [ ] Identify refactoring goals
- [ ] Write tests for existing functionality
- [ ] Create backup/branch
- [ ] Communicate changes to team

### During Refactoring
- [ ] Make small, incremental changes
- [ ] Run tests after each change
- [ ] Commit frequently with descriptive messages
- [ ] Maintain functionality throughout
- [ ] Monitor performance impact

### After Refactoring
- [ ] Verify all tests pass
- [ ] Run performance benchmarks
- [ ] Update documentation
- [ ] Code review with team
- [ ] Monitor for issues in production

## Common Code Smells and Solutions

### Long Method
- **Problem**: Methods doing too many things
- **Solution**: Extract smaller methods with single responsibilities

### Large Class
- **Problem**: Classes with too many responsibilities
- **Solution**: Split into smaller, focused classes

### Duplicate Code
- **Problem**: Same code in multiple places
- **Solution**: Extract to shared functions or classes

### Long Parameter List
- **Problem**: Methods with many parameters
- **Solution**: Use parameter objects or builder pattern

### Feature Envy
- **Problem**: Method using more data from other classes
- **Solution**: Move method to the appropriate class

### Data Clumps
- **Problem**: Same group of parameters passed together
- **Solution**: Extract into a class or struct

### Primitive Obsession
- **Problem**: Using primitives instead of objects
- **Solution**: Create value objects with behavior

## Refactoring Tools

### Go
- `gofmt`: Code formatting
- `goimports`: Import management
- `golangci-lint`: Comprehensive linting
- `go vet`: Static analysis
- `gorename`: Safe renaming

### TypeScript/JavaScript
- Prettier: Code formatting
- ESLint: Linting and code quality
- TypeScript compiler: Type checking
- VS Code: Refactoring support

### Automated Refactoring
- Use IDE refactoring tools when possible
- Test thoroughly after automated changes
- Review automated changes before committing

## Continuous Refactoring

### Integration with Development Workflow
1. **Pre-commit**: Run formatting and linting
2. **Pre-push**: Run test suite
3. **CI/CD**: Comprehensive checks
4. **Code Review**: Refactoring opportunities
5. **Sprint Review**: Technical debt assessment

### Technical Debt Management
- Track refactoring tasks in backlog
- Allocate time for refactoring in sprints
- Prioritize based on impact and effort
- Monitor code quality metrics over time

## Conclusion

Refactoring should be a continuous process, not a one-time event. By following these guidelines and making small, incremental improvements, the codebase will remain maintainable, performant, and high-quality.

Remember:
- Test before refactoring
- Make small changes
- Maintain functionality
- Learn from each refactoring session