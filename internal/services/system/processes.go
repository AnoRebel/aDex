package system

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"aDex-UI/internal/models"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/process"
)

// ProcessMonitor handles process metrics collection
type ProcessMonitor struct {
	processCache map[int32]*process.Process
	lastUpdate   time.Time
	isInitialized bool
}

// NewProcessMonitor creates a new process monitor instance
func NewProcessMonitor() *ProcessMonitor {
	return &ProcessMonitor{
		processCache: make(map[int32]*process.Process),
		isInitialized: false,
	}
}

// GetProcessMetrics collects current process metrics
func (pm *ProcessMonitor) GetProcessMetrics(ctx context.Context, limit int) (*models.ProcessMetrics, error) {
	// Get all processes
	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get processes: %w", err)
	}

	// Convert to our model and collect metrics
	var processInfos []models.ProcessInfo
	var totalProcesses, runningProcesses, sleepingProcesses, stoppedProcesses, zombieProcesses int

	for _, p := range processes {
		info, err := pm.getProcessInfo(ctx, p)
		if err != nil {
			// Skip processes we can't get info for
			continue
		}

		processInfos = append(processInfos, info)
		totalProcesses++

		// Count by status
		switch strings.ToLower(info.Status) {
		case "running", "run":
			runningProcesses++
		case "sleeping", "sleep":
			sleepingProcesses++
		case "stopped", "stop":
			stoppedProcesses++
		case "zombie":
			zombieProcesses++
		default:
			// Default to sleeping for unknown statuses
			sleepingProcesses++
		}
	}

	// Sort by CPU usage (descending)
	sort.Slice(processInfos, func(i, j int) bool {
		return processInfos[i].CPUPercent > processInfos[j].CPUPercent
	})

	// Limit results if requested
	if limit > 0 && len(processInfos) > limit {
		processInfos = processInfos[:limit]
	}

	// Get host info
	hostInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		hostInfo = &host.InfoStat{}
	}

	metrics := &models.ProcessMetrics{
		Processes:        processInfos,
		TotalProcesses:   totalProcesses,
		RunningProcesses: runningProcesses,
		SleepingProcesses: sleepingProcesses,
		StoppedProcesses: stoppedProcesses,
		ZombieProcesses:  zombieProcesses,
		Timestamp:        time.Now(),
	}

	// Update cache
	pm.updateProcessCache(processes)
	pm.lastUpdate = metrics.Timestamp
	pm.isInitialized = true

	return metrics, nil
}

// getProcessInfo converts a gopsutil process to our model
func (pm *ProcessMonitor) getProcessInfo(ctx context.Context, p *process.Process) (models.ProcessInfo, error) {
	pid := p.Pid

	// Get basic info
	name, err := p.NameWithContext(ctx)
	if err != nil {
		name = "unknown"
	}

	// Get command line
	cmdline, err := p.CmdlineWithContext(ctx)
	if err != nil {
		cmdline = name
	}

	// Get executable path
	exe, err := p.ExeWithContext(ctx)
	if err != nil {
		exe = ""
	}

	// Get user info
	username := ""
	uid, err := p.UidsWithContext(ctx)
	if err == nil && len(uid) > 0 {
		if u, err := user.LookupId(strconv.Itoa(int(uid[0]))); err == nil {
			username = u.Username
		}
	}

	// Get status
	status, err := p.StatusWithContext(ctx)
	if err != nil {
		status = "unknown"
	} else if len(status) > 0 {
		status = status[0]
	}

	// Get CPU percentage
	cpuPercent, err := p.CPUPercentWithContext(ctx)
	if err != nil {
		cpuPercent = 0
	}

	// Get memory info
	memInfo, err := p.MemoryInfoWithContext(ctx)
	if err != nil {
		memInfo = &process.MemoryInfoStat{}
	}

	// Get memory percentages
	memPercent, err := p.MemoryPercentWithContext(ctx)
	if err != nil {
		memPercent = 0
	}

	// Get creation time
	createTime, err := p.CreateTimeWithContext(ctx)
	var createTimeFormatted time.Time
	if err == nil {
		createTimeFormatted = time.Unix(createTime/1000, 0)
	}

	// Get number of threads
	numThreads, err := p.NumThreadsWithContext(ctx)
	if err != nil {
		numThreads = 0
	}

	// Get number of file descriptors (Unix-like systems only)
	var numFDs int
	if runtime.GOOS != "windows" {
		fds, err := p.NumFDsWithContext(ctx)
		if err == nil {
			numFDs = fds
		}
	}

	// Get current working directory
	cwd, err := p.CwdWithContext(ctx)
	if err != nil {
		cwd = ""
	}

	// Build process info
	info := models.ProcessInfo{
		PID:           int(pid),
		Name:          name,
		Command:       cmdline,
		User:          username,
		Status:        status,
		CPUPercent:    cpuPercent,
		MemoryPercent: memPercent,
		MemoryRSS:     memInfo.RSS,
		MemoryVMS:     memInfo.VMS,
		CreateTime:    createTimeFormatted,
		NumThreads:    numThreads,
		NumFDs:        numFDs,
		CWD:           cwd,
		Executable:    exe,
	}

	// Get parent PID
	ppid, err := p.PpidWithContext(ctx)
	if err == nil {
		info.PPID = int(ppid)
	}

	return info, nil
}

// GetProcessByPID returns information for a specific process
func (pm *ProcessMonitor) GetProcessByPID(ctx context.Context, pid int) (*models.ProcessInfo, error) {
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return nil, fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	info, err := pm.getProcessInfo(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("failed to get process info for %d: %w", pid, err)
	}

	return &info, nil
}

// GetProcessesByName returns all processes with the given name
func (pm *ProcessMonitor) GetProcessesByName(ctx context.Context, name string) ([]models.ProcessInfo, error) {
	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get processes: %w", err)
	}

	var matches []models.ProcessInfo
	for _, p := range processes {
		processName, err := p.NameWithContext(ctx)
		if err != nil {
			continue
		}

		if processName == name {
			info, err := pm.getProcessInfo(ctx, p)
			if err != nil {
				continue
			}
			matches = append(matches, info)
		}
	}

	return matches, nil
}

// GetProcessesByUser returns all processes owned by the given user
func (pm *ProcessMonitor) GetProcessesByUser(ctx context.Context, username string) ([]models.ProcessInfo, error) {
	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get processes: %w", err)
	}

	var matches []models.ProcessInfo
	for _, p := range processes {
		uids, err := p.UidsWithContext(ctx)
		if err != nil {
			continue
		}

		if len(uids) > 0 {
			if u, err := user.LookupId(strconv.Itoa(int(uids[0]))); err == nil && u.Username == username {
				info, err := pm.getProcessInfo(ctx, p)
				if err != nil {
					continue
				}
				matches = append(matches, info)
			}
		}
	}

	return matches, nil
}

// GetTopProcesses returns the top processes by the specified metric
func (pm *ProcessMonitor) GetTopProcesses(ctx context.Context, metric string, limit int) ([]models.ProcessInfo, error) {
	metrics, err := pm.GetProcessMetrics(ctx, 0) // Get all processes
	if err != nil {
		return nil, err
	}

	processes := make([]models.ProcessInfo, len(metrics.Processes))
	copy(processes, metrics.Processes)

	// Sort by the specified metric
	switch strings.ToLower(metric) {
	case "cpu":
		sort.Slice(processes, func(i, j int) bool {
			return processes[i].CPUPercent > processes[j].CPUPercent
		})
	case "memory":
		sort.Slice(processes, func(i, j int) bool {
			return processes[i].MemoryPercent > processes[j].MemoryPercent
		})
	case "threads":
		sort.Slice(processes, func(i, j int) bool {
			return processes[i].NumThreads > processes[j].NumThreads
		})
	case "fds":
		sort.Slice(processes, func(i, j int) bool {
			return processes[i].NumFDs > processes[j].NumFDs
		})
	default:
		// Default to CPU
		sort.Slice(processes, func(i, j int) bool {
			return processes[i].CPUPercent > processes[j].CPUPercent
		})
	}

	// Limit results
	if limit > 0 && len(processes) > limit {
		processes = processes[:limit]
	}

	return processes, nil
}

// GetProcessTree returns a tree structure of processes
func (pm *ProcessMonitor) GetProcessTree(ctx context.Context) (map[int][]models.ProcessInfo, error) {
	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get processes: %w", err)
	}

	// Create a map of PID to process info
	processMap := make(map[int32]models.ProcessInfo)
	var rootProcesses []models.ProcessInfo

	for _, p := range processes {
		info, err := pm.getProcessInfo(ctx, p)
		if err != nil {
			continue
		}
		processMap[p.Pid] = info

		// Track potential root processes (orphaned processes)
		if info.PPID == 0 || info.PPID == 1 {
			rootProcesses = append(rootProcesses, info)
		}
	}

	// Build parent-child relationships
	tree := make(map[int][]models.ProcessInfo)
	for pid, info := range processMap {
		if info.PPID > 0 && info.PPID != 1 {
			parentPID := info.PPID
			tree[parentPID] = append(tree[parentPID], info)
		}
	}

	return tree, nil
}

// GetSystemProcesses returns system-wide process statistics
func (pm *ProcessMonitor) GetSystemProcesses(ctx context.Context) (map[string]interface{}, error) {
	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system processes: %w", err)
	}

	stats := map[string]interface{}{
		"total_processes": len(processes),
		"os":             runtime.GOOS,
		"arch":           runtime.GOARCH,
	}

	// Count processes by status
	statusCounts := make(map[string]int)
	for _, p := range processes {
		status, err := p.StatusWithContext(ctx)
		if err != nil {
			continue
		}

		if len(status) > 0 {
			statusCounts[status[0]]++
		}
	}
	stats["status_counts"] = statusCounts

	// Get current process info
	currentPID := os.Getpid()
	currentProcess, err := process.NewProcess(int32(currentPID))
	if err == nil {
		stats["current_process"] = map[string]interface{}{
			"pid": currentPID,
		}

		if name, err := currentProcess.NameWithContext(ctx); err == nil {
			stats["current_process"].(map[string]interface{})["name"] = name
		}
	}

	return stats, nil
}

// IsProcessRunning checks if a process is currently running
func (pm *ProcessMonitor) IsProcessRunning(pid int) bool {
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return false
	}

	// Try to get the process status
	_, err = p.StatusWithContext(context.Background())
	return err == nil
}

// KillProcess terminates a process
func (pm *ProcessMonitor) KillProcess(pid int) error {
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	return p.Kill()
}

// GetProcessChildren returns all child processes of a given process
func (pm *ProcessMonitor) GetProcessChildren(ctx context.Context, pid int) ([]models.ProcessInfo, error) {
	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get processes: %w", err)
	}

	var children []models.ProcessInfo
	for _, p := range processes {
		ppid, err := p.PpidWithContext(ctx)
		if err != nil {
			continue
		}

		if int(ppid) == pid {
			info, err := pm.getProcessInfo(ctx, p)
			if err != nil {
				continue
			}
			children = append(children, info)
		}
	}

	return children, nil
}

// GetProcessEnvironment returns environment variables for a process
func (pm *ProcessMonitor) GetProcessEnvironment(ctx context.Context, pid int) (map[string]string, error) {
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return nil, fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	environ, err := p.EnvironWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get process environment: %w", err)
	}

	envMap := make(map[string]string)
	for _, env := range environ {
		if parts := strings.SplitN(env, "=", 2); len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	return envMap, nil
}

// GetProcessOpenFiles returns open file descriptors for a process
func (pm *ProcessMonitor) GetProcessOpenFiles(ctx context.Context, pid int) ([]string, error) {
	if runtime.GOOS == "windows" {
		return nil, fmt.Errorf("open files not available on Windows")
	}

	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return nil, fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	files, err := p.OpenFilesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get open files: %w", err)
	}

	var filePaths []string
	for _, file := range files {
		filePaths = append(filePaths, file.Path)
	}

	return filePaths, nil
}

// GetProcessConnections returns network connections for a process
func (pm *ProcessMonitor) GetProcessConnections(ctx context.Context, pid int) ([]string, error) {
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return nil, fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	connections, err := p.ConnectionsWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get process connections: %w", err)
	}

	var connStrings []string
	for _, conn := range connections {
		connStr := fmt.Sprintf("%s:%d->%s:%d (%s)",
			conn.Laddr.IP, conn.Laddr.Port,
			conn.Raddr.IP, conn.Raddr.Port,
			conn.Status)
		connStrings = append(connStrings, connStr)
	}

	return connStrings, nil
}

// updateProcessCache updates the internal process cache
func (pm *ProcessMonitor) updateProcessCache(processes []*process.Process) {
	// Clear existing cache
	for k := range pm.processCache {
		delete(pm.processCache, k)
	}

	// Add current processes to cache
	for _, p := range processes {
		pm.processCache[p.Pid] = p
	}
}

// ValidateProcessMetrics validates process metrics for consistency
func (pm *ProcessMonitor) ValidateProcessMetrics(metrics *models.ProcessMetrics) error {
	if metrics == nil {
		return fmt.Errorf("process metrics is nil")
	}

	// Validate process counts
	total := metrics.RunningProcesses + metrics.SleepingProcesses +
		metrics.StoppedProcesses + metrics.ZombieProcesses

	if total > metrics.TotalProcesses {
		return fmt.Errorf("status process counts (%d) exceed total (%d)",
			total, metrics.TotalProcesses)
	}

	// Validate individual processes
	for i, proc := range metrics.Processes {
		if proc.PID <= 0 {
			return fmt.Errorf("invalid PID for process %d: %d", i, proc.PID)
		}

		if proc.Name == "" {
			return fmt.Errorf("empty name for process %d", i)
		}

		if proc.CPUPercent < 0 || proc.CPUPercent > 100*float64(runtime.NumCPU()) {
			return fmt.Errorf("invalid CPU usage for process %d: %f", i, proc.CPUPercent)
		}

		if proc.MemoryPercent < 0 || proc.MemoryPercent > 100 {
			return fmt.Errorf("invalid memory usage for process %d: %f", i, proc.MemoryPercent)
		}

		if proc.MemoryRSS == 0 && proc.MemoryVMS == 0 {
			return fmt.Errorf("both RSS and VMS are zero for process %d", i)
		}
	}

	// Validate timestamp
	if metrics.Timestamp.IsZero() {
		return fmt.Errorf("process metrics timestamp is zero")
	}

	return nil
}

// GetProcessStatistics returns additional process statistics
func (pm *ProcessMonitor) GetProcessStatistics(ctx context.Context) (map[string]interface{}, error) {
	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get process statistics: %w", err)
	}

	stats := map[string]interface{}{
		"total_processes": len(processes),
		"timestamp":       time.Now(),
	}

	// Calculate statistics
	var totalCPU, totalMemory float64
	var totalThreads, totalFDs int
	statusCounts := make(map[string]int)

	for _, p := range processes {
		// CPU usage
		if cpuPercent, err := p.CPUPercentWithContext(ctx); err == nil {
			totalCPU += cpuPercent
		}

		// Memory usage
		if memPercent, err := p.MemoryPercentWithContext(ctx); err == nil {
			totalMemory += memPercent
		}

		// Threads
		if threads, err := p.NumThreadsWithContext(ctx); err == nil {
			totalThreads += threads
		}

		// File descriptors (Unix-like systems only)
		if runtime.GOOS != "windows" {
			if fds, err := p.NumFDsWithContext(ctx); err == nil {
				totalFDs += fds
			}
		}

		// Status counts
		if status, err := p.StatusWithContext(ctx); err == nil && len(status) > 0 {
			statusCounts[status[0]]++
		}
	}

	stats["average_cpu_percent"] = totalCPU / float64(len(processes))
	stats["average_memory_percent"] = totalMemory / float64(len(processes))
	stats["total_threads"] = totalThreads
	stats["total_file_descriptors"] = totalFDs
	stats["status_counts"] = statusCounts

	// Find processes with highest resource usage
	topCPU, err := pm.GetTopProcesses(ctx, "cpu", 5)
	if err == nil {
		stats["top_cpu_processes"] = topCPU
	}

	topMemory, err := pm.GetTopProcesses(ctx, "memory", 5)
	if err == nil {
		stats["top_memory_processes"] = topMemory
	}

	return stats, nil
}

// Reset resets the process monitor state
func (pm *ProcessMonitor) Reset() {
	pm.processCache = make(map[int32]*process.Process)
	pm.lastUpdate = time.Time{}
	pm.isInitialized = false
}

// SearchProcesses searches for processes by name or command
func (pm *ProcessMonitor) SearchProcesses(ctx context.Context, query string, limit int) ([]models.ProcessInfo, error) {
	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to search processes: %w", err)
	}

	var matches []models.ProcessInfo
	query = strings.ToLower(query)

	for _, p := range processes {
		name, err := p.NameWithContext(ctx)
		if err != nil {
			continue
		}

		cmdline, err := p.CmdlineWithContext(ctx)
		if err != nil {
			cmdline = ""
		}

		// Check if query matches name or command
		if strings.Contains(strings.ToLower(name), query) ||
			strings.Contains(strings.ToLower(cmdline), query) {
			info, err := pm.getProcessInfo(ctx, p)
			if err != nil {
				continue
			}
			matches = append(matches, info)
		}
	}

	// Sort by CPU usage and limit results
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].CPUPercent > matches[j].CPUPercent
	})

	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
	}

	return matches, nil
}