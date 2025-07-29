package http_extra

import (
	"net/http"
	"time"
)

var HttpClient = http.Client{Timeout: 2000 * time.Millisecond}
var HttpClientHighTimeout = http.Client{Timeout: 30 * time.Second}
