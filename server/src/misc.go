package src

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
)

var writeFile *os.File

func log(input string) {
	currentTime := time.Now()
	timeFormat := currentTime.Format("2006-01-02 15:04:05.000")

	writeFile.WriteString("[" + timeFormat + "] " + input + "\n")
}

func createInitLog(inputTask InputTask) string {
	s, _ := json.Marshal(inputTask.IDs)
	ret := "inputData:\n"
	ret += "BaseURI: " + inputTask.BaseURI + "\n"
	ret += "Appender: " + inputTask.Appender + "\n"
	ret += "Gateways: " + strings.Join(inputTask.Gateways, ", ") + "\n"
	ret += "IsIPFS: " + strconv.FormatBool(inputTask.IsIpfs) + "\n"
	ret += "TaskLimit: " + strconv.Itoa(inputTask.TaskLimit) + "\n"
	ret += "TimeOut: " + strconv.Itoa(inputTask.TaskLimit) + "\n"
	ret += "IDs: " + string(s) + "\n"
	return ret
}
