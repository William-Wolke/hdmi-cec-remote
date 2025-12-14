package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	stateFile     = "/tmp/dpad_mode"
	defaultMode   = "tab" // "tab" or "mouse"
	mouseSpeed    = 25
)

var keyMap map[string]int

func getKeyMap() map[string]int {
	file, err := os.Open("/usr/include/linux/input-event-codes.h")
	if err != nil {
		log.Fatalf("Failed to open input-event-codes.h: %v", err)
	}
	defer file.Close()

	re := regexp.MustCompile(`#define\s+(KEY_\w+)\s+(\d+)`)
	keyMap := make(map[string]int)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if m := re.FindStringSubmatch(line); m != nil {
			code, _ := strconv.Atoi(m[2])
			keyMap[m[1]] = code
		}
	}
	return keyMap
}

var keyAliases = map[string]string{
	"ESCAPE":  "ESC",
	"ALT":     "LEFTALT",
	"CTRL":    "LEFTCTRL",
	"CONTROL": "LEFTCTRL",
	"SHIFT":   "LEFTSHIFT",
	"SUPER":   "LEFTMETA",
	"META":    "LEFTMETA",
	"RETURN":  "ENTER",
}

func getKeyCode(keyName string) (int, bool) {
	normalized := strings.ToUpper(strings.ReplaceAll(keyName, " ", "_"))
	if alias, ok := keyAliases[normalized]; ok {
		normalized = alias
	}
	code, ok := keyMap["KEY_"+normalized]
	return code, ok
}

func pressKeys(keys ...string) {
	var args []string
	args = append(args, "key")
	var codes []int
	for _, key := range keys {
		code, ok := getKeyCode(key)
		if !ok {
			log.Printf("Unknown key: %s", key)
			continue
		}
		args = append(args, fmt.Sprintf("%d:1", code))
		codes = append(codes, code)
	}
	for i := len(codes) - 1; i >= 0; i-- {
		args = append(args, fmt.Sprintf("%d:0", codes[i]))
	}
	exec.Command("ydotool", args...).Run()
}

func moveMouse(direction string) {
	var x, y int
	switch direction {
	case "up":
		x, y = 0, -mouseSpeed
	case "down":
		x, y = 0, mouseSpeed
	case "left":
		x, y = -mouseSpeed, 0
	case "right":
		x, y = mouseSpeed, 0
	}
	exec.Command("ydotool", "mousemove", "--", fmt.Sprintf("%d", x), fmt.Sprintf("%d", y)).Run()
}

func clickMouse(button int) {
	exec.Command("ydotool", "click", fmt.Sprintf("0xC%d", button)).Run()
}

func getMode() string {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return defaultMode
	}
	mode := strings.TrimSpace(string(data))
	if mode != "mouse" && mode != "tab" {
		return defaultMode
	}
	return mode
}

func setMode(mode string) {
	os.WriteFile(stateFile, []byte(mode), 0644)
}

func toggleMode() string {
	mode := getMode()
	if mode == "mouse" {
		mode = "tab"
	} else {
		mode = "mouse"
	}
	setMode(mode)
	return mode
}

func printUsage() {
	fmt.Println("Usage: dpad <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  up, down, left, right  - D-pad direction (behavior depends on mode)")
	fmt.Println("  select                 - Select/click (Enter in tab mode, left-click in mouse mode)")
	fmt.Println("  toggle                 - Toggle between mouse and tab mode")
	fmt.Println("  mode                   - Print current mode")
	fmt.Println("  set-mouse              - Set mouse mode")
	fmt.Println("  set-tab                - Set tab mode")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	keyMap = getKeyMap()
	cmd := os.Args[1]

	switch cmd {
	case "toggle":
		newMode := toggleMode()
		fmt.Printf("Mode: %s\n", newMode)

	case "mode":
		fmt.Println(getMode())

	case "set-mouse":
		setMode("mouse")
		fmt.Println("Mode: mouse")

	case "set-tab":
		setMode("tab")
		fmt.Println("Mode: tab")

	case "up":
		if getMode() == "mouse" {
			moveMouse("up")
		} else {
			pressKeys("Up") // Scroll up
		}

	case "down":
		if getMode() == "mouse" {
			moveMouse("down")
		} else {
			pressKeys("Down") // Scroll down
		}

	case "left":
		if getMode() == "mouse" {
			moveMouse("left")
		} else {
			pressKeys("Shift", "Tab") // Tab backward
		}

	case "right":
		if getMode() == "mouse" {
			moveMouse("right")
		} else {
			pressKeys("Tab") // Tab forward
		}

	case "select":
		if getMode() == "mouse" {
			clickMouse(0) // Left click
		} else {
			pressKeys("Enter")
		}

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}
