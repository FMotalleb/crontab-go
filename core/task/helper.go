package task

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

func LogHTTPResponse(r *http.Response) logrus.LogFunction {
	return func() []any {
		result := &ResponseWriter{}
		err := r.Write(result)
		return []any{
			fmt.Sprintf("error: %v", err),
			"\n",
			result.String(),
		}
	}
}

type ResponseWriter struct {
	buffer []byte
}

// Write implements io.Writer.
func (r *ResponseWriter) Write(p []byte) (n int, err error) {
	initial := len(r.buffer)
	r.buffer = append(r.buffer, p...)
	return len(r.buffer) - initial, nil
}

func (r *ResponseWriter) String() string {
	return string(r.buffer)
}
