import asyncio

from workflows.scheduled_workflow import ScheduledInput, scheduled_task


async def main() -> None:
    result = await scheduled_task.aio_run(
        ScheduledInput(message="hello from a manual run")
    )

    print(result.message)


if __name__ == "__main__":
    asyncio.run(main())
