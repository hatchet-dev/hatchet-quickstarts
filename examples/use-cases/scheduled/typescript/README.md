# Hatchet Scheduled Workflow Example

This is an example project demonstrating a scheduled Hatchet workflow in
TypeScript. The worker registers the workflow with a cron schedule through the
`on` cron option, so it runs every five minutes. You can also run it on demand.
For detailed setup instructions, see the [Hatchet Setup Guide](https://docs.hatchet.run/home/setup).

## Prerequisites

Before running this project, make sure you have the following:
1. [Node.js 22 or later](https://nodejs.org/en/download)
2. [pnpm](https://pnpm.io/installation) package manager

## Setup

1. Create the project using the Hatchet CLI:

```bash
hatchet quickstart --use-case scheduled --language typescript
```

2. Set the required environment variable `HATCHET_CLIENT_TOKEN` created in the [Getting Started Guide](https://docs.hatchet.run/home/hatchet-cloud-quickstart).

```bash
export HATCHET_CLIENT_TOKEN=<token>
```

> Note: If you're self hosting you may need to set `HATCHET_CLIENT_TLS_STRATEGY=none` to disable TLS

3. Install the project dependencies:

```bash
pnpm install
```

### Running an example

1. Start a Hatchet worker:

```bash
pnpm start
```

The worker runs the task on its cron schedule while it is connected.

2. To run the task on demand, open a new terminal and run:

```bash
pnpm run run:manual
```

This triggers the task on the worker running in the first terminal and prints the output to the second terminal.
