package src

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type Task struct {
	BaseURI  string
	Token    int
	Appender string
}

type Trait struct {
	Image      string `json:"image"`
	Attributes []struct {
		TraitType string      `json:"trait_type"`
		Value     interface{} `json:"value"`
	} `json:"attributes"`
}

func taskManager(tasks chan Task, output chan Output, terminate chan bool) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	for {
		select {
		case <-terminate:
			cancel()
			return
		default:
			task := <-tasks
			if task != (Task{}) {
				go execute(task, ctx, output, terminate, 0)
			}
		}
	}
}

func execute(task Task, ctx context.Context, output chan Output, terminate chan bool, attempt int) {
	// NewRequestWithContext will cancel its request immediately if
	// the caller cancels the context
	url := task.BaseURI + strconv.Itoa(task.Token) + task.Appender
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		url,
		nil,
	)
	if err != nil {
		handleError(task, ctx, output, terminate, attempt, err)
		return
	}

	// Issue the request
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		handleError(task, ctx, output, terminate, attempt, err)
		return
	} else if rsp.StatusCode != http.StatusOK {
		handleStatusCode(rsp.StatusCode, task, ctx, output, terminate, attempt)
		return
	}

	// Parse the response into piece
	defer rsp.Body.Close()
	var traits Trait
	if err := json.NewDecoder(rsp.Body).Decode(&traits); err != nil {
		handleError(task, ctx, output, terminate, attempt, err)
		return
	}
	safeWriteChannel(task.Token, traits, output, terminate)
}

func handleStatusCode(statusCode int, task Task, ctx context.Context, output chan Output, terminate chan bool, attempt int) {

	switch statusCode {
	case http.StatusTooManyRequests:
		log(strconv.Itoa(task.Token) + " - Rate Limit: retrying in 500ms")
		time.Sleep(500 * time.Millisecond)
		execute(task, ctx, output, terminate, attempt)

	case http.StatusInternalServerError:
		log(strconv.Itoa(task.Token) + " - Error: " + strconv.Itoa(statusCode) + " - Retrying")
		time.Sleep(500 * time.Millisecond)
		execute(task, ctx, output, terminate, attempt)

	case http.StatusServiceUnavailable:
		log(strconv.Itoa(task.Token) + " - Error: " + strconv.Itoa(statusCode) + " - Retrying")
		time.Sleep(500 * time.Millisecond)
		execute(task, ctx, output, terminate, attempt)

	default:
		if attempt < 3 {
			log(strconv.Itoa(task.Token) + " - Error: " + strconv.Itoa(statusCode) + " - Retrying, attempt " + strconv.Itoa(attempt+1))
			time.Sleep(500 * time.Millisecond)
			execute(task, ctx, output, terminate, attempt+1)
		} else {
			log(strconv.Itoa(task.Token) + " - Error: " + strconv.Itoa(statusCode) + " - Retry limit exceeded, aborting task")
			safeWriteChannel(task.Token, Trait{}, output, terminate)
		}
	}
}

func handleError(task Task, ctx context.Context, output chan Output, terminate chan bool, attempt int, err error) {
	if attempt < 3 {
		log(strconv.Itoa(task.Token) + " - Error: " + err.Error() + " - Retrying, attempt " + strconv.Itoa(attempt+1))
		execute(task, ctx, output, terminate, attempt+1)
	} else {
		log(strconv.Itoa(task.Token) + " - Error: " + err.Error() + " - Retry limit exceeded, aborting task")
		safeWriteChannel(task.Token, Trait{}, output, terminate)
	}
}

func safeWriteChannel(ID int, traits Trait, output chan Output, terminate chan bool) {
	select {
	case <-terminate:
		return
	default:
		output <- Output{
			ID:       ID,
			Response: traits,
		}
	}
}
