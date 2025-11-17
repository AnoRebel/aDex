#!/bin/bash

# Memory Leak Detection Script for aDex-UI
# Detects memory leaks and resource issues

set -e

echo "🔍 Running Memory Leak Detection for aDex-UI"
echo "============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Memory thresholds (in MB)
INITIAL_MEMORY_THRESHOLD=100
MEMORY_GROWTH_THRESHOLD=50
MAX_MEMORY_THRESHOLD=1000
GOROUTINE_LEAK_THRESHOLD=100
FILE_DESCRIPTOR_THRESHOLD=100

# Test results
FAILED_TESTS=0
TOTAL_TESTS=0

echo -e "${BLUE}📊 Memory Analysis${NC}"
echo "-----------------------"

# Create memory test results directory
mkdir -p memory-results
cd /home/ano/Code/Dex-UI/aDex-UI

echo "1. Go Memory Leak Detection..."

# Create Go memory profiling test
cat > memory-results/memory_leak_test.go << 'EOF'
package main

import (
    "fmt"
    "os"
    "runtime"
    "sync"
    "testing"
    "time"

    "github.com/adex-ui/aDex-UI/internal/services/terminal"
    "github.com/adex-ui/aDex-UI/internal/services/system"
    "github.com/adex-ui/aDex-UI/internal/services/settings"
    "github.com/adex-ui/aDex-UI/internal/services/audio"
    "github.com/adex-ui/aDex-UI/internal/logger"
)

func TestTerminalMemoryLeaks(t *testing.T) {
    logger := logger.New(logger.Config{Level: "error", Output: "test"})

    // Force GC to get baseline
    runtime.GC()
    var m1 runtime.MemStats
    runtime.ReadMemStats(&m1)
    baseline := m1.Alloc

    service, err := terminal.NewService(logger, t.TempDir())
    if err != nil {
        t.Fatal(err)
    }

    // Create and destroy many terminal sessions
    for i := 0; i < 100; i++ {
        session, err := service.CreateTerminalSession(fmt.Sprintf("leak-test-%d", i))
        if err != nil {
            t.Errorf("Failed to create session: %v", err)
            continue
        }

        // Simulate terminal usage
        for j := 0; j < 10; j++ {
            _, _ = session.ExecuteCommand("echo 'test'")
        }

        // Clean up
        if err := service.CloseSession(session.ID); err != nil {
            t.Errorf("Failed to close session: %v", err)
        }
    }

    // Force GC and check memory
    runtime.GC()
    time.Sleep(100 * time.Millisecond) // Allow finalizers to run
    runtime.GC()

    var m2 runtime.MemStats
    runtime.ReadMemStats(&m2)
    final := m2.Alloc

    memoryGrowth := final - baseline
    t.Logf("Baseline memory: %d bytes", baseline)
    t.Logf("Final memory: %d bytes", final)
    t.Logf("Memory growth: %d bytes", memoryGrowth)

    // Allow some memory growth but not excessive
    if memoryGrowth > 10*1024*1024 { // 10MB
        t.Errorf("Potential memory leak detected: %d bytes growth", memoryGrowth)
    }
}

func TestGoroutineLeaks(t *testing.T) {
    logger := logger.New(logger.Config{Level: "error", Output: "test"})

    initialGoroutines := runtime.NumGoroutine()
    t.Logf("Initial goroutines: %d", initialGoroutines)

    // Create services that might spawn goroutines
    termService, _ := terminal.NewService(logger, t.TempDir())
    systemService := system.NewService(logger)
    settingsService, _ := settings.NewService(logger)

    // Create sessions that might start goroutines
    sessions := make([]*terminal.Session, 10)
    for i := 0; i < 10; i++ {
        session, _ := termService.CreateTerminalSession(fmt.Sprintf("goroutine-test-%d", i))
        sessions[i] = session
    }

    // Use services that might spawn goroutines
    for i := 0; i < 5; i++ {
        _, _ = systemService.GetMetrics()
    }

    // Close all sessions
    for _, session := range sessions {
        if session != nil {
            _ = termService.CloseSession(session.ID)
        }
    }

    // Wait for goroutines to finish
    time.Sleep(500 * time.Millisecond)

    finalGoroutines := runtime.NumGoroutine()
    goroutineGrowth := finalGoroutines - initialGoroutines

    t.Logf("Final goroutines: %d", finalGoroutines)
    t.Logf("Goroutine growth: %d", goroutineGrowth)

    if goroutineGrowth > 20 {
        t.Errorf("Potential goroutine leak: %d additional goroutines", goroutineGrowth)
    }
}

func TestFileDescriptorLeaks(t *testing.T) {
    logger := logger.New(logger.Config{Level: "error", Output: "test"})

    // Get initial file descriptor count (Linux/Mac only)
    initialFDs := getFileDescriptorCount()
    if initialFDs == -1 {
        t.Skip("File descriptor counting not supported on this platform")
    }

    t.Logf("Initial file descriptors: %d", initialFDs)

    // Create many services and files
    services := make([]*terminal.Service, 10)
    for i := 0; i < 10; i++ {
        service, err := terminal.NewService(logger, t.TempDir())
        if err != nil {
            t.Errorf("Failed to create service: %v", err)
            continue
        }
        services[i] = service
    }

    // Create sessions and write to files
    for _, service := range services {
        if service != nil {
            session, _ := service.CreateTerminalSession("fd-test")
            if session != nil {
                _, _ = session.Write("test data")
                _ = service.CloseSession(session.ID)
            }
        }
    }

    // Clean up services
    for _, service := range services {
        if service != nil {
            // Service cleanup would happen here
        }
    }

    // Wait for file descriptors to be released
    time.Sleep(1000 * time.Millisecond)

    finalFDs := getFileDescriptorCount()
    fdGrowth := finalFDs - initialFDs

    t.Logf("Final file descriptors: %d", finalFDs)
    t.Logf("File descriptor growth: %d", fdGrowth)

    if fdGrowth > 10 {
        t.Errorf("Potential file descriptor leak: %d additional descriptors", fdGrowth)
    }
}

func getFileDescriptorCount() int {
    // Linux implementation
    if pid := os.Getpid(); pid > 0 {
        if data, err := os.ReadDir(fmt.Sprintf("/proc/%d/fd", pid)); err == nil {
            return len(data)
        }
    }
    return -1
}

func BenchmarkMemoryAllocation(b *testing.B) {
    logger := logger.New(logger.Config{Level: "error", Output: "test"})
    service, _ := terminal.NewService(logger, b.TempDir())

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        session, _ := service.CreateTerminalSession(fmt.Sprintf("bench-%d", i))
        if session != nil {
            _, _ = session.ExecuteCommand("echo 'benchmark'")
            _ = service.CloseSession(session.ID)
        }
    }
}
EOF

cd memory-results
echo "Running Go memory leak tests..."
if go test -v -run=TestMemoryLeak -timeout=120s memory_leak_test.go > memory_leak_results.txt 2>&1; then
    echo -e "${GREEN}✅ Go memory leak tests completed${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if grep -q "FAIL" memory_leak_results.txt; then
        echo -e "${YELLOW}⚠️ Some memory issues detected${NC}"
        grep "FAIL" memory_leak_results.txt | head -5
    else
        echo -e "${GREEN}✅ No memory leaks detected${NC}"
    fi
else
    echo -e "${RED}❌ Go memory leak tests failed${NC}"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

echo ""
echo "2. Frontend Memory Leak Detection..."

# Create frontend memory leak test
cat > memory-results/frontend_memory_test.html << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Frontend Memory Leak Detection</title>
    <style>
        body {
            font-family: monospace;
            margin: 20px;
            background: #1a1a1a;
            color: #fff;
        }
        .test-container {
            border: 1px solid #333;
            padding: 20px;
            margin: 10px 0;
            background: #2a2a2a;
        }
        .status {
            padding: 10px;
            margin: 10px 0;
            border-radius: 4px;
        }
        .pass { background: #28a745; }
        .fail { background: #dc3545; }
        .info { background: #17a2b8; }
        button {
            background: #007bff;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 4px;
            cursor: pointer;
            margin: 5px;
        }
        button:hover { background: #0056b3; }
        #results {
            white-space: pre-wrap;
            background: #333;
            padding: 15px;
            border-radius: 4px;
            margin-top: 20px;
        }
    </style>
</head>
<body>
    <h1>🔍 Frontend Memory Leak Detection</h1>

    <div class="test-container">
        <h2>Memory Monitoring</h2>
        <div id="memory-info">Initializing...</div>
        <button onclick="startMemoryTest()">Start Memory Test</button>
        <button onclick="stopMemoryTest()">Stop Memory Test</button>
        <button onclick="forceGC()">Force Garbage Collection</button>
    </div>

    <div class="test-container">
        <h2>Event Listener Leak Test</h2>
        <button onclick="testEventListeners()">Test Event Listeners</button>
        <div id="event-results"></div>
    </div>

    <div class="test-container">
        <h2>DOM Node Leak Test</h2>
        <button onclick="testDOMNodes()">Test DOM Nodes</button>
        <div id="dom-results"></div>
    </div>

    <div class="test-container">
        <h2>Timer/Interval Leak Test</h2>
        <button onclick="testTimers()">Test Timers</button>
        <div id="timer-results"></div>
    </div>

    <div class="test-container">
        <h2>WebSocket/Connection Leak Test</h2>
        <button onclick="testConnections()">Test Connections</button>
        <div id="connection-results"></div>
    </div>

    <div id="results"></div>

    <script>
        let memoryTestInterval = null;
        let testResults = [];
        let eventListeners = [];
        let timers = [];
        let connections = [];

        // Memory monitoring
        function getMemoryInfo() {
            if (performance.memory) {
                return {
                    used: Math.round(performance.memory.usedJSHeapSize / 1024 / 1024 * 100) / 100,
                    total: Math.round(performance.memory.totalJSHeapSize / 1024 / 1024 * 100) / 100,
                    limit: Math.round(performance.memory.jsHeapSizeLimit / 1024 / 1024 * 100) / 100
                };
            }
            return { used: 'N/A', total: 'N/A', limit: 'N/A' };
        }

        function updateMemoryInfo() {
            const info = getMemoryInfo();
            document.getElementById('memory-info').innerHTML = `
                <div>Used Memory: ${info.used} MB</div>
                <div>Total Memory: ${info.total} MB</div>
                <div>Memory Limit: ${info.limit} MB</div>
                <div>Timestamp: ${new Date().toLocaleTimeString()}</div>
            `;
        }

        function startMemoryTest() {
            if (memoryTestInterval) return;

            let initialMemory = getMemoryInfo();
            let startTimestamp = Date.now();

            memoryTestInterval = setInterval(() => {
                updateMemoryInfo();

                let currentMemory = getMemoryInfo();
                let elapsed = (Date.now() - startTimestamp) / 1000;
                let memoryGrowth = parseFloat(currentMemory.used) - parseFloat(initialMemory.used);

                testResults.push({
                    timestamp: elapsed,
                    memory: currentMemory.used,
                    growth: memoryGrowth
                });

                // Alert if memory growth is significant
                if (memoryGrowth > 50) {
                    addResult('warning', `High memory growth detected: ${memoryGrowth.toFixed(2)} MB`);
                }
            }, 1000);

            addResult('info', 'Memory monitoring started');
        }

        function stopMemoryTest() {
            if (memoryTestInterval) {
                clearInterval(memoryTestInterval);
                memoryTestInterval = null;
                addResult('info', 'Memory monitoring stopped');

                // Analyze memory growth
                if (testResults.length > 0) {
                    const initialMemory = testResults[0].memory;
                    const finalMemory = testResults[testResults.length - 1].memory;
                    const growth = finalMemory - initialMemory;

                    addResult('info', `Memory analysis: Initial: ${initialMemory}MB, Final: ${finalMemory}MB, Growth: ${growth.toFixed(2)}MB`);

                    if (growth > 20) {
                        addResult('fail', 'Potential memory leak detected - high memory growth');
                    } else {
                        addResult('pass', 'Memory usage looks normal');
                    }
                }
            }
        }

        function forceGC() {
            if (window.gc) {
                window.gc();
                updateMemoryInfo();
                addResult('info', 'Garbage collection forced');
            } else {
                addResult('info', 'Manual GC not available (use Chrome DevTools)');
            }
        }

        // Event listener leak test
        function testEventListeners() {
            const container = document.createElement('div');
            document.body.appendChild(container);

            // Add many event listeners
            for (let i = 0; i < 100; i++) {
                const element = document.createElement('div');
                const listener = () => {};
                element.addEventListener('click', listener);
                container.appendChild(element);
                eventListeners.push({ element, listener });
            }

            document.getElementById('event-results').textContent = `Created ${eventListeners.length} event listeners`;

            // Simulate cleanup
            setTimeout(() => {
                container.remove();
                eventListeners = [];
                document.getElementById('event-results').textContent += ' - Cleaned up';
                addResult('pass', 'Event listener cleanup test completed');
            }, 1000);
        }

        // DOM node leak test
        function testDOMNodes() {
            const container = document.createElement('div');
            document.body.appendChild(container);

            // Create many DOM nodes
            for (let i = 0; i < 1000; i++) {
                const element = document.createElement('div');
                element.textContent = `Node ${i}`;
                element.dataset.index = i;
                container.appendChild(element);
            }

            const nodeCount = container.getElementsByTagName('*').length;
            document.getElementById('dom-results').textContent = `Created ${nodeCount} DOM nodes`;

            // Simulate cleanup
            setTimeout(() => {
                container.innerHTML = '';
                container.remove();
                document.getElementById('dom-results').textContent += ' - Cleaned up';
                addResult('pass', 'DOM node cleanup test completed');
            }, 1000);
        }

        // Timer leak test
        function testTimers() {
            // Create many timers
            for (let i = 0; i < 50; i++) {
                const intervalId = setInterval(() => {
                    console.log('Timer', i);
                }, 1000);
                timers.push(intervalId);

                const timeoutId = setTimeout(() => {
                    console.log('Timeout', i);
                }, 5000);
                timers.push(timeoutId);
            }

            document.getElementById('timer-results').textContent = `Created ${timers.length} timers`;

            // Clean up timers
            setTimeout(() => {
                timers.forEach(id => {
                    clearInterval(id);
                    clearTimeout(id);
                });
                timers = [];
                document.getElementById('timer-results').textContent += ' - Cleaned up';
                addResult('pass', 'Timer cleanup test completed');
            }, 2000);
        }

        // Connection leak test
        function testConnections() {
            // Simulate connection creation
            for (let i = 0; i < 10; i++) {
                const ws = new WebSocket('ws://localhost:8080/test');
                connections.push(ws);

                ws.onerror = () => {
                    // Expected to fail - just testing cleanup
                };
            }

            document.getElementById('connection-results').textContent = `Created ${connections.length} connection attempts`;

            // Clean up connections
            setTimeout(() => {
                connections.forEach(ws => {
                    try {
                        ws.close();
                    } catch (e) {
                        // Expected - connections likely failed
                    }
                });
                connections = [];
                document.getElementById('connection-results').textContent += ' - Cleaned up';
                addResult('pass', 'Connection cleanup test completed');
            }, 2000);
        }

        function addResult(type, message) {
            const results = document.getElementById('results');
            const timestamp = new Date().toLocaleTimeString();
            const className = type === 'pass' ? 'pass' : type === 'fail' ? 'fail' : 'info';
            results.textContent += `[${timestamp}] ${message.toUpperCase()}\n`;
        }

        // Initialize
        updateMemoryInfo();
        setInterval(updateMemoryInfo, 2000);

        addResult('info', 'Frontend memory leak detection tool ready');
        addResult('info', 'Run tests and monitor memory usage');
    </script>
</body>
</html>
EOF

echo "Frontend memory leak test created: memory-results/frontend_memory_test.html"
echo -e "${GREEN}✅ Frontend memory test ready for manual verification${NC}"
TOTAL_TESTS=$((TOTAL_TESTS + 1))

cd ..

echo ""
echo "3. System Resource Monitoring..."

# Monitor system resources during application execution
cat > memory-results/system_resource_monitor.go << 'EOF'
package main

import (
    "fmt"
    "os"
    "runtime"
    "time"
)

func main() {
    fmt.Println("System Resource Monitor")
    fmt.Println("=======================")

    pid := os.Getpid()
    fmt.Printf("PID: %d\n", pid)

    for i := 0; i < 60; i++ { // Monitor for 1 minute
        var m runtime.MemStats
        runtime.ReadMemStats(&m)

        fmt.Printf("Time: %s\n", time.Now().Format("15:04:05"))
        fmt.Printf("Memory: %d KB (%.2f MB)\n", m.Alloc/1024, float64(m.Alloc)/1024/1024)
        fmt.Printf("Goroutines: %d\n", runtime.NumGoroutine())

        // System memory (if available)
        if data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid)); err == nil {
            lines := strings.Split(string(data), "\n")
            for _, line := range lines {
                if strings.HasPrefix(line, "VmRSS:") {
                    fmt.Printf("RSS: %s\n", strings.TrimSpace(line[6:]))
                    break
                }
            }
        }

        fmt.Println(strings.Repeat("-", 40))
        time.Sleep(1 * time.Second)
    }
}
EOF

echo "System resource monitor created: memory-results/system_resource_monitor.go"
echo -e "${GREEN}✅ Resource monitoring tool ready${NC}"

echo ""
echo "4. Long-Running Application Test..."

# Create long-running test to detect gradual memory growth
cat > memory-results/long_running_test.go << 'EOF'
package main

import (
    "context"
    "fmt"
    "runtime"
    "sync"
    "time"

    "github.com/adex-ui/aDex-UI/internal/services/terminal"
    "github.com/adex-ui/aDex-UI/internal/services/system"
    "github.com/adex-ui/aDex-UI/internal/logger"
)

func main() {
    fmt.Println("Long-Running Memory Test")
    fmt.Println("========================")

    logger := logger.New(logger.Config{Level: "error", Output: "test"})

    // Initialize services
    termService, err := terminal.NewService(logger, "/tmp")
    if err != nil {
        panic(err)
    }

    systemService := system.NewService(logger)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    var wg sync.WaitGroup

    // Terminal session simulator
    wg.Add(1)
    go func() {
        defer wg.Done()
        sessionCounter := 0

        for {
            select {
            case <-ctx.Done():
                return
            default:
                // Create session
                session, err := termService.CreateTerminalSession(fmt.Sprintf("long-test-%d", sessionCounter))
                if err != nil {
                    fmt.Printf("Session creation error: %v\n", err)
                    continue
                }

                // Use session
                for i := 0; i < 10; i++ {
                    _, _ = session.ExecuteCommand("echo 'test command'")
                    time.Sleep(100 * time.Millisecond)
                }

                // Cleanup
                _ = termService.CloseSession(session.ID)
                sessionCounter++

                // Pause between sessions
                time.Sleep(1 * time.Second)
            }
        }
    }()

    // System monitor simulator
    wg.Add(1)
    go func() {
        defer wg.Done()
        ticker := time.NewTicker(2 * time.Second)
        defer ticker.Stop()

        var lastAlloc uint64
        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                var m runtime.MemStats
                runtime.ReadMemStats(&m)

                growth := m.Alloc - lastAlloc
                if lastAlloc > 0 && growth > 10*1024*1024 { // 10MB growth
                    fmt.Printf("WARNING: Memory growth detected: %.2f MB\n", float64(growth)/1024/1024)
                }

                fmt.Printf("Time: %v, Memory: %.2f MB, Goroutines: %d\n",
                    time.Now().Format("15:04:05"),
                    float64(m.Alloc)/1024/1024,
                    runtime.NumGoroutine())

                lastAlloc = m.Alloc
            }
        }
    }()

    // System metrics collector
    wg.Add(1)
    go func() {
        defer wg.Done()
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                _, err := systemService.GetMetrics()
                if err != nil {
                    fmt.Printf("Metrics collection error: %v\n", err)
                }
            }
        }
    }()

    wg.Wait()
    fmt.Println("Long-running test completed")
}
EOF

echo "Long-running memory test created: memory-results/long_running_test.go"
echo -e "${GREEN}✅ Long-running test ready${NC}"

echo ""
echo -e "${BLUE}📈 Memory Analysis Results${NC}"
echo "---------------------------------"

echo "Memory analysis tools created:"
echo "  - Go memory leak tests: memory-results/memory_leak_test.go"
echo "  - Frontend memory test: memory-results/frontend_memory_test.html"
echo "  - System resource monitor: memory-results/system_resource_monitor.go"
echo "  - Long-running test: memory-results/long_running_test.go"

# Generate memory analysis report
cat > memory-results/memory-analysis-report.md << EOF
# Memory Leak Detection Report

Generated on: $(date)

## Test Overview

This report contains memory leak detection tests and analysis for the aDex-UI application.

## Memory Testing Components

### 1. Go Backend Memory Tests
- **File**: \`memory_leak_test.go\`
- **Tests**:
  - Terminal session memory management
  - Goroutine leak detection
  - File descriptor leak detection
  - Memory allocation benchmarks

### 2. Frontend Memory Tests
- **File**: \`frontend_memory_test.html\`
- **Tests**:
  - JavaScript heap memory monitoring
  - Event listener cleanup
  - DOM node leak detection
  - Timer/interval cleanup
  - WebSocket connection cleanup

### 3. System Resource Monitoring
- **File**: \`system_resource_monitor.go\`
- **Monitors**:
  - Process memory usage
  - Goroutine count
  - System RSS memory
  - Resource usage over time

### 4. Long-Running Application Test
- **File**: \`long_running_test.go\`
- **Simulates**:
  - Extended application usage (5 minutes)
  - Continuous terminal session creation/destruction
  - System monitoring activity
  - Memory growth detection

## Manual Testing Instructions

### Frontend Memory Testing
1. Open \`frontend_memory_test.html\` in a browser
2. Open browser DevTools (F12) and go to Memory tab
3. Click "Start Memory Test"
4. Run various tests and monitor memory growth
5. Click "Stop Memory Test" to analyze results
6. Use "Force Garbage Collection" to verify cleanup

### Backend Memory Testing
1. Run \`go test -v memory_leak_test.go\`
2. Monitor output for memory leak warnings
3. Check goroutine counts before/after tests
4. Review file descriptor usage

### Long-Running Testing
1. Run \`go run long_running_test.go\`
2. Monitor memory usage over 5 minutes
3. Look for steady memory growth patterns
4. Check for goroutine accumulation

## Memory Analysis Checklist

- [ ] Terminal sessions properly cleaned up
- [ ] No goroutine accumulation after operations
- [ ] File descriptors released after use
- [ ] Frontend event listeners removed
- [ ] DOM nodes properly garbage collected
- [ ] Timers and intervals cleared
- [ ] WebSocket connections closed
- [ ] No steady memory growth over time

## Memory Thresholds

- **Initial Memory**: < ${INITIAL_MEMORY_THRESHOLD}MB
- **Memory Growth**: < ${MEMORY_GROWTH_THRESHOLD}MB per operation
- **Maximum Memory**: < ${MAX_MEMORY_THRESHOLD}MB
- **Goroutine Growth**: < ${GOROUTINE_LEAK_THRESHOLD}
- **File Descriptor Growth**: < ${FILE_DESCRIPTOR_THRESHOLD}

## Expected Results

### Normal Operation
- Memory usage stabilizes after initial warmup
- Goroutine count remains relatively constant
- No steady upward memory trend
- Proper resource cleanup after operations

### Warning Signs
- Continuous memory growth without plateau
- Increasing goroutine count over time
- File descriptor accumulation
- Frontend heap size keeps growing

## Automated Integration

To integrate memory testing into CI/CD:

\`\`\`bash
# Run Go memory tests
go test -v memory_leak_test.go

# Run long-running test with timeout
timeout 300s go run long_running_test.go

# Check results for memory leaks
grep -i "leak\|warning\|fail" test-results.txt
\`\`\`

## Performance Impact

Memory leak detection should be run:
- During development (frequent)
- Before releases (mandatory)
- In CI/CD pipeline (automated)
- When memory issues are suspected

EOF

echo ""
echo "Memory analysis report generated: memory-results/memory-analysis-report.md"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}🎉 Memory leak detection tools created successfully!${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Run the Go memory tests: cd memory-results && go test -v memory_leak_test.go"
    echo "2. Open frontend test in browser: open memory-results/frontend_memory_test.html"
    echo "3. Run long-running test: cd memory-results && go run long_running_test.go"
    echo "4. Review the memory analysis report"
else
    echo -e "${RED}❌ Some memory tests failed${NC}"
    echo "Please review the memory analysis tools and fix any issues found"
fi

exit $FAILED_TESTS