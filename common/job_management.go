/*
Copyright (c) 2020-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Mozilla Public License Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://mozilla.org/MPL/2.0/


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

const (
	// TimeBetweenAttempts will be used to set the time between job checks (variable is in seconds)
	TimeBetweenAttempts int = 10
	// Timeout will be used to consider a job task failed (variable is in seconds)
	Timeout int = 300
	// StatusCodeSuccess will denote http.response success code
	StatusCodeSuccess int = 200
	// JobErrorWithState will be used to denote that the job has finished with an error
	JobErrorWithState = "the job has finished unsucessfully with a %s state"
	// Percentage to track completion of job
	Percentage int = 100
)

// WaitForTaskToFinish waits for a redfish job to finish.
// Parameters:
//   - jobURI -> URI for the job to check.
//   - timeBetweenAttempts -> time to wait between attempts. I.e. 30 means 30 seconds.
//   - timeout -> maximun time to wait until job is considered failed.
func WaitForTaskToFinish(service *gofish.Service, jobURI string, timeBetweenAttempts int64, timeout int64) error {
	// Create tickers
	attemptTick := time.NewTicker(time.Duration(timeBetweenAttempts) * time.Second)
	timeoutTick := time.NewTicker(time.Duration(timeout) * time.Second)
	// Below 17G device returns location as /redfish/v1/TaskService/Tasks/JOB_ID for same GET call return status as 200 with all the job status.
	// where as 17G device returns location as /redfish/v1/TaskService/TaskMonitors/JOB_ID for same GET call return no content hence
	// we are replacing TaskMonitors to Tasks.
	jobURI = strings.Replace(jobURI, "TaskMonitors", "Tasks", 1)
	for {
		select {
		case <-attemptTick.C:
			// For some reason iDRAC 4.40.00.0 from time to time gives the following error:
			// iDRAC is not ready. The configuration values cannot be accessed. Please retry after a few minutes.
			job, err := redfish.GetTask(service.GetClient(), jobURI)
			if err == nil {
				log.Printf("[DEBUG] - Attempting one more time... Job state is %s\n", job.TaskState)
				// Check if job has finished
				switch status := job.TaskState; status {
				case redfish.CompletedTaskState:
					return nil
				case redfish.KilledTaskState:
					return fmt.Errorf(JobErrorWithState, job.TaskState)
				case redfish.ExceptionTaskState:
					return fmt.Errorf(JobErrorWithState, job.TaskState)
				}
			}
		case <-timeoutTick.C:
			log.Printf("[DEBUG] - Error. Timeout reached\n")
			return fmt.Errorf("timeout waiting for the job to finish")
		}
	}
}

// WaitForJobToFinish waits for a redfish job to finish.
// Parameters:
//   - jobURI -> URI for the job to check.
//   - timeBetweenAttempts -> time to wait between attempts. I.e. 30 means 30 seconds.
//   - timeout -> maximun time to wait until job is considered failed.
func WaitForJobToFinish(service *gofish.Service, jobURI string, timeBetweenAttempts int64, timeout int64) error {
	// Create tickers
	attemptTick := time.NewTicker(time.Duration(timeBetweenAttempts) * time.Second)
	timeoutTick := time.NewTicker(time.Duration(timeout) * time.Second)
	for {
		select {
		case <-attemptTick.C:
			// For some reason iDRAC 4.40.00.0 from time to time gives the following error:
			// iDRAC is not ready. The configuration values cannot be accessed. Please retry after a few minutes.
			job, err := redfish.GetJob(service.GetClient(), jobURI)
			if err == nil {
				log.Printf("[DEBUG] - Attempting one more time... Job state is %s\n", job.JobState)
				// Check if job has finished
				switch status := job.JobState; status {
				case redfish.CompletedJobState:
					return nil
				case redfish.ExceptionJobState:
					return fmt.Errorf(JobErrorWithState, job.JobState)
				}
			}
		case <-timeoutTick.C:
			log.Printf("[DEBUG] - Error. Timeout reached\n")
			return fmt.Errorf("timeout waiting for the job to finish")
		}
	}
}

// GetJobDetailsOnFinish waits for a redfish job to finish and returns the job details.
// Parameters:
//   - jobURI -> URI for the job to check.
//   - timeBetweenAttempts -> time to wait between attempts. I.e. 30 means 30 seconds.
//   - timeout -> maximum time to wait until job is considered failed.
func GetJobDetailsOnFinish(service *gofish.Service, jobURI string, timeBetweenAttempts int64, timeout int64) (*redfish.Job, error) {
	// Create tickers
	attemptTick := time.NewTicker(time.Duration(timeBetweenAttempts) * time.Second)
	timeoutTick := time.NewTicker(time.Duration(timeout) * time.Second)
	defer attemptTick.Stop()
	defer timeoutTick.Stop()

	var job *redfish.Job
	var err error

	for {
		select {
		case <-attemptTick.C:
			// For some reason iDRAC 4.40.00.0 from time to time gives the following error:
			// iDRAC is not ready. The configuration values cannot be accessed. Please retry after a few minutes.
			job, err = redfish.GetJob(service.GetClient(), jobURI)
			if err == nil {
				log.Printf("[DEBUG] - Attempting one more time... Job state is %s\n", job.JobState)
				// Check if job has finished or has some error and return the details
				if job.JobState == redfish.CompletedJobState && job.PercentComplete == Percentage {
					return job, nil
				}
				if job.JobState == redfish.ExceptionJobState {
					return job, fmt.Errorf(JobErrorWithState, job.JobState)
				}
			}
		case <-timeoutTick.C:
			if job == nil {
				return nil, fmt.Errorf("Job details not available. Possible timeout")
			}
			if job.JobState == redfish.StartingJobState ||
				job.JobState == redfish.RunningJobState ||
				job.JobState == redfish.PendingJobState ||
				job.JobState == redfish.NewJobState {
				return job, nil
			}
			log.Printf("[DEBUG] - Error. Timeout reached\n")
			return nil, fmt.Errorf("Job wait timed out after %d minutes", timeout/60)
		}
	}
}

// GetJobAttachment waits for a redfish job to finish and returns the job attachment.
func GetJobAttachment(service *gofish.Service, jobURI string, timeBetweenAttempts int64, timeout int64) ([]byte, error) {
	attemptTick := time.NewTicker(time.Duration(timeBetweenAttempts) * time.Second)
	timeoutTick := time.NewTicker(time.Duration(timeout) * time.Second)
	// Below 17G device returns location as /redfish/v1/TaskService/Tasks/JOB_ID for same GET call return status as 200 with all the job status.
	// where as 17G device returns location as /redfish/v1/TaskService/TaskMonitors/JOB_ID for same GET call return no content hence
	// we are replacing TaskMonitors to Tasks.
	jobURI = strings.Replace(jobURI, "TaskMonitors", "Tasks", 1)
	defer attemptTick.Stop()
	defer timeoutTick.Stop()
	for {
		select {
		case <-attemptTick.C:
			// For some reason iDRAC 4.40.00.0 from time to time gives the following error:
			// iDRAC is not ready. The configuration values cannot be accessed. Please retry after a few minutes.
			resp, err := service.GetClient().Get(jobURI)
			if err != nil {
				return nil, fmt.Errorf("error making request: %w", err)
			}
			// check is required for 17G because jobURI took 7-8 min to return xml data in 17G
			Content := resp.Header["Content-Type"][0]
			if Content != "application/xml;odata.metadata=minimal;charset=utf-8" {
				for {
					time.Sleep(20 * time.Second)
					resp, err = service.GetClient().Get(jobURI)
					if err != nil {
						return nil, fmt.Errorf("error making request: %w", err)
					}
					Content := resp.Header["Content-Type"][0]
					if Content == "application/xml;odata.metadata=minimal;charset=utf-8" {
						break
					}
				}
			}

			if resp.StatusCode == StatusCodeSuccess {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("error reading body: %w", err)
				}
				if resp.Body != nil {
					err = resp.Body.Close()
					if err != nil {
						return nil, fmt.Errorf("error closing body: %w", err)
					}
				}
				return body, nil
			}

		case <-timeoutTick.C:
			return nil, fmt.Errorf("job wait timed out after %d minutes", timeout/60)
		}
	}
}

// OEMJob contains the job details
type OEMJob struct {
	JobState       string `json:"JobState"`
	JobType        string `json:"JobType"`
	Message        string `json:"Message"`
	MessageId      string `json:"MessageId"`
	Name           string `json:"Name"`
	CompletionTime string `json:"CompletionTime"`
}

// Dell contains the job details
type Dell struct {
	OEMJob
}

// DellJob contains the job details
type DellJob struct {
	Dell Dell
}

// WaitForDellJobToFinish waits for a redfish job to finish and returns the job details.
func WaitForDellJobToFinish(service *gofish.Service, jobURI string, timeBetweenAttempts int64, timeout int64) error {
	var oemJob DellJob
	// Create tickers
	attemptTick := time.NewTicker(time.Duration(timeBetweenAttempts) * time.Second)
	timeoutTick := time.NewTicker(time.Duration(timeout) * time.Second)
	// Below 17G device returns location as /redfish/v1/TaskService/Tasks/JOB_ID for same GET call return status as 200 with all the job status.
	// where as 17G device returns location as /redfish/v1/TaskService/TaskMonitors/JOB_ID for same GET call return no content hence
	// we are replacing TaskMonitors to Tasks.
	jobURI = strings.Replace(jobURI, "TaskMonitors", "Tasks", 1)
	for {
		select {
		case <-attemptTick.C:
			// For some reason iDRAC 4.40.00.0 from time to time gives the following error:
			// iDRAC is not ready. The configuration values cannot be accessed. Please retry after a few minutes.
			job, err := redfish.GetTask(service.GetClient(), jobURI)
			if err == nil {
				// Check if job has finished
				switch status := job.TaskState; status {
				case redfish.CompletedTaskState:
					if job.Oem != nil {
						err = json.Unmarshal(job.Oem, &oemJob)
						if err != nil {
							return err
						}
						if oemJob.Dell.JobState == "Failed" {
							return fmt.Errorf("job failed with message: %s", oemJob.Dell.Message)
						}
					}
					return nil
				case redfish.KilledTaskState:
					return fmt.Errorf(JobErrorWithState, job.TaskState)
				case redfish.ExceptionTaskState:
					return fmt.Errorf(JobErrorWithState, job.TaskState)
				}
			}
		case <-timeoutTick.C:
			return fmt.Errorf("timeout waiting for the job to finish")
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
	if resp.StatusCode != StatusCodeSuccess {
		return fmt.Errorf(" error when deleting the task, Delete status code was %d", resp.StatusCode)
	}
	return nil
}
