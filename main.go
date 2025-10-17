package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"
)

var (
	intmsbetweenkeys   = 4000 // ms
	intmousestartspeed = 10
	intmouseacc        = 10
	intmousespeed      = intmousestartspeed
	datlastkey         = time.Now()
	strlastkey         = ""
	intkeychar         = 0
	keyIsPressed       = false
	navKeys            = []string{"up", "right", "down", "left"}
	moveCancel         chan struct{} // channel to signal movement goroutine to stop
)

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
	if spaceIdx := strings.Index(namePart, " "); spaceIdx != -1 {
		namePart = namePart[:spaceIdx]
	}
	return namePart, true
}
func runXdotool(args ...string) {
	cmd := exec.Command("xdotool", args...)
	_ = cmd.Run() // ignore errors for now
}

func keychar(parin1 string, parin2 int) {
	parin1len := len(parin1)
	parin2pos := parin2 % parin1len
	char := string(parin1[parin2pos])
	if parin2 > 0 {
		runXdotool("key", "BackSpace")
	}
	switch char {
	case " ":
		char = "space"
	case ".":
		char = "period"
	case "-":
		char = "minus"
	}
	runXdotool("key", char)
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
			runXdotool("mousemove_relative", "--", fmt.Sprintf("%d", x), fmt.Sprintf("%d", y))
			if time.Now().After(accelStart) && speed < 50 {
				speed += intmouseacc
			}
			time.Sleep(30 * time.Millisecond)
		}
	}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	moveCancel = make(chan struct{})
	var moveDir string

	for scanner.Scan() {
		line := scanner.Text()

		// Debug: Print every line
		// fmt.Printf("[DEBUG] Raw line: %s\n", line)

		keyName, isKeyPressed := getKeyEvent(line, "key pressed: ")
		if isKeyPressed {
			isNavKey := slices.Contains(navKeys, keyName)
			if keyIsPressed && keyName == strlastkey && !isNavKey {
				log.Printf("[DEBUG] Ignored duplicate Key pressed: %s\n", keyName)
				continue
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
				keychar("1jkl", intkeychar)
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
			case "0":
				keychar(" 09wxyz", intkeychar)
			case "channel up":
				runXdotool("key", "Right")
			case "channel down":
				runXdotool("key", "Left")
			case "channels list":
				runXdotool("click", "3")
			case "up", "down", "left", "right":
				// Start smooth movement goroutine if not already moving
				if moveDir != keyName {
					// Cancel previous movement if any
					close(moveCancel)
					moveCancel = make(chan struct{})
					moveDir = keyName
					go mouseMoveLoop(keyName, moveCancel)
				}
			case "select":
				runXdotool("click", "1")
			case "return":
				runXdotool("key", "Alt_L+Left")
			case "exit":
				runXdotool("key", "BackSpace")
			case "F1":
				intpixels := 1 * intmousespeed
				runXdotool("mousemove_relative", "--", "0", fmt.Sprintf("%d", intpixels))
				intmousespeed += intmouseacc
			case "F2":
				runXdotool("key", "Pause")
			case "F3":
				runXdotool("key", "C")
			case "F4":
				fmt.Println("Key Pressed: YELLOW C")
			case "rewind":
				fmt.Println("Key Pressed: REWIND")
			case "pause":
				fmt.Println("Key Pressed: PAUSE")
			case "Fast forward":
				fmt.Println("Key Pressed: FAST FORWARD")
			case "play":
				fmt.Println("Key Pressed: PLAY")
			case "stop":
				fmt.Println("Key Pressed: STOP")
			default:
				fmt.Printf("Unrecognized Key Pressed: %s\n", keyName)
			}
		}
		keyName, isKeyReleased := getKeyEvent(line, "key released: ")
		if isKeyReleased {
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
	}
}
