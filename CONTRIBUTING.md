# Contributing to Pentlog

First off, thanks for taking the time to contribute! 🎉

The following is a set of guidelines for contributing to `pentlog`. These are mostly guidelines, not rules. Use your best judgment, and feel free to propose changes to this document in a pull request.

## Development Setup

### Prerequisites

- **Go**: Version 1.24.0 or higher.
- **Git**: For version control.
- **golangci-lint**: For linting the code.

### Installation

1.  **Clone the repository**:
    ```bash
    git clone https://github.com/aancw/pentlog.git
    cd pentlog
    ```

2.  **Download dependencies**:
    ```bash
    go mod download
    ```

## Project Structure

A high-level overview of the codebase:

- **`cmd/`**: Contains the main application entry points and Cobra command definitions. Each command (e.g., `pentlog create`) typically has its own file here (e.g., `cmd/create.go`).
- **`pkg/`**: Contains the core library code.
    - **`metadata/`**: Handles engagement context (client, phase, etc.).
    - **`vulns/`**: Logic for managing vulnerabilities and findings.
    - **`tui/`**: User interface components.
    - **`utils/`**: Helper functions.

## Implementation Guide

### Adding a New Command

`pentlog` uses [Cobra](https://github.com/spf13/cobra) for its CLI interface. To add a new command:

1.  **Create a new file** in `cmd/` (e.g., `cmd/mycommand.go`).

2.  **Define the Command**:
    ```go
    package cmd

    import (
        "fmt"
        "github.com/spf13/cobra"
    )

    // Define flags if needed
    var myFlag string

    var myCmd = &cobra.Command{
        Use:   "mycommand",
        Short: "A brief description of what mycommand does",
        Run: func(cmd *cobra.Command, args []string) {
            // Your implementation logic here
            fmt.Println("Hello from mycommand!")
        },
    }

    func init() {
        // Register flags
        myCmd.Flags().StringVarP(&myFlag, "flag", "f", "", "Description for flag")

        // Add to root command
        rootCmd.AddCommand(myCmd)
    }
    ```

### Using Engagement Context

Most commands need to know the current engagement details (Client, Phase, etc.). You can load this from the global context:

```go
import (
    "pentlog/pkg/metadata"
    "fmt"
)

func runMyCommand(cmd *cobra.Command, args []string) {
    ctx, err := metadata.Load()
    if err != nil {
        fmt.Println("Error: Not in an active engagement. Run 'pentlog create' first.")
        return
    }

    fmt.Printf("Current Client: %s\n", ctx.Client)
}
```

## Building and Testing

### Build

To build the `pentlog` binary:

```bash
go build -o pentlog main.go
```

### Test

To run the test suite:

```bash
go test ./... -v
```

### Lint

Ensure your code follows the project's style guidelines:

```bash
golangci-lint run
```

## Pull Request Process

1.  Fork the repo and create your branch from `main`.
2.  If you've added code that should be tested, add tests.
3.  Ensure the test suite passes.
4.  Make sure your code lints.
5.  Issue that pull request!
