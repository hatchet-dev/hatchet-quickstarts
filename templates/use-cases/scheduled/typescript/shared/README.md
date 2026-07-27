# Hatchet Scheduled Workflow Example

This is an example project demonstrating a scheduled Hatchet workflow in
TypeScript. The worker registers the workflow with a cron schedule through the
`on` cron option, so it runs every five minutes. You can also run it on demand.
For detailed setup instructions, see the [Hatchet Setup Guide](https://docs.hatchet.run/home/setup).

## Prerequisites

Before running this project, make sure you have the following:

1. [Node.js v16 or higher](https://nodejs.org/en/download)
{{- if eq .PackageManager "npm"}}
2. npm package manager (included with Node.js)
{{- else if eq .PackageManager "pnpm"}}
2. [pnpm](https://pnpm.io/installation) package manager
{{- else if eq .PackageManager "yarn"}}
2. [Yarn](https://yarnpkg.com/getting-started/install) package manager
{{- else if eq .PackageManager "bun"}}
2. [Bun](https://bun.sh/) runtime and package manager
{{- end}}

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
{{- if eq .PackageManager "npm"}}
npm install
{{- else if eq .PackageManager "pnpm"}}
pnpm install
{{- else if eq .PackageManager "yarn"}}
yarn install
{{- else if eq .PackageManager "bun"}}
bun install
{{- end}}
```

### Running an example

1. Start a Hatchet worker:

```bash
{{- if eq .PackageManager "npm"}}
npm run start
{{- else if eq .PackageManager "pnpm"}}
pnpm start
{{- else if eq .PackageManager "yarn"}}
yarn start
{{- else if eq .PackageManager "bun"}}
bun start
{{- end}}
```

The worker runs the task on its cron schedule while it is connected.

2. To run the task on demand, open a new terminal and run:

```bash
{{- if eq .PackageManager "npm"}}
npm run run:manual
{{- else if eq .PackageManager "pnpm"}}
pnpm run run:manual
{{- else if eq .PackageManager "yarn"}}
yarn run:manual
{{- else if eq .PackageManager "bun"}}
bun run run:manual
{{- end}}
```

This triggers the task on the worker running in the first terminal and prints the output to the second terminal.
