# Hatchet Scheduled Workflow Example

This is an example project demonstrating a scheduled Hatchet workflow in
Python. The worker registers the task with a cron schedule through `on_crons`,
so it runs every five minutes. You can also run it on demand. For detailed
setup instructions, see the [Hatchet Setup Guide](https://docs.hatchet.run/home/setup).

## Prerequisites

Before running this project, make sure you have the following:

1. [Python v3.10 or higher](https://www.python.org/downloads/)
2. [Poetry](https://python-poetry.org/docs/#installation) for dependency management

## Setup

1. Create the project using the Hatchet CLI:

```bash
hatchet quickstart --use-case scheduled --language python
```

2. Set the required environment variable `HATCHET_CLIENT_TOKEN` created in the [Getting Started Guide](https://docs.hatchet.run/home/hatchet-cloud-quickstart).

```bash
export HATCHET_CLIENT_TOKEN=<token>
```

> Note: If you're self hosting you may need to set `HATCHET_CLIENT_TLS_STRATEGY=none` to disable TLS

3. Install the project dependencies:

```bash
poetry install
```

### Running an example

1. Start a Hatchet worker:

```shell
poetry run python src/worker.py
```

The worker runs the task on its cron schedule while it is connected.

2. To run the task on demand, open a new terminal and run:

```shell
poetry run python src/run.py
```

This triggers the task on the worker running in the first terminal and prints the output to the second terminal.
