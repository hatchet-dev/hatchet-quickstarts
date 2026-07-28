import { hatchet } from '../hatchet-client';

type ScheduledInput = {
  message?: string;
};

type ScheduledOutput = {
  'scheduled-task': {
    message: string;
    ranAt: string;
  };
};

// The cron schedule registers when the worker starts, so the task runs every
// five minutes. It also runs on demand through src/run.ts.
export const scheduled = hatchet.workflow<ScheduledInput, ScheduledOutput>({
  name: 'scheduled-workflow',
  description: 'Runs every 5 minutes',
  on: {
    cron: '*/5 * * * *',
  },
});

scheduled.task({
  name: 'scheduled-task',
  // If a cron run carries no input the SDK passes null, so we return an
  // empty string to avoid null errors.
  fn: (input) => {
    console.log('scheduled task ran');

    return {
      message: input?.message ?? '',
      ranAt: new Date().toISOString(),
    };
  },
});
