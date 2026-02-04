# pulumi-rollback

A CLI tool for rolling back Pulumi stack deployments to previous states from the deployment history.

## Features

- **List deployment history** - View all past deployments with version numbers, timestamps, and results
- **Preview rollbacks** - See what changes would be made before executing
- **Execute rollbacks** - Roll back to any previous deployment version
- **Multi-backend support** - Works with Pulumi Cloud, S3, Azure Blob, GCS, and local filesystem

## Installation

### From Source

```bash
go install github.com/PegasusHeavyIndustries/pulumi-rollback@latest
```

### Build Locally

```bash
git clone https://github.com/pegasusheavy/pulumi-rollback.git
cd pulumi-rollback
go build -o pulumi-rollback .
```

## Usage

### List Deployment History

```bash
# List all deployment history for a stack
pulumi-rollback list --stack mystack

# List last 10 deployments
pulumi-rollback list --stack mystack --limit 10
```

### Preview a Rollback

```bash
# Preview what would change when rolling back to version 5
pulumi-rollback preview --stack mystack --version 5
```

### Execute a Rollback

```bash
# Roll back to version 5 (with confirmation prompt)
pulumi-rollback to --stack mystack --version 5

# Roll back without confirmation
pulumi-rollback to --stack mystack --version 5 --yes
```

### Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--stack` | `-s` | Name of the Pulumi stack |
| `--cwd` | `-C` | Path to the Pulumi project directory (default: `.`) |
| `--verbose` | `-v` | Enable verbose output |

## How It Works

1. **List**: Queries the Pulumi stack history using the Automation API
2. **Preview**: Temporarily imports the target state and runs a preview to show changes
3. **Rollback**: Imports the target state, refreshes to reconcile with actual infrastructure, and runs `up` to apply changes

## Requirements

- Go 1.25 or later
- Pulumi CLI installed and configured
- Access to the Pulumi backend where your stacks are stored

## License

Copyright 2026 Pegasus Heavy Industries LLC

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.

## Contact

- Email: pegasusheavyindustries@gmail.com
- Patreon: https://www.patreon.com/c/PegasusHeavyIndustries
