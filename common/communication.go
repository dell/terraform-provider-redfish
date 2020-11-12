package common

/*ResourceResult struct is used for go-routines spanwed in methods CREATE, UPDATE and DELETE
to retrieve the actual result of the execution of that go-routine.
Depending on what actually happened, the values entries will be set accordingly.
Values:
  - Endpoint: Endpoint of the client. This is used to keep track of each redfish endpoint (i.e. https://my-wonderful-idrac.net).
  - ID: ID of the resource created. If for some reason the resource creation fails, it must be either not set or set to "".
  - Error: value where says if there was an error on the go-routine execution. True if there was an error, false if it wasn't.
  - ErrorMsg: if Error value has been set to true, here must be specified the error message that lead the go-routine to fail.*/
type ResourceResult struct {
	Endpoint string
	ID       string
	Error    bool
	ErrorMsg string
}

/*ResourceChanged struct is used for go-routines spanwed in method READ to see if a resource must be updates or not.
Values:
  - HasChanged: if the go-routine figures out that the state in the terraform state file is different from the actual infrastructure,
				this must be set to true. False must be used if there was no changes.
  - Error: value where says if there was an error on the go-routine execution. True if there was an error, false if it wasn't.
  - ErrorMsg:  if Error value has been set to true, here must be specified the error message that lead the go-routine to fail.*/
type ResourceChanged struct {
	HasChanged   bool
	Error        bool
	ErrorMessage string
}
