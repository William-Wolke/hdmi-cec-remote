package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"time"
)

var (
	intmsbetweenkeys    = 2000 // ms
	intmousestartspeed  = 10
	intmouseacc         = 10
	intmousespeed       = intmousestartspeed
	datlastkey          = time.Now()
	strlastkey          = ""
	intkeychar          = 0
	reKeyPressed        = regexp.MustCompile(`key pressed: ([^ ]+) \((\d+)\)`)
	reKeyReleased       = regexp.MustCompile(`key released: ([^ ]+) \((\d+)\)`)
)

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

		if m := reKeyPressed.FindStringSubmatch(oneline); len(m) > 0 {
			strkey := m[1]
			fmt.Printf("[DEBUG] Key pressed: %s\n", strkey)
			datnow := time.Now()
			datdiff := int(datnow.Sub(datlastkey).Milliseconds())

			if strkey == strlastkey && datdiff < intmsbetweenkeys {
				intkeychar++
			} else {
				intkeychar = 0
			}
			datlastkey = datnow
			strlastkey = strkey

			switch strkey {
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
				fmt.Printf("Unrecognized Key Pressed: %s\n", strkey)
			}
		} else if m := reKeyReleased.FindStringSubmatch(oneline); len(m) > 0 {
			strkey := m[1]
			fmt.Printf("[DEBUG] Key released: %s\n", strkey)
			switch strkey {
			case "stop":
				fmt.Println("Key Released: STOP")
			case "up", "down", "left", "right":
				intmousespeed = intmousestartspeed
			}
		}
	}
}
