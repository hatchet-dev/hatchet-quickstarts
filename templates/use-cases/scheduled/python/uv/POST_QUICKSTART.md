The worker runs this task every five minutes on its cron schedule.

To run it on demand (in another terminal):
```sh
hatchet trigger manual-run
```

**Notes:**
- The virtual environment will be automatically created by `hatchet worker dev`
- `uv run` automatically uses the virtual environment, no need to activate it manually
- The `uv.lock` file will be automatically generated when dependencies are installed
