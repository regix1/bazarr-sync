<div align="center">

### 🎯 Bulk sync subtitles in Bazarr with ease

[![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?style=for-the-badge&logo=docker&logoColor=white)](https://ghcr.io/regix1/bazarr-sync)
[![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green?style=for-the-badge)](LICENSE)

</div>

---

## 📖 Why bazarr-sync?

Bazarr downloads subtitles automatically, but syncing them requires clicking through each file individually. This tool lets you sync everything at once - movies, shows, or both - with smart caching and scheduling.

```
┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐
│                 │      │                 │      │                 │
│     Bazarr      │ API  │  bazarr-sync    │      │   Subtitles     │
│     Server      ├─────►│     Engine      ├─────►│    Synced!      │
│                 │      │                 │      │                 │
└─────────────────┘      └─────────────────┘      └─────────────────┘
```

---

## 🚀 Quick Start

### Docker (Recommended)

```bash
# Pull the image
docker pull ghcr.io/regix1/bazarr-sync:latest

# Create config file
nano config.yaml

# Run sync with cache
docker run -it --rm \
  -v ${PWD}/config.yaml:/config/config.yaml \
  -v ${PWD}/cache:/config/cache \
  ghcr.io/regix1/bazarr-sync:latest \
  sync movies --use-cache
```

### Docker Compose

```yaml
services:
  bazarr-sync:
    image: ghcr.io/regix1/bazarr-sync:latest
    container_name: bazarr-sync
    volumes:
      - ./config.yaml:/config/config.yaml
      - ./cache:/config/cache
    command: --schedule  # Run on schedule
    restart: unless-stopped
```

---

## ⚙️ Configuration

Create a `config.yaml` file:

```yaml
# ┌─────────────────────────────────────────────────────────────┐
# │                    BASIC CONNECTION                         │
# └─────────────────────────────────────────────────────────────┘
Address: localhost        # Or container name if using Docker
Port: 6767               # Default Bazarr port
Protocol: http           # http or https
ApiToken: your_token     # From Bazarr Settings > General

# ┌─────────────────────────────────────────────────────────────┐
# │                    SCHEDULING (Optional)                    │
# └─────────────────────────────────────────────────────────────┘
Schedule:
  Enabled: true
  SyncMovies: true
  SyncShows: true
  CronExpression: "0 1 * * 0"    # Weekly on Sunday at 1 AM
  Timezone: "America/Chicago"

# ┌─────────────────────────────────────────────────────────────┐
# │                    CACHE (Optional)                         │
# └─────────────────────────────────────────────────────────────┘
Cache:
  Enabled: true
  MoviesCache: "/config/cache/movies"
  ShowsCache: "/config/cache/shows"

# ┌─────────────────────────────────────────────────────────────┐
# │                    SYNC OPTIONS (Optional)                  │
# └─────────────────────────────────────────────────────────────┘
SyncOptions:
  GoldenSection: false      # Use Golden Section Search
  NoFramerateFix: true     # Skip framerate correction
```

---

## 📚 Commands

### Basic Sync Operations

```bash
┌─────────────────────────────────────────────────────────────┐
│ SYNC ALL MOVIES                                            │
├─────────────────────────────────────────────────────────────┤
│ $ bazarr-sync sync movies                                  │
│                                                             │
│ Output:                                                     │
│ [1/155] PROCESSING: Inception (2 subtitles)                │
│   └─ SYNCING [en]: ⠋ (animated spinner)                    │
│   └─ SYNCING [en]: ✓ Success                              │
│   └─ SYNCING [es]: ✓ Already in sync                      │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ SYNC ALL TV SHOWS                                          │
├─────────────────────────────────────────────────────────────┤
│ $ bazarr-sync sync shows                                   │
│                                                             │
│ Output:                                                     │
│ [1/45] PROCESSING: Breaking Bad (62 episodes)              │
│   └─ SYNCING [S01E01 - en]: ✓ Success                     │
└─────────────────────────────────────────────────────────────┘
```

### Advanced Features

```bash
┌─────────────────────────────────────────────────────────────┐
│ LIST ALL MEDIA WITH IDs                                    │
├─────────────────────────────────────────────────────────────┤
│ $ bazarr-sync sync movies --list                           │
│                                                             │
│ Title                                          RadarrId    │
│ ──────────────────────────────────────────────────────     │
│ Inception                                      123         │
│ The Matrix                                     456         │
│ Interstellar                                   789         │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ SYNC SPECIFIC ITEMS ONLY                                   │
├─────────────────────────────────────────────────────────────┤
│ $ bazarr-sync sync movies --radarr-id 123,456             │
│ $ bazarr-sync sync shows --sonarr-id 789,012              │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ USE SMART CACHE (Skip already synced)                      │
├─────────────────────────────────────────────────────────────┤
│ $ bazarr-sync sync movies --use-cache                      │
│                                                             │
│ Output:                                                     │
│   └─ CACHED [en]: Already synced                          │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ CONTINUE FROM INTERRUPTION                                 │
├─────────────────────────────────────────────────────────────┤
│ $ bazarr-sync sync movies --continue-from 456              │
│                                                             │
│ # Skips all movies before ID 456 and continues            │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ CANCEL RUNNING SYNC                                        │
├─────────────────────────────────────────────────────────────┤
│ $ bazarr-sync cancel                                       │
│                                                             │
│ Output:                                                     │
│ 🛑 Sent cancel signal to sync process (PID: 1234)         │
│ ✅ Cancel signal sent. The sync will stop gracefully.      │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ RUN ON SCHEDULE                                            │
├─────────────────────────────────────────────────────────────┤
│ $ bazarr-sync --schedule                                   │
│                                                             │
│ Output:                                                     │
│ INFO  Scheduler started. Next sync: 2025-01-21 01:00 CST  │
│ INFO  Schedule: 0 1 * * 0 (Timezone: America/Chicago)     │
└─────────────────────────────────────────────────────────────┘
```

### Command Options

```bash
┌─────────────────────────────────────────────────────────────┐
│ GLOBAL FLAGS                                               │
├─────────────────────────────────────────────────────────────┤
│ --config <file>      │ Config file path (default: ./config.yaml)
│ --schedule           │ Run on schedule defined in config
│ --run-initial        │ Run sync immediately when scheduler starts
│ --help              │ Show help information
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ SYNC FLAGS                                                 │
├─────────────────────────────────────────────────────────────┤
│ --list              │ List all media with IDs
│ --use-cache         │ Skip already synced subtitles
│ --golden-section    │ Use Golden Section Search algorithm
│ --no-framerate-fix  │ Skip framerate correction
│ --verbose           │ Show detailed error messages
│ --continue-from <id>│ Resume from specific movie/episode ID
│ --radarr-id <ids>   │ Sync specific movies (comma-separated)
│ --sonarr-id <ids>   │ Sync specific shows (comma-separated)
└─────────────────────────────────────────────────────────────┘
```

---

## 📊 Statistics

After each sync run, you'll see a summary:

```
────────────────────────────────────────────────────────────
Sync completed:
  ✅ 42 newly synced
  ✓  108 already in sync  
  ⏭️  23 skipped (cached/embedded)
  ❌ 5 failed

💡 Tip: Run with --verbose to see detailed error messages
```

---

## 🎯 Features

| Feature | Description |
|---------|-------------|
| **🔄 Bulk Sync** | Process entire library at once |
| **💾 Smart Cache** | Skip already synced files automatically |
| **⏰ Scheduler** | Set up automatic weekly/daily syncs |
| **⏸️ Resume Support** | Continue after interruption |
| **🎨 Progress Tracking** | Visual feedback with animated spinners |
| **🎯 Selective Sync** | Choose specific movies/shows |
| **🛑 Cancel Command** | Gracefully stop running operations |
| **📝 Verbose Mode** | Detailed error messages for debugging |

---

## 📋 Requirements

- ✅ Bazarr instance with API access enabled
- ✅ Subsync installed in Bazarr (Settings > Subtitles > Synchronization)
- ✅ Docker or Go runtime

---

## 🐛 Troubleshooting

| Issue | Solution |
|-------|----------|
| **401 Unauthorized** | Check your API token in config.yaml |
| **Connection refused** | Verify Bazarr address and port |
| **Sync failures** | Ensure subsync is installed in Bazarr |
| **Already synced** | Use `--use-cache` to skip them |

---

## 📄 License

MIT License - See [LICENSE](LICENSE) file for details

---

<div align="center">

**Made with ❤️ for the Bazarr community**

[Report Bug](https://github.com/regix1/bazarr-sync/issues) · [Request Feature](https://github.com/regix1/bazarr-sync/issues)

</div>
```

This README uses line art boxes and visual separators to make it more readable and organized. The command examples show actual output to help users understand what to expect. The layout is clean, professional, and easy to navigate.
