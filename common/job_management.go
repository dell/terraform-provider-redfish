package common

import (
	"fmt"
	"log"
	"time"

	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

const (
	// TimeBetweenAttempts will be used to set the time between job checks (variable is in seconds)
	TimeBetweenAttempts int = 10
	// Timeout will be used to consider a job task failed (variable is in seconds)
	Timeout int = 300
)

// WaitForJobToFinish waits for a redfish job to finish.
// Parameters:
//   - jobURI -> URI for the job to check.
//   - timeBetweenAttempts -> time to wait between attempts. I.e. 30 means 30 seconds.
//   - timeout -> maximun time to wait until job is considered failed.
func WaitForJobToFinish(service *gofish.Service, jobURI string, timeBetweenAttempts int, timeout int) error {
	// Create tickers
	attemptTick := time.NewTicker(time.Duration(timeBetweenAttempts) * time.Second)
	timeoutTick := time.NewTicker(time.Duration(timeout) * time.Second)
	for {
		select {
		case <-attemptTick.C:
			// For some reason iDRAC 4.40.00.0 from time to time gives the following error: iDRAC is not ready. The configuration values cannot be accessed. Please retry after a few minutes.
			job, err := redfish.GetTask(service.GetClient(), jobURI)
			if err == nil {
				log.Printf("[DEBUG] - Attempting one more time... Job state is %s\n", job.TaskState)
				// Check if job has finished
				switch status := job.TaskState; status {
				case redfish.CompletedTaskState:
					return nil
				case redfish.KilledTaskState:
					return fmt.Errorf("the job has finished unsucessfully with a %s state", job.TaskState)
				case redfish.ExceptionTaskState:
					return fmt.Errorf("the job has finished unsucessfully with a %s state", job.TaskState)
				}
			}
		case <-timeoutTick.C:
			log.Printf("[DEBUG] - Error. Timeout reached\n")
			return fmt.Errorf("Timeout waiting for the job to finish")
		}
	}
}

// DeleteDellJob is intended to delete a task schedules in a Dell system.
// This function is only a workaround until HTTP DELETE is supported under each task o taskmonitor
//
//	Parameters:
//	- taskID: Id of the tasks to delete
func DeleteDellJob(service *gofish.Service, taskID string) error {
	url := "/redfish/v1/Managers/iDRAC.Embedded.1/Jobs/"
	resp, err := service.GetClient().Delete(fmt.Sprintf("%s%s", url, taskID))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error when deleting the task, Delete status code was %d", resp.StatusCode)
	}
	return nil
}
