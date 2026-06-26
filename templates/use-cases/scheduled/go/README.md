# Hatchet Scheduled Workflow Example

This is an example project demonstrating a scheduled Hatchet workflow in Go. The
worker registers the task with a cron schedule using `WithWorkflowCron`, so it
runs every five minutes. You can also run it on demand. For detailed setup
instructions, see the [Hatchet Setup Guide](https://docs.hatchet.run/home/setup).

## Prerequisites

Before running this project, make sure you have the following:

1. [Go v1.22 or higher](https://go.dev/doc/install)

## Setup

1. Clone the repository:

```bash
git clone https://github.com/hatchet-dev/hatchet-go-quickstart.git
cd hatchet-go-quickstart
```

2. Set the required environment variable `HATCHET_CLIENT_TOKEN` created in the [Getting Started Guide](https://docs.hatchet.run/home/hatchet-cloud-quickstart).

```bash
export HATCHET_CLIENT_TOKEN=<token>
```

> Note: If you're self hosting you may need to set `HATCHET_CLIENT_TLS_STRATEGY=none` to disable TLS

3. Install the project dependencies:

```bash
go mod tidy
```

### Running an example

1. Start a Hatchet worker:

```bash
go run cmd/worker/main.go
```

The worker runs the task on its cron schedule while it is connected.

2. To run the task on demand, in a new terminal:

```bash
go run cmd/run/main.go
```

This triggers the task on the worker running in the first terminal and prints the output to the second terminal.
