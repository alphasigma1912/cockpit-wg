# Resource Usage

This project aims to minimize resource consumption while monitoring WireGuard interfaces.

- Metrics sampling is timer-driven with a minimum interval of 2 seconds to avoid busy loops.
- Ring buffers hold a bounded history, automatically discarding the oldest samples.
- WireGuard state uses `wgctrl` instead of spawning external commands, except for `systemctl` or package-management tasks.
- UI refreshes are debounced to roughly 500ms to prevent excessive updates.
- Optional profiling is available by building with the `pprof` tag:
  ```
  go build -tags pprof
  ```
  A pprof server will listen on `localhost:6060` when enabled.

Resource targets on typical systems:

- Idle CPU usage near 0%
- Metrics sampling below 1% CPU
- Resident memory under 50MB
