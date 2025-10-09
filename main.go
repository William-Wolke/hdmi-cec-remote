package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	intmsbetweenkeys    = 4000 // ms
	intmousestartspeed  = 10
	intmouseacc         = 10
	intmousespeed       = intmousestartspeed
	datlastkey          = time.Now()
	strlastkey          = ""
	intkeychar          = 0
	keyIsPressed 	    = false
)

func getBaseKeyName(line, eventType string) (keyName string, ok bool) {
    prefix := fmt.Sprintf("%s: ", eventType)
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

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		oneline := scanner.Text()

		// Debug: Print every line
		fmt.Printf("[DEBUG] Raw line: %s\n", oneline)

		// Detect key pressed event using strings.Contains
		if strings.Contains(oneline, "key pressed: ") {
			keyName, ok := getBaseKeyName(oneline, "key pressed")
			if !ok {
				continue
			}
			if keyIsPressed && keyName == strlastkey {
				fmt.Printf("[DEBUG] Ignored duplicate Key pressed: %s\n", keyName)
				continue
			}
			fmt.Printf("[DEBUG] Key pressed: %s\n", keyName)
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
			// case "previous channel":
			// 	runXdotool("key", "Return")
			case "channel up":
				runXdotool("key", "Right")
			case "channel down":
				runXdotool("key", "Left")
			case "channels list":
				runXdotool("click", "3")
			case "up":
				intpixels := -1 * intmousespeed
				runXdotool("mousemove_relative", "--", "0", fmt.Sprintf("%d", intpixels))
				intmousespeed += intmouseacc
			case "down":
				intpixels := 1 * intmousespeed
				runXdotool("mousemove_relative", "--", "0", fmt.Sprintf("%d", intpixels))
				intmousespeed += intmouseacc
			case "left":
				intpixels := -1 * intmousespeed
				runXdotool("mousemove_relative", "--", fmt.Sprintf("%d", intpixels), "0")
				intmousespeed += intmouseacc
			case "right":
				intpixels := 1 * intmousespeed
				runXdotool("mousemove_relative", "--", fmt.Sprintf("%d", intpixels), "0")
				intmousespeed += intmouseacc
			case "select":
				runXdotool("click", "1")
			case "return":
				runXdotool("key", "Alt_L+Left")
			case "exit":
				runXdotool("key", "BackSpace")
			case "F1":
				// Shitty controller has broken down key
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
		} else if strings.Contains(oneline, "key released: ") {
			keyName, ok := getBaseKeyName(oneline, "key released")
			if !ok {
				continue
			}
			fmt.Printf("[DEBUG] Key released: %s\n", keyName)
			keyIsPressed = false
			switch keyName {
			case "stop":
				fmt.Println("Key Released: STOP")
			case "up", "down", "left", "right":
				intmousespeed = intmousestartspeed
			}
		}
	}
}
