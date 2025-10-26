# Hibiscus Architecture

## Overview

Hibiscus is a modern terminal-based user interface for AWS services built in Go. It provides a keyboard-driven, intuitive interface for managing AWS resources directly from the terminal, offering an alternative to the AWS web console.

## Architectural Approach

Hibiscus follows a clean architecture pattern with clear separation of concerns:

1. **Terminal UI Layer**: Built with the [tview](https://github.com/rivo/tview) framework for immediate-mode terminal widgets, composed inside a shared shell
2. **CLI Layer**: Uses [Cobra](https://github.com/spf13/cobra) for command-line interface functionality
3. **AWS Integration Layer**: Handles AWS API calls and business logic
4. **Configuration Layer**: Manages application and AWS configuration

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

The production UI is implemented with tview. The `tviewapp/hibiscus` package owns:

1. A global shell (`App`) that lays out header, status, error bars, and an interchangeable content area.
2. A lightweight command palette bound to `:` that lets users jump directly to `ecr`, `route53`, or `elb`.
3. A `Service` interface so each AWS surface can supply its own primitives, filter handling, refresh logic, and keybindings.

Concrete services live under `tviewapp/hibiscus/services/<service>` and compose the shared infrastructure:

- **ECR**: repository table → image table with copy-to-clipboard helpers
- **Route53**: hosted zone table → record table with smart alias rendering
- **ELB**: load balancer table → listeners → rules

Legacy Bubble Tea code remains in `tui/` for historical context but is no longer wired into the CLI.

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
5. The tview event loop renders the active service, while global keybindings (`:`, `/`, `Esc`, `R`) are intercepted by the shell

Navigation is hierarchical: the command palette swaps between services, `Enter` drills into child resources (e.g., hosted zones → records), and `Esc` climbs back up one level.

## Implementation Details

### tview Components

Hibiscus relies on core tview primitives:

- `Table`: renders paginated AWS data sets with keyboard selection
- `InputField`: powers filter mode (`/`) and the `:` command palette
- `Pages` / `Flex`: swap between hierarchy levels (repos → images, etc.) without rebuilding widgets
- `Grid`: centers overlays such as the command palette and future modals

### AWS Integration

AWS integration is handled through the AWS SDK for Go, with credentials managed through AWS CLI profiles. The application supports switching between profiles and handles authentication transparently.

### Concurrency

Each service fetches AWS data inside goroutines and marshals UI updates through `Application.QueueUpdateDraw`, ensuring the terminal remains responsive while API calls are in flight.

## Current Status and Future Direction

As of the current implementation, Hibiscus supports viewing resources for:

- Amazon ECR (Elastic Container Registry)
- Amazon ELB (Elastic Load Balancer)
- Amazon Route53 (DNS)

Future plans include:

- Adding support for more AWS services
- Implementing resource editing capabilities
- Adding more advanced filtering and search functionality
