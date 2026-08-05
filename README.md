# monitor

A lightweight system monitoring CLI tool written in Go. Continuously collects and displays CPU, memory, network, and disk I/O metrics in the terminal with configurable polling intervals.

## Features

- **CPU** - usage percentage, physical/logical core count
- **Memory** - total, used, free, usage percentage
- **Network** - per-interface bandwidth (b/s, Kb/s, Mb/s) and packet rate (pps), with interface filtering
- **Disk** - per-device read/write throughput (B/s, KB/s, MB/s), physical disk auto-detection (Linux)
- Auto-refresh terminal display at configurable intervals
- Graceful shutdown on Ctrl+C
- Cross-platform: Windows, Linux, macOS (amd64 / arm64)

## Build

```bash
make all          # build for all platforms
make linux        # build for Linux only
make dev          # hot-reload dev mode (requires CompileDaemon)
make clean        # remove build artifacts
```

Binaries are output to the `build/` directory.

## Usage

```bash
# Monitor CPU, memory, and network with a 2-second interval
./monitor -c -m -n -i 2 run

# Monitor all metrics (CPU + memory + network + disk)
./monitor -a run

# Filter to a specific network interface
./monitor -n --if eth0 run

# Show current flag values without running
./monitor -c -m -n show

# Print version
./monitor version
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--cpu` | `-c` | `false` | Enable CPU metric |
| `--mem` | `-m` | `false` | Enable memory metric |
| `--net` | `-n` | `false` | Enable network metric |
| `--disk` | `-d` | `false` | Enable disk metric |
| `--all` | `-a` | `false` | Enable all metrics |
| `--interval` | `-i` | `2` | Polling interval in seconds |
| `--if` | | `""` | Network interface name filter (all interfaces if empty) |

## Install

```bash
make install      # copy binaries to ~/bucket/tools/monitor and /usr/local/bin/monitor
```

## Dependencies

- [cobra](https://github.com/spf13/cobra) - CLI framework
- [gopsutil](https://github.com/shirou/gopsutil) - system/process metrics
- [term](https://pkg.go.dev/golang.org/x/term) - terminal size detection

## License

See [LICENSE](LICENSE).
