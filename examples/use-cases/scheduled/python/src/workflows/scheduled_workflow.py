from datetime import datetime, timezone

from pydantic import BaseModel

from hatchet_client import hatchet
from hatchet_sdk import Context


class ScheduledInput(BaseModel):
    # Cron runs supply no input, so the message must have a default.
    message: str = ""


class ScheduledOutput(BaseModel):
    message: str
    ran_at: str


# The cron schedule registers when the worker starts, so the task runs every
# five minutes. It also runs on demand through src/run.py.
@hatchet.task(
    name="scheduled-workflow",
    description="Runs every 5 minutes",
    on_crons=["*/5 * * * *"],
    input_validator=ScheduledInput,
)
def scheduled_task(input: ScheduledInput, ctx: Context) -> ScheduledOutput:
    print("scheduled task ran")

    return ScheduledOutput(
        message=input.message,
        ran_at=datetime.now(timezone.utc).isoformat(),
    )
