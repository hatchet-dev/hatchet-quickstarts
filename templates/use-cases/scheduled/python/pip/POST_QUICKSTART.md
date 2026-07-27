The worker runs this task every five minutes on its cron schedule.

To run it on demand (in another terminal):
```sh
hatchet trigger manual-run
```

**Notes:**
- The virtual environment (`.venv`) will be automatically created by `hatchet worker dev`
- If running commands manually (not via `hatchet worker`), make sure to activate the virtual environment first:
  ```sh
  source .venv/bin/activate  # On Windows: .venv\Scripts\activate
  ```
