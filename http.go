package gloat

import (
	"net/http"
)

func HttpGet(url string) TestFunction {
	return func() TestResult {
		response, err := http.Get(url)
		if err != nil {
			return Status_Failure
		}
		if response.StatusCode/100 != 2 {
			return Status_Failure
		}
		return Status_Success
	}
}
