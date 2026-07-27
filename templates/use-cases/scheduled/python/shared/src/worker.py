from hatchet_client import hatchet
from workflows.scheduled_workflow import scheduled_task


def main() -> None:
    worker = hatchet.worker("scheduled-worker", workflows=[scheduled_task])
    worker.start()


if __name__ == "__main__":
    main()
