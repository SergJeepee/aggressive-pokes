package utils

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
	"time"
)

const printBoxLength = 124

var (
	SeparatorLine  = fmt.Sprintf("%v", strings.Repeat("-", printBoxLength-4))
	boxLine        = fmt.Sprintf("%v", strings.Repeat("=", printBoxLength))
	boxLineOpening = fmt.Sprintf("\n%v", strings.Repeat("=", printBoxLength))
)

func PrintBoxed(title string, logs ...string) {
	if len(title) != 0 {
		printTitledBoxLine(title)
	} else {
		printBoxLine(true)
	}
	for _, log := range logs {
		wrapPrint(log)
	}
	printBoxLine(false)
}

func printBoxLine(opening bool) {
	if opening {
		fmt.Println(boxLineOpening)
	} else {
		fmt.Println(boxLine)
	}
}

func printTitledBoxLine(title string) {
	runes := []rune(strings.TrimSpace(strings.ReplaceAll(title, "\n", " ")))
	repeat := printBoxLength - len(runes) - 6
	if repeat <= 0 {
		repeat = 3
		runes = runes[:printBoxLength-12]
		runes = append(runes, []rune("...")...)
	}
	fmt.Printf("\n==== %s %s\n", string(runes), strings.Repeat("=", repeat))
}

func wrapPrint(str string) {
	var wrapped string
	for _, s := range strings.Split(str, "\n") {
		wrapped += wrapPrintLine(s)
	}
	fmt.Printf("%v", wrapped)
}

func wrapPrintLine(str string) string {
	return fmt.Sprintf("| %-120v |\n", str)
}

func PrettyDuration(duration time.Duration) string {
	if duration.Milliseconds() <= 999 {
		return "<1s"
	}
	hours := int64(math.Mod(duration.Hours(), 24))
	minutes := int64(math.Mod(duration.Minutes(), 60))
	seconds := int64(math.Mod(duration.Seconds(), 60))
	//millis := int64(math.Mod(float64(duration.Milliseconds()), 1000))

	chunks := []struct {
		name   string
		amount int64
	}{
		{"h", hours},
		{"m", minutes},
		{"s", seconds},
		//{"ms", millis},
	}

	var parts []string

	for _, chunk := range chunks {
		switch chunk.amount {
		case 0:
			continue
		default:
			parts = append(parts, fmt.Sprintf("%d%s", chunk.amount, chunk.name))
		}
	}

	return strings.Join(parts, " ")
}

func ClearConsole() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		fmt.Printf(err.Error())
	}
}
