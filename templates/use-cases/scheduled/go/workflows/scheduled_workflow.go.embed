package workflows

import (
	"fmt"
	"time"

	hatchet "github.com/hatchet-dev/hatchet/sdks/go"
)

type ScheduledInput struct {
	Message string `json:"message"`
}

type ScheduledOutput struct {
	Message string `json:"message"`
	RanAt   string `json:"ran_at"`
}

// ScheduledWorkflow runs on a cron schedule. WithWorkflowCron registers the
// schedule when the worker starts, so the task fires every five minutes. The
// task also runs on demand through cmd/run.
func ScheduledWorkflow(c *hatchet.Client) *hatchet.StandaloneTask {
	return c.NewStandaloneTask("scheduled-workflow", func(ctx hatchet.Context, input ScheduledInput) (ScheduledOutput, error) {
		fmt.Println("scheduled task ran")

		return ScheduledOutput{
			Message: input.Message,
			RanAt:   time.Now().Format(time.RFC3339),
		}, nil
	},
		hatchet.WithWorkflowCron("*/5 * * * *"),
		hatchet.WithWorkflowDescription("Runs every 5 minutes"),
	)
}
