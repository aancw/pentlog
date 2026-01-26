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

Most commands need to know the current engagement details (Client, Phase, etc.). You can load this from the centralized config manager:

```go
import (
    "pentlog/pkg/config"
    "fmt"
)

func runMyCommand(cmd *cobra.Command, args []string) {
    mgr := config.Manager()
    ctx, err := mgr.LoadContext()
    if err != nil {
        fmt.Println("Error: Not in an active engagement. Run 'pentlog create' first.")
        return
    }

    fmt.Printf("Current Client: %s\n", ctx.Client)
}
```

### Using the Enhanced Error System

All error messages should use the `pkg/errors` package for consistency and user guidance. This transforms generic errors into actionable messages with reasons and solutions.

#### Available Error Types

The package provides 18 error types covering common failures:

- **Context**: `NoActiveContext`, `ContextNotFound`, `InvalidContext`
- **Sessions**: `SessionNotFound`, `SessionCrashed`, `AlreadyInSession`
- **Dependencies**: `DependencyMissing`, `DependencyVersionMismatch`
- **Database**: `DatabaseError`, `DatabaseLocked`
- **Files**: `FileNotFound`, `DirectoryNotFound`, `PermissionDenied`
- **Configuration**: `AIConfigMissing`, `AIConfigInvalid`
- **Archives**: `ArchiveNotFound`, `ArchiveCorrupted`, `ArchiveEncrypted`
- **Generic**: `Generic` (for other errors)

#### Error Usage Examples

**Example 1: Missing Context (Fatal)**
```go
import "pentlog/pkg/errors"

func runMyCommand(cmd *cobra.Command, args []string) {
    mgr := config.Manager()
    ctx, err := mgr.LoadContext()
    if err != nil {
        errors.NoContext().Fatal()
    }
    // ...
}
```

**Example 2: Database Error (Non-Fatal)**
```go
import "pentlog/pkg/errors"

sessions, err := logs.ListSessions()
if err != nil {
    errors.DatabaseErr("list_sessions", err).Print()
    return
}
```

**Example 3: File Operation Error**
```go
import "pentlog/pkg/errors"

data, err := os.ReadFile(path)
if err != nil {
    errors.FileErr(path, err).Print()
    return
}
```

**Example 4: Custom Error with Details**
```go
import "pentlog/pkg/errors"

errors.NewError(errors.SessionNotFound, "Session 123 not found").
    AddReason("Session file was deleted").
    AddSolution("Run: pentlog recover").
    Print()
```

#### Helper Functions

For common scenarios, use pre-built helpers:

```go
errors.NoContext()                      // Missing engagement context
errors.ContextMissing()                 // Context file not found
errors.SessionMissing(sessionID)        // Session not found
errors.SessionCrashedError(sessionID)   // Session crashed
errors.AlreadyInShell()                 // Already in pentlog shell
errors.MissingDependency(tool, cmd)     // Missing dependency
errors.DatabaseErr(op, err)             // Database error
errors.DatabaseLockedErr()              // Database locked
errors.FileErr(path, err)               // File error (auto-detects type)
errors.DirErr(path, err)                // Directory error (auto-detects type)
errors.AIConfigErr()                    // AI config missing
errors.ArchiveErr(path, err)            // Archive operation error
errors.ArchivePasswordErr(path)         // Archive encrypted
```

#### Output Example

```
❌ Error: No active engagement context found.

This can happen when:
  1. You haven't started an engagement yet
  2. The context was reset with 'pentlog reset'
  3. Running pentlog from a different user account

To fix:
  $ pentlog create   # Start a new engagement
  $ pentlog switch   # Switch to an existing context
```

See [ERROR_HANDLING_GUIDE.md](ERROR_HANDLING_GUIDE.md) for detailed documentation.

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
3.  **Use enhanced error messages** with `pkg/errors` for any user-facing errors (no generic error strings).
4.  Ensure the test suite passes (`go test ./...`).
5.  Make sure your code lints (`golangci-lint run`).
6.  Test error messages manually to ensure they're helpful:
    ```bash
    go build -o pentlog main.go
    # Test commands that trigger errors
    ./pentlog <command> # Verify error output is clear and actionable
    ```
7.  Issue that pull request!

## Error Message Guidelines

When handling errors in commands:

- ✅ **DO** use `pkg/errors` helpers for user-facing errors
- ✅ **DO** provide context in error messages (why it happened)
- ✅ **DO** include actionable solutions (how to fix)
- ❌ **DON'T** use generic `fmt.Fprintf(os.Stderr, "Error: %v")` for user-visible errors
- ❌ **DON'T** expose raw Go error messages without context

Example of good error handling:
```go
ctx, err := mgr.LoadContext()
if err != nil {
    errors.NoContext().Fatal()  // ✅ Clear, actionable, helpful
}

// Instead of:
// fmt.Fprintf(os.Stderr, "Error: %v\n", err)  // ❌ Generic
```
