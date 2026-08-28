# aDex Resource Usage

Measured figures for the packaged desktop application, so expectations are
based on observation rather than estimation.

## Measurement environment

| | |
|---|---|
| **Machine** | 12th Gen Intel Core i7-1265U (12 logical cores), 31 GB RAM |
| **OS** | Arch Linux, Wayland (Hyprland) |
| **Renderer** | WebKitGTK 2.52.6 / GTK 4.22.4 |
| **Build** | `wails3 task build` — production build, `-trimpath -ldflags="-w -s"` |
| **Wails** | v3.0.0-beta.14 |
| **Date** | 2026-08-27 |

Method: launch the packaged binary from a directory outside the repository,
wait 35 seconds for services to settle, then sample. Memory is the sum of
resident set size across the whole process tree. CPU is sampled with `top`
over multiple intervals — `ps` reports a cumulative average since process
start, which overstates a process that was briefly busy during boot.

## Disk

| | |
|---|---|
| **Binary** | ~26 MB (single self-contained executable, including 4 MB of embedded audio) |

The frontend is embedded in the binary, so there are no external asset files
to deploy. The application also writes a small settings file under
`~/.config/aDex/`.

## Memory

| | |
|---|---|
| **Total (process tree)** | ~337 MB |
| **Main process** | ~233 MB |
| **Processes** | 5 |

The five processes are the Go application itself plus the WebKit process
group it spawns — a network process, two sandbox helpers, and the web
process. This is typical for a WebKit-based desktop application: the bulk of
the memory belongs to the renderer, not the Go backend.

Terminal sessions add a shell process each. Memory grows modestly with the
number of open tabs and with terminal scrollback.

## CPU

| State | CPU (of one core) |
|---|---|
| **Window visible, idle** | ~23–50% |
| **Window hidden or minimised** | near zero |

Idle cost is dominated by the animated globe in the right column, which
renders continuously through `requestAnimationFrame`. Measured at ~64% of one
core before optimisation; the render loop now suspends when the window is not
visible, which removes the cost entirely when the application is in the
background.

On this 12-core machine, ~25% of a core is roughly 2% of total CPU capacity.
The figure is nonetheless higher than a mostly-static interface warrants, and
the globe is the reason.

### Reducing CPU further

- Minimise or switch away from the window: rendering stops.
- Themes without the globe panel (the `notype` and `fulltype` layout presets
  change which panels are shown) avoid the animation cost.

## Startup

Backend services initialise in well under a second on this machine —
colour schemes, then fonts (a scan of 1,343 system fonts completed in ~80 ms),
then network, filesystem and the terminal service. The boot splash sequence is
paced for effect rather than blocked on that work.

## Notes on measuring this yourself

Sample CPU with `top -b -n 3 -d 2 -p <pid>` and take the last reading, not
`ps -o pcpu`. `ps` reports the average over the process's entire lifetime, so
a process that used significant CPU during startup keeps reporting a high
figure long after it goes idle — this initially made the application look
busier at rest than it is.

Sum RSS across the process tree, not just the main process: the WebKit
processes hold most of the memory, so measuring only the parent understates
the real total by roughly a third.
