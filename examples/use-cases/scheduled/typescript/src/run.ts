import { scheduled } from './workflows/scheduled-workflow';

async function main() {
  const res = await scheduled.run({
    message: 'hello from a manual run',
  });

  console.log(res['scheduled-task'].message);
}

if (require.main === module) {
  main().catch(console.error).finally(() => process.exit(0));
}
