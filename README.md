# logviewer

A terminal-based interactive log viewer built with [Bubble Tea](https://github.com/charmbracelet/bubbletea). It connects to remote servers via SSH, fetches and parses structured JSON logs (plain or gzipped), and displays them.

---

## Installation

To build and install the `logviewer` binary:

```bash
make           # Builds the binary and installs it to /usr/local/bin
````

---

## Usage

```bash
logviewer
```

Once the viewer starts:

* `1` → Filter logs by origin: **email**
* `2` → Filter logs by origin: **nextjs**
* `0` → Clear origin filter
* `i` → Filter logs by level: **info**
* `e` → Filter logs by level: **error**
* `f` → Filter logs by level: **fatal**
* `c` → Clear level filter
* `q` → Quit the viewer

Use arrow keys or `j/k` to scroll through log entries.

---

## Development

### Common Make Targets:

* `make` — Build and install the binary to `/usr/local/bin`
* `make run` — Build and run the app locally
* `make clean` — Remove the built binary and uninstall from system

---

## Dependencies

This project uses the following Charmbracelet libraries:

* [Bubble Tea](https://github.com/charmbracelet/bubbletea) — The TUI engine
* [Bubbles](https://github.com/charmbracelet/bubbles) — UI components
* [Lipgloss](https://github.com/charmbracelet/lipgloss) — Layout and styling

---

## Log Format

The application expects each log line to be a valid JSON object (either plain text or gzipped).

Example log line:

```json
{
  "level": "error",
  "timestamp": "2025-07-28 12:00:00",
  "message": "Failed to send email"
}
```

Fields like `level`, `timestamp`, and `message` are always shown at the top. All additional fields are shown below, automatically aligned and formatted (including multi-line support for stacktraces).

---

## Configuration

Make sure your `.env` file or configuration mechanism used by `LoadConfig()` sets the necessary SSH and log-related values:

* `SSH_USER`
* `SSH_KEY`
* `SSH_HOST`
* `EMAIL_PORT`, `EMAIL_LOG_PATH`
* `NEXTJS_PORT`, `NEXTJS_LOG_DIR`
