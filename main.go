package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	intmsbetweenkeys   = 4000 // ms
	intmousestartspeed = 5
	intmouseacc        = 2
	intmousemaxspeed   = 25
	SCROLL_UP          = 4
	SCROLL_DOWN        = 5
	LEFT_CLICK         = 0
	RIGHT_CLICK        = 1
)

var (
	intmousespeed = intmousestartspeed
	datlastkey    = time.Now()
	strlastkey    = ""
	intkeychar    = 0
	keyIsPressed  = false
	navKeys       = []string{"up", "right", "down", "left"}
	moveCancel    chan struct{} // channel to signal movement goroutine to stop
	moveDir       string
	keyMap        map[string]int = map[string]int{}
)

func getKeyMap() map[string]int {
	file, err := os.Open("/usr/include/linux/input-event-codes.h")
	if err != nil {
		panic(err)
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

func getKeyCode(keyName string) (int, bool) {
	code, ok := keyMap["KEY_"+strings.ToUpper(strings.ReplaceAll(keyName, " ", "_"))]
	return code, ok
}

func getBaseKeyName(line, eventType string) (keyName string, ok bool) {
	prefix := eventType
	idx := strings.Index(line, prefix)
	if idx == -1 {
		return "", false
	}
	after := line[idx+len(prefix):]
	// Find first " (" which marks the start of extra info/code
	parenIdx := strings.Index(after, " (")
	if parenIdx == -1 {
		return "", false
	}
	// Take everything up to " (" and trim
	namePart := strings.TrimSpace(after[:parenIdx])
	// If there is extra info in parentheses, remove it (e.g. "F1 (blue)" → "F1")
	// Only keep the part before any " (" or before any " (" after the key name
	return namePart, true
}

func pressKey(keys ...string) {
	runYdotoolArgs := []string{"key"}
	var codes []int
	for _, key := range keys {
		code, ok := getKeyCode(key)
		if !ok {
			log.Printf("Unknown key: %s", key)
			continue
		}
		runYdotoolArgs = append(runYdotoolArgs, fmt.Sprintf("%d:1", code))
		codes = append(codes, code)
	}
	// Release keys in reverse order
	for i := len(codes) - 1; i >= 0; i-- {
		runYdotoolArgs = append(runYdotoolArgs, fmt.Sprintf("%d:0", codes[i]))
	}
	runYdotool(runYdotoolArgs...)
	return
}

func clickMouse(button int) {
	runYdotool("click", fmt.Sprintf("0xC%d", button))
	return
}

func runYdotool(args ...string) {
	cmd := exec.Command("ydotool", args...)
	if err := cmd.Run(); err != nil {
		log.Printf("ydotool error: %v", err)
	}
}

func keychar(parin1 string, parin2 int) {
	parin1len := len(parin1)
	parin2pos := parin2 % parin1len
	char := string(parin1[parin2pos])
	if parin2 > 0 {
		pressKey("backspace")
	}
	switch char {
	case " ":
		char = "space"
	case ".":
		char = "dot"
	case "-":
		char = "minus"
	}
	pressKey(char)
}

func getKeyEvent(line string, eventType string) (keyName string, isKeyPressed bool) {
	if !strings.Contains(line, eventType) {
		return "", false
	}
	keyName, ok := getBaseKeyName(line, eventType)
	if !ok {
		return "", false
	}
	return keyName, true
}

func mouseMoveLoop(dir string, cancelChan chan struct{}) {
	speed := intmousestartspeed
	accelStart := time.Now().Add(1 * time.Second)
	for {
		select {
		case <-cancelChan:
			return
		default:
			var x, y int
			switch dir {
			case "up":
				x, y = 0, -speed
			case "down":
				x, y = 0, speed
			case "left":
				x, y = -speed, 0
			case "right":
				x, y = speed, 0
			}
			runYdotool("mousemove", "--", fmt.Sprintf("%d", x), fmt.Sprintf("%d", y))
			if time.Now().After(accelStart) && speed < intmousemaxspeed {
				speed += intmouseacc
			}
			time.Sleep(30 * time.Millisecond)
		}
	}
}

func moveMouse(keyName string) {
	// Start smooth movement goroutine if not already moving
	if moveDir != keyName {
		// Cancel previous movement if any
		close(moveCancel)
		moveCancel = make(chan struct{})
		moveDir = keyName
		go mouseMoveLoop(keyName, moveCancel)
	}
}

func onKeyPress(keyName string) {
	isNavKey := slices.Contains(navKeys, keyName)
	isScrollKey := keyName == "channel up" || keyName == "channel down"
	if keyIsPressed && keyName == strlastkey && !isNavKey && !isScrollKey {
		log.Printf("[DEBUG] Ignored duplicate Key pressed: %s\n", keyName)
		return
	}
	log.Printf("[DEBUG] Key pressed: %s\n", keyName)
	keyIsPressed = true
	datnow := time.Now()
	datdiff := int(datnow.Sub(datlastkey).Milliseconds())

	if keyName == strlastkey && datdiff < intmsbetweenkeys {
		intkeychar++
	} else {
		intkeychar = 0
	}
	datlastkey = datnow
	strlastkey = keyName

	switch keyName {
	case "1":
		keychar("1.!", intkeychar)
	case "2":
		keychar("abc2", intkeychar)
	case "3":
		keychar("def3", intkeychar)
	case "4":
		keychar("ghi4", intkeychar)
	case "5":
		keychar("jkl5", intkeychar)
	case "6":
		keychar("mno6", intkeychar)
	case "7":
		keychar("pqrs7", intkeychar)
	case "8":
		keychar("tuv8", intkeychar)
	case "9":
		keychar("wxyz9", intkeychar)
	// case "0":
	// 	keychar(" 0", intkeychar)
	case "channel up":
		pressKey("up")
	case "channel down":
		pressKey("down")
	case "channels list":
		clickMouse(RIGHT_CLICK)
	case "select":
		clickMouse(LEFT_CLICK)
	case "up", "down", "left", "right":
		moveMouse(keyName)
	case "return":
		pressKey("Alt", "L", "Left")
	case "exit":
		pressKey("BackSpace")
	case "clear":
		pressKey("Escape")
	case "F1":
		pressKey("Right") // Skip forward
	case "F2":
		pressKey("Left") // Skip backward
	case "F3":
		pressKey("C") // Toggle subtitles
	case "F4":
		pressKey("F") // Full screen
	default:
		log.Printf("Unrecognized Key Pressed: %s\n", keyName)
	}
}

func onKeyRelease(keyName string) {
	log.Printf("[DEBUG] Key released: %s\n", keyName)
	keyIsPressed = false
	switch keyName {
	case "stop":
		fmt.Println("Key Released: STOP")
	case "up", "down", "left", "right":
		intmousespeed = intmousestartspeed
		// Stop movement goroutine
		close(moveCancel)
		moveCancel = make(chan struct{})
		moveDir = ""
	}
}

func main() {
	keyMap = getKeyMap()
	scanner := bufio.NewScanner(os.Stdin)
	moveCancel = make(chan struct{})

	for scanner.Scan() {
		line := scanner.Text()

		// Debug: Print every line
		// fmt.Printf("[DEBUG] Raw line: %s\n", line)

		keyName, isKeyPressed := getKeyEvent(line, "key pressed: ")
		if isKeyPressed {
			onKeyPress(keyName)
		}
		keyName, isKeyReleased := getKeyEvent(line, "key released: ")
		if isKeyReleased {
			onKeyRelease(keyName)
		}
	}
}
