package godiawi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// StatusRequest is used to poll diawi to see the status
// of the upload.
type StatusRequest struct {
	// Required parameters
	Token         string
	JobIdentifier string
}

// GetJobStatus makes a status request to see the current progress
// for the app upload.
func (s *StatusRequest) GetJobStatus() (*StatusResponse, error) {

	url := statusURL + "?token=" + s.Token + "&job=" + s.JobIdentifier
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Submit the request
	client := &http.Client{Timeout: StatusTimeoutSeconds * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status %s. Response body: %s", res.Status, res.Body)
	}

	resData, err := ioutil.ReadAll(res.Body)

	statusRes := StatusResponse{}
	err = json.Unmarshal(resData, &statusRes)
	if err != nil {
		return nil, err
	}

	return &statusRes, nil
}

const MaxPollsReachedError = errors.New("Exceeded max number of polls to get upload status.")
const UnknownStatusError = errors.New("Unknown status error")

// WaitForFinishedStatus continually pings diawi using the GetJobStatus
// until the service provides a DiawiStatus other than
// Processing (2001)
func (s *StatusRequest) WaitForFinishedStatus() (*StatusResponse, error) {
	sr, err := s.GetJobStatus()
	if err != nil {
		return nil, err
	}

	numberOfPolls := 0
	for {
		switch sr.Status {
		case Processing:

			if numberOfPolls > StatusPollingMax {
				return nil, MaxPollsReachedError
			}

			time.Sleep(1 * time.Second) // diawi documentation recommends 1 second between each status request.
			sr, err = s.GetJobStatus()
			if err != nil {
				return nil, err
			}

		case Ok:
			return sr, nil
		case ErrorOccured:
			return sr, fmt.Errorf("Response included error status = 4000. sr: %s", sr.String())
		default:
			return sr, UnknownStatusError
		}
	}
}

// StatusResponse is the response provided by the
type StatusResponse struct {
	Status  DiawiStatus
	Message string

	// Only provided when upload successful
	Hash string
	Link string
}

func (s *StatusResponse) String() string {
	return fmt.Sprintf("status: %d messsage: %s hash: %s link: %s", s.Status, s.Message, s.Hash, s.Link)
}
