# Hibiscus Architecture

## Overview

Hibiscus is a modern, keyboard-first terminal UI for AWS resources built in Go. The project recently migrated from Bubble Tea to [tview](https://github.com/rivo/tview), giving us richer layout control, native widgets, and a unified application shell that coordinates every AWS surface. The primary CLI (`cmd/root.go`) now launches this tview experience, while a standalone Route53 PoC (`cmd/route53tview`) remains available for record editing experiments.

## Architectural Approach

The system follows a clean, layered structure:

1. **Terminal UI Layer** – Built with tview components hosted inside a shared shell (`tviewapp/hibiscus`). This layer owns navigation (command mode `:`, filter mode `/`, refresh `r/R`, Esc backtracking), focus management, and shared status/error bars.
2. **CLI Layer** – Uses Cobra to parse CLI options, select AWS profiles, and wire up service factories before handing control to the UI layer.
3. **AWS Integration Layer** – Located under `internal/aws/<service>`, encapsulating SDK clients, pagination, and domain helpers (e.g., Route53 alias detection). Recent work added full pagination for Route53 records and ECR images to ensure every resource appears in the UI.
4. **Configuration Layer** – `config/` persists lightweight state such as the last active tab via `config.SetTabKey` and loads AWS profile information.

## Folder Structure

```
hibiscus/
├── cmd/                 # CLI command definitions
│   └── root.go          # Main application command
├── config/              # Configuration management
│   ├── config.go        # Config structures and functions
│   └── constant.go      # Configuration constants
├── docs/                # Documentation and assets
├── internal/            # Internal implementation code
│   └── aws/             # AWS service implementations
│       ├── ecr/         # ECR service implementation
│       ├── elbv2/       # ELB (Elastic Load Balancer) service implementation
│       ├── route53/     # Route53 service implementation
│       └── aws_common.go# Common AWS functionality
├── tviewapp/            # Terminal UI components (tview)
│   ├── hibiscus/        # Shared shell, layout, nav modes
│   │   └── services/    # Service-specific UI packages
│   │       ├── ecr/
│   │       ├── route53/
│   │       └── elb/
│   └── route53/         # Standalone proof-of-concept with edit modals
├── tui/                 # Legacy Bubble Tea UI kept for reference
├── utils/               # Utility functions
├── main.go              # Application entry point
└── go.mod, go.sum       # Go module definitions
```

## Core Components

### Command Layer (cmd/)

The command layer is built with Cobra and provides the CLI interface for the application. The main command is defined in `cmd/root.go`, where it initializes configuration, builds the service list, and boots the shared tview application shell.

### Terminal UI Layer (tviewapp/hibiscus)

The production UI is entirely tview-based. The `hibiscus.App` shell handles layout, global keybindings, and service switching. Notable capabilities added during the migration include:

- **Mode-aware shortcuts** – When a service enters filter mode (`/`), global keys such as `r` (refresh) and `:` (command palette) are temporarily suppressed until Esc exits the filter, preventing accidental reloads while typing.
- **Focus isolation** – Each service tracks whether it is active; background data refreshes no longer steal focus from the visible view. Empty tables (e.g., an ECR repo with zero images) render placeholder rows that keep keyboard focus anchored.
- **Command palette** – Typing `:` opens a centered palette listing all services. Enter now selects the highlighted suggestion even if the typed text is only a prefix, speeding up navigation (`:r` then Enter jumps to Route53).

Concrete services live under `tviewapp/hibiscus/services/<service>`:

- **ECR**: repositories → images, clipboard shortcuts, and full image pagination.
- **Route53**: hosted zones → records with alias annotations plus record pagination.
- **ELB**: load balancers → listeners → rules with summarized conditions/actions.

Legacy Bubble Tea code (`tui/`) is retained for reference but not invoked.

### AWS Integration Layer (internal/aws/)

This layer handles the actual AWS API calls and business logic, separated from the UI concerns. Each AWS service has its own implementation directory under `internal/aws/`.

### Configuration Layer (config/)

Manages application configuration using a global singleton pattern with mutex protection for thread safety. Handles settings such as:

- Current AWS profile
- Current UI tab
- Other application settings

## Application Flow

1. User starts the application with `hibiscus` or `hibiscus --profile <aws-profile>`
2. `main.go` calls `cmd.Execute()` to start the application
3. The root command initializes configuration and constructs the tview application with service factories
4. Each service kicks off its initial AWS fetch asynchronously, scheduling redraws via `Application.QueueUpdateDraw`
5. The tview event loop renders the active service, while global keybindings (`:`, `/`, `Esc`, `R`) are intercepted by the shell. Focus is only granted to the currently visible service; background refreshes are queued via `Application.QueueUpdateDraw`.

Navigation is hierarchical: the command palette swaps between services, `Enter` drills into child resources (e.g., hosted zones → records), and `Esc` returns to the previous table.

## Implementation Details

### tview Components

Hibiscus relies on core tview primitives:

- `Table`: renders paginated AWS data sets with keyboard selection
- `InputField`: powers filter mode (`/`) and the `:` command palette
- `Pages` / `Flex`: swap between hierarchy levels (repos → images, etc.) without rebuilding widgets
- `Grid`: centers overlays such as the command palette and future modals

### AWS Integration

The AWS SDK for Go v2 powers every service. Clients are lazily constructed per package (e.g., `internal/aws/route53`) and reused. Long lists are paginated: both `ListResourceRecordSets` and `DescribeImages` iterate on `IsTruncated/NextToken` so the UI always renders the full dataset.

### Concurrency

Each service performs network requests in goroutines and schedules UI mutations via `Application.QueueUpdateDraw`. This keeps the TUI responsive, even when multiple services refresh in parallel. Focus helpers ensure those callbacks never override the active widget.

## Adding a New Hibiscus Service

1. **Create the AWS client package (if needed)**
   - Add an SDK helper under `internal/aws/<service>/`.
   - Encapsulate pagination, sorting, and any domain helpers so the UI can remain dumb.

2. **Implement the `hibiscus.Service` contract**
   - Create `tviewapp/hibiscus/services/<service>/service.go`.
   - Build your layout (usually a `Flex` that stacks a filter `InputField` and a `Pages` host with one `Table` per hierarchy level).
   - Track state slices (raw vs. filtered data) and expose methods:
     - `Init` – kick off initial AWS fetches in goroutines; call `ctx.SetStatus`/`SetError` appropriately.
     - `Activate`/`Deactivate` – toggle an `active` flag and call a helper that restores focus to the correct table. Only set focus when `active` to avoid stealing input while hidden.
     - `Refresh` – reload the relevant subset based on the current tab/selection.
     - `EnterFilterMode` – focus the filter input and return `true` so the shell knows to pause global shortcuts.
     - `HandleInput` – react to Esc/Enter and any service-specific shortcuts (copy, edit, etc.).

3. **Register the service**
   - In `cmd/root.go`, append a factory to the slice passed into `hibiscus.New`. Order determines palette listing and the default tab.
   - Optionally add `config` constants for persisting the new tab index.

4. **Update documentation and README**
   - Mention the new surface in `README.md` and this architecture guide.
   - Describe any new shortcuts or required IAM permissions.

5. **Test manually**
   - Run `go build ./...` to ensure compilation.
   - Launch `go run ./cmd/hibiscus --profile <profile>` and verify navigation, filter mode, Esc behavior, and status messaging for the new service.

Following this pattern keeps services decoupled while letting the shell provide a consistent navigation experience.
