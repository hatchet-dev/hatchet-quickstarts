package main

import (
	"context"
	"fmt"
	"log"

	"github.com/hatchet-dev/hatchet-go-quickstart/client"
	"github.com/hatchet-dev/hatchet-go-quickstart/workflows"
)

func main() {
	c, err := client.HatchetClient()
	if err != nil {
		log.Fatalf("Failed to create Hatchet client: %v", err)
	}

	scheduled := workflows.ScheduledWorkflow(c)

	result, err := scheduled.Run(context.Background(), workflows.ScheduledInput{
		Message: "hello from a manual run",
	})
	if err != nil {
		log.Fatalf("Failed to run Hatchet task: %v", err)
	}

	var scheduledResult workflows.ScheduledOutput

	err = result.Into(&scheduledResult)
	if err != nil {
		log.Fatalf("Failed to convert result to ScheduledOutput: %v", err)
	}

	fmt.Println(scheduledResult.Message)
}
