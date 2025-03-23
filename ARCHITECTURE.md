# Hibiscus Architecture

## Overview

Hibiscus is a modern terminal-based user interface for AWS services built in Go. It provides a keyboard-driven, intuitive interface for managing AWS resources directly from the terminal, offering an alternative to the AWS web console.

## Architectural Approach

Hibiscus follows a clean architecture pattern with clear separation of concerns:

1. **Terminal UI Layer**: Built with the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework, which implements the Model-View-Update (MVU) pattern
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
├── tui/                 # Terminal UI components
│   ├── aws/             # AWS service UI components
│   │   ├── ecr/         # ECR UI implementation
│   │   ├── elb/         # ELB UI implementation
│   │   └── route53/     # Route53 UI implementation
│   ├── main/            # Main application UI
│   └── styles/          # UI styling components
├── utils/               # Utility functions
├── main.go              # Application entry point
└── go.mod, go.sum       # Go module definitions
```

## Core Components

### Command Layer (cmd/)

The command layer is built with Cobra and provides the CLI interface for the application. The main command is defined in `cmd/root.go` which initializes the application and starts the Bubble Tea program.

### Terminal UI Layer (tui/)

The TUI layer is built with Bubble Tea and follows the Model-View-Update (MVU) pattern:

1. **Model**: Defines the application state
2. **View**: Renders the UI based on the model state
3. **Update**: Handles user input and updates the model state

Each AWS service has its own UI implementation in the `tui/aws/` directory, which includes:
- `model.go`: Defines the service-specific state
- `view.go`: Renders the service-specific UI
- `update.go`: Handles user input for the service
- `cmd.go`: Defines commands for the service (e.g., fetching data)
- `init.go`: Initializes the service UI components

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
3. The root command initializes the configuration and creates a new Bubble Tea program
4. The main TUI model is initialized with sub-models for each AWS service
5. The Bubble Tea program starts the event loop:
   - Renders the UI based on the current model state
   - Captures user input
   - Updates the model state based on user input
   - Rerenders the UI

Navigation follows a tab-based approach, where users can switch between different AWS services. Each service has its own view with tables, forms, and other UI components for interacting with the service's resources.

## Implementation Details

### Bubble Tea Components

Hibiscus uses several Bubble Tea components:
- `table`: For displaying tabular data (e.g., ECR repositories, Route53 records, ELB load balancers and listeners)
- `spinner`: For showing loading states when fetching data from AWS
- `textinput`: For text input fields

### AWS Integration

AWS integration is handled through the AWS SDK for Go, with credentials managed through AWS CLI profiles. The application supports switching between profiles and handles authentication transparently.

### Concurrency

The application uses Go's concurrency features (goroutines and channels) through Bubble Tea's command system to handle asynchronous operations like fetching data from AWS APIs without blocking the UI.

## Current Status and Future Direction

As of the current implementation, Hibiscus supports viewing resources for:
- Amazon ECR (Elastic Container Registry)
- Amazon ELB (Elastic Load Balancer)
- Amazon Route53 (DNS)

Future plans include:
- Adding support for more AWS services
- Implementing resource editing capabilities
- Adding more advanced filtering and search functionality 