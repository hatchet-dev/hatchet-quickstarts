import { hatchet } from './hatchet-client';
import { scheduled } from './workflows/scheduled-workflow';

async function main() {
  const worker = await hatchet.worker('scheduled-worker', {
    workflows: [scheduled],
  });

  await worker.start();
}

if (require.main === module) {
  main();
}
