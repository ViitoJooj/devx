package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/ViitoJooj/devx/pkg/errorx"
	"golang.org/x/term"
)

// printList writes rows as a table, or items as JSON when asJSON is set.
func printList(asJSON bool, items any, header []string, rows [][]string) {

	if asJSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(items); err != nil {
			errorx.CannotGenerate(err)
		}
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, strings.Join(header, "\t"))
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	w.Flush()
}

// choose shows numbered options and returns the index; scripts must use the command flags instead.
func choose(question string, options []string) int {

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		errorx.NeedDBFlag()
	}

	fmt.Println(question)
	for i, option := range options {
		fmt.Printf("  [%d] %s\n", i+1, option)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")

		answer, err := reader.ReadString('\n')
		if err != nil {
			return len(options) - 1
		}

		index, err := strconv.Atoi(strings.TrimSpace(answer))
		if err == nil && index >= 1 && index <= len(options) {
			return index - 1
		}
	}
}

// confirm asks before destructive commands; --yes skips it and is required without a terminal.
func confirm(question string, yes bool) {

	if yes {
		return
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		errorx.NeedConfirmation()
	}

	fmt.Print(question + " [y/N] ")

	answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))

	if answer != "y" && answer != "yes" {
		fmt.Println("aborted")
		os.Exit(0)
	}
}
