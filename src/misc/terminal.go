package misc

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"golang.org/x/sys/windows"
)

const (
	GREEN  = "\u001b[32m"
	YELLOW = "\u001b[33m"
	RED    = "\u001b[31m"
	RESET  = "\u001b[0m"
)

func PrintGreen(message string, printTime bool) {
	if printTime {
		printFormattedTime()
	}
	fmt.Println(GREEN + message + RESET)
}

func PrintYellow(message string, printTime bool) {
	if printTime {
		printFormattedTime()
	}
	fmt.Println(YELLOW + message + RESET)
}

func PrintMonitorIteration(iteration int) {
	if iteration != 0 {
		ClearLastLine()
	}
	fmt.Print(YELLOW + "Monitoring for reveal, #" + strconv.Itoa(iteration) + RESET)
}

func PrintRed(message string, printTime bool) {
	if printTime {
		printFormattedTime()
	}
	fmt.Println(RED + message + RESET)
}

func printFormattedTime() {
	currentTime := time.Now()
	timeFormat := currentTime.Format("2006-01-02 15:04:05.000")

	fmt.Print("[" + timeFormat + "] ")
}

func SetTitle(title string) error {
	cmd := exec.Command("cmd", "/C", "tite"+title)
	return cmd.Run()
}

func HideCursor() {
	fmt.Print("\x1b[?25l")
}

func ShowCursor() {
	fmt.Print("\x1b[25h")
}

func InitTerminal() {
	stdout := windows.Handle(os.Stdout.Fd())
	var originalMode uint32

	windows.GetConsoleMode(stdout, &originalMode)
	windows.SetConsoleMode(stdout, originalMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)

	SetTitle("Sniping NFT bot")
	HideCursor()
}

func ResetTerminal() {
	fmt.Print("\033c")
	InitTerminal()
}

func ClearLastLine() {
	fmt.Print("\033[2K\r")
}
