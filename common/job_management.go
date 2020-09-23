package common

import (
	"fmt"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
	"time"
)

const (
	// TimeBetweenAttempts will be used to set the time between job checks (variable is in seconds)
	TimeBetweenAttempts int = 10
	// Timeout will be used to consider a job task failed (variable is in seconds)
	Timeout int = 300
)

// WaitForJobToFinish waits for a redfish job to finish.
// Parameters:
// 	- jobURI -> URI for the job to check.
// 	- timeBetweenAttempts -> time to wait between attempts. I.e. 30 means 30 seconds.
//	- timeout -> maximun time to wait until job is considered failed.
func WaitForJobToFinish(c *gofish.APIClient, jobURI string, timeBetweenAttempts int, timeout int) error {
	// Create tickers
	attemptTick := time.NewTicker(time.Duration(timeBetweenAttempts) * time.Second)
	timeoutTick := time.NewTicker(time.Duration(timeout) * time.Second)
	for {
		select {
		case <-attemptTick.C:
			job, err := redfish.GetTask(c, jobURI)
			if err != nil {
				return err
			}
			fmt.Printf("[DEBUG] - Attempting one more time... Job state is %s\n", job.TaskState)
			//Check if job has finished
			switch status := job.TaskState; status {
			case "Completed":
				return nil
			case "Killed ":
				return fmt.Errorf("the job has finished unsucessfully with a %s state", job.TaskState)
			case "Exception":
				return fmt.Errorf("the job has finished unsucessfully with a %s state", job.TaskState)
			}
		case <-timeoutTick.C:
			fmt.Printf("[DEBUG] - Error. Timeout reached\n")
			return fmt.Errorf("Timeout waiting for the job to finish")
		}
	}
}
