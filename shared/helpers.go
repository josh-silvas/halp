package shared

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	input "github.com/tcnksm/go-input"
)

const (
	// Color codes to highlight terminal text

	// Clr0 terminal color Black
	Clr0 = "\x1b[30;1m"
	// ClrR terminal color Red
	ClrR = "\x1b[31;1m"
	// ClrG terminal color Green
	ClrG = "\x1b[32;1m"
	// ClrY terminal color Yellow
	ClrY = "\x1b[33;1m"
	// ClrB terminal color Blue
	ClrB = "\x1b[34;1m"
	// ClrM terminal color Magenta
	ClrM = "\x1b[35;1m"
	// ClrC terminal color Cyan
	ClrC = "\x1b[36;1m"
	// ClrW terminal color White
	ClrW = "\x1b[37;1m"
	// ClrN will end the color text
	ClrN = "\x1b[0m"
)

var (
	// UI defines basic configuration for the ui ask pkg
	UI = &input.UI{Writer: os.Stdout, Reader: os.Stdin}

	// UIOpts defines basic options for the ui ask pkg
	UIOpts = &input.Options{Required: true, HideOrder: true}
)

// Truncate shortens a string if the length of the string exceeds the
// number passed in.
func Truncate(str string, max int) string {
	bnoden := str
	if len(str) > max {
		if max > 3 {
			max -= 3
		}
		bnoden = str[0:max] + "..."
	}
	return bnoden
}

// ConfirmPrompt is a help function to prompt the user for a confirmation.
func ConfirmPrompt(s string) bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s [y/n]: ", s)
		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		response = strings.ToLower(strings.TrimSpace(response))
		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

// ConcatVLANs function will take a list of int vlan numbers, sort the data, then
// convert them to a list of summarized vlan strings for use in cli context
// E.g. Input:  []int{101, 102, 103, 105, 109, 110, 111}
//      Output: []string{"101-103", "105", "109-111"}
func ConcatVLANs(vlans []int) (resp []string) {
	// Sort the integers so we can compare next/last
	// iterables
	sort.Ints(vlans)

	var start = 0

	for i, current := range vlans {
		hasNext := false
		hasLast := false

		// If the next index is within range
		if i+1 != len(vlans) {
			hasNext = vlans[i+1] == current+1
		}

		// If the last index is within range
		if i-1 != -1 {
			hasLast = vlans[i-1] == current-1
		}

		// If the current iteration does not have a neighboring vlan in-front
		// or behind it, add this vlan id as an independent vlan.
		if !hasNext && !hasLast {
			resp = append(resp, strconv.Itoa(current))
			continue
		}

		// If this iteration has a number in front, but not behind,
		// this can be considered the starting vlan.
		if hasLast && !hasNext {
			resp = append(resp, fmt.Sprintf("%d-%d", start, current))
			start = 0
			continue
		}

		// If this iteration has a number in front, but not behind,
		// this can be considered the starting vlan.
		if hasNext && !hasLast {
			start = current
		}
	}
	return resp
}

// IContains is a case-insensitive contains search for a string
func IContains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// IntInSlice is a helper to dermine if a slice of int has a single integer in it.
func IntInSlice(integer int, sl []int) bool {
	for i := range sl {
		if sl[i] == integer {
			return true
		}
	}
	return false
}
