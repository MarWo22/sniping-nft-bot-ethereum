package misc

import (
	"bufio"
	"fmt"
	"os"
)

// Gives a prompt and exits the program
func ExitProgram() {

	fmt.Println("Press any key to exit")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadRune()
	os.Exit(1)
}
