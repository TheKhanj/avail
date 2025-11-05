# avail
`avail` is a lightweight HTTP site monitoring tool that tracks availability, latency, and health of configured websites. It exposes metrics through files, making it easy to integrate with other monitoring tools or custom scripts.

# Features
Monitor multiple HTTP sites periodically.
Output latency and health metrics as files.
Query raw HTTP responses and extract status, headers, or body.
JSON-based configuration with a strict schema.

# Installation
You can build or download the `avail` binary and place it in your `PATH`.

# Configuration
`avail` uses a JSON configuration file (default `avail.json`) to define the sites to monitor.
Example schema (simplified):

```json
{
  "$schema": "...",
  "sites": [
    {
      "title": "example",
      "url": "https://example.com",
      "interval": "60s"
    }
  ]
}
```

Metrics are updated in:
```
/var/run/avail/{host}/latency
/var/run/avail/{host}/health
```

# Usage
Run the daemon

`avail run [-c config.json]`

# Check status
`avail status [title...]`

# List monitored sites
`avail list`

# HTTP Response Commands
`avail` can read a raw HTTP response from the file set in AVAIL_HTTP and extract parts of it. This is especially useful if you want to implement custom logic for determining the availability of a site based on the HTTP response.

`avail http <command> [arguments]`

## Available commands:
- `status`: get HTTP status line
- `header` <name>: get a specific header value
- `body`: get response body

Examples:
```bash
avail http status
avail http header content-type
avail http body
```

# Schema Command
`avail schema`

Displays the JSON schema used for configuration.

# File-based Metrics
`avail` stores metrics in a per-user or system-wide directory depending on the environment. The base directory is determined as follows:

## Linux / macOS
- Root user: `/var/run/avail/{host}/`
- Non-root user: `/var/run/user/{uid}/avail/{host}/`

## Windows
`%LOCALAPPDATA%\avail\{host}\` or fallback to the temp directory

# License
MIT License
