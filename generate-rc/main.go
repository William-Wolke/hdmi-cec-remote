package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the YAML configuration
type Config struct {
	Keybinds []Keybind     `yaml:"keybinds"`
	Browser  BrowserConfig `yaml:"browser"`
}

type Keybind struct {
	Key          string   `yaml:"key"`
	Action       string   `yaml:"action"`
	Command      string   `yaml:"command,omitempty"`
	URL          string   `yaml:"url,omitempty"`
	Keys         []string `yaml:"keys,omitempty"`
	Button       string   `yaml:"button,omitempty"`        // left, right, middle
	Direction    string   `yaml:"direction,omitempty"`     // up, down, left, right
	Speed        int      `yaml:"speed,omitempty"`         // pixels to move (default 20)
	RemoteButton string   `yaml:"remote_button,omitempty"` // flirc mapping reference
	Comment      string   `yaml:"comment,omitempty"`
	Steps        []Step   `yaml:"steps,omitempty"` // for sequence action
}

type Step struct {
	Action    string   `yaml:"action,omitempty"`
	Command   string   `yaml:"command,omitempty"`
	URL       string   `yaml:"url,omitempty"`
	Keys      []string `yaml:"keys,omitempty"`
	Button    string   `yaml:"button,omitempty"`
	Direction string   `yaml:"direction,omitempty"`
	Speed     int      `yaml:"speed,omitempty"`
	Delay     int      `yaml:"delay,omitempty"` // delay in milliseconds before this step
}

type BrowserConfig struct {
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
}

var keyMap map[string]int

// getKeyMap reads key codes from the Linux input event codes header
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

// keyAliases maps common key names to their Linux key code names
var keyAliases = map[string]string{
	"ESCAPE":    "ESC",
	"ALT":       "LEFTALT",
	"CTRL":      "LEFTCTRL",
	"CONTROL":   "LEFTCTRL",
	"SHIFT":     "LEFTSHIFT",
	"SUPER":     "LEFTMETA",
	"META":      "LEFTMETA",
	"WIN":       "LEFTMETA",
	"RETURN":    "ENTER",
	"DEL":       "DELETE",
	"PAGEUP":    "PAGEUP",
	"PAGEDOWN":  "PAGEDOWN",
	"CAPSLOCK":  "CAPSLOCK",
	"NUMLOCK":   "NUMLOCK",
	"SCROLLOCK": "SCROLLLOCK",
}

// getKeyCode converts a key name to its Linux key code
func getKeyCode(keyName string) (int, bool) {
	normalized := strings.ToUpper(strings.ReplaceAll(keyName, " ", "_"))
	// Check for alias
	if alias, ok := keyAliases[normalized]; ok {
		normalized = alias
	}
	code, ok := keyMap["KEY_"+normalized]
	return code, ok
}

// Mouse button codes for ydotool
var mouseButtons = map[string]int{
	"left":   0,
	"right":  1,
	"middle": 2,
}

// toYdotoolClick generates the ydotool command for mouse clicks
func toYdotoolClick(button string) string {
	buttonCode, ok := mouseButtons[strings.ToLower(button)]
	if !ok {
		log.Printf("Warning: Unknown mouse button: %s, defaulting to left", button)
		buttonCode = 0
	}
	return fmt.Sprintf("ydotool click 0xC%d", buttonCode)
}

// toYdotoolMousemove generates the ydotool command for mouse movement
func toYdotoolMousemove(direction string, speed int) string {
	if speed == 0 {
		speed = 20 // default speed
	}
	var x, y int
	switch strings.ToLower(direction) {
	case "up":
		x, y = 0, -speed
	case "down":
		x, y = 0, speed
	case "left":
		x, y = -speed, 0
	case "right":
		x, y = speed, 0
	default:
		log.Printf("Warning: Unknown direction: %s", direction)
		return ""
	}
	return fmt.Sprintf("ydotool mousemove -- %d %d", x, y)
}

// toYdotoolScroll generates the ydotool command for scrolling
func toYdotoolScroll(direction string, amount int) string {
	if amount == 0 {
		amount = 3 // default scroll amount
	}
	var y int
	switch strings.ToLower(direction) {
	case "up":
		y = amount
	case "down":
		y = -amount
	default:
		log.Printf("Warning: Unknown scroll direction: %s", direction)
		return ""
	}
	return fmt.Sprintf("ydotool mousemove --wheel -- 0 %d", y)
}

// toYdotoolKeypress generates the ydotool command for pressing keys
func toYdotoolKeypress(keys ...string) string {
	var args []string
	var codes []int
	for _, key := range keys {
		code, ok := getKeyCode(key)
		if !ok {
			log.Printf("Warning: Unknown key: %s", key)
			continue
		}
		args = append(args, fmt.Sprintf("%d:1", code))
		codes = append(codes, code)
	}
	// Release keys in reverse order
	for i := len(codes) - 1; i >= 0; i-- {
		args = append(args, fmt.Sprintf("%d:0", codes[i]))
	}
	return "ydotool key " + strings.Join(args, " ")
}

// toBrowserCommand generates the command to open a browser window
func toBrowserCommand(cfg BrowserConfig, url string) string {
	args := append(cfg.Args, url)
	return cfg.Command + " " + strings.Join(args, " ")
}

// stepToCommand converts a Step to a shell command
func stepToCommand(step Step, cfg BrowserConfig) string {
	switch step.Action {
	case "execute":
		return step.Command
	case "browser":
		return toBrowserCommand(cfg, step.URL)
	case "keypress":
		return toYdotoolKeypress(step.Keys...)
	case "click":
		return toYdotoolClick(step.Button)
	case "mousemove":
		return toYdotoolMousemove(step.Direction, step.Speed)
	case "scroll":
		return toYdotoolScroll(step.Direction, step.Speed)
	default:
		return ""
	}
}

// toSequenceCommands generates a list of commands for multiple labwc actions
func toSequenceCommands(steps []Step, cfg BrowserConfig) []string {
	var commands []string

	for _, step := range steps {
		// Add delay if specified (convert ms to seconds)
		if step.Delay > 0 {
			commands = append(commands, fmt.Sprintf("sleep %.1f", float64(step.Delay)/1000))
		}
		// For browser actions, close the previous tab first
		if step.Action == "browser" {
			commands = append(commands, toYdotoolKeypress("Ctrl", "W"))
		}
		cmd := stepToCommand(step, cfg)
		if cmd != "" {
			commands = append(commands, cmd)
		}
	}
	return commands
}

// escapeXML escapes special characters for XML
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

func generateRcXML(config Config) string {
	var sb strings.Builder

	sb.WriteString(`<?xml version="1.0" ?>
<labwc_config>
  <keyboard>
    <default />
    <!-- Generated keybindings from keybinds.yml -->
`)

	for _, kb := range config.Keybinds {
		// Add comment if present
		if kb.Comment != "" {
			sb.WriteString(fmt.Sprintf("    <!-- %s -->\n", kb.Comment))
		}

		sb.WriteString(fmt.Sprintf("    <keybind key=\"%s\">\n", kb.Key))

		// Handle sequence action separately - generates multiple <action> elements
		if kb.Action == "sequence" {
			commands := toSequenceCommands(kb.Steps, config.Browser)
			if len(commands) == 0 {
				log.Printf("Warning: Empty sequence for key %s", kb.Key)
				continue
			}
			for _, cmd := range commands {
				sb.WriteString(fmt.Sprintf("      <action name=\"Execute\" command=\"%s\" />\n", escapeXML(cmd)))
			}
			sb.WriteString("    </keybind>\n")
			continue
		}

		var command string
		switch kb.Action {
		case "execute":
			command = kb.Command
		case "browser":
			command = toBrowserCommand(config.Browser, kb.URL)
		case "keypress":
			command = toYdotoolKeypress(kb.Keys...)
		case "click":
			command = toYdotoolClick(kb.Button)
		case "mousemove":
			command = toYdotoolMousemove(kb.Direction, kb.Speed)
			if command == "" {
				continue
			}
		case "scroll":
			command = toYdotoolScroll(kb.Direction, kb.Speed)
			if command == "" {
				continue
			}
		case "dpad":
			// Uses the dpad-helper binary for mode-aware input
			command = "dpad-helper " + kb.Direction
		default:
			log.Printf("Warning: Unknown action type: %s", kb.Action)
			continue
		}

		// For browser actions, close the previous tab first
		if kb.Action == "browser" {
			closeTabCmd := toYdotoolKeypress("Ctrl", "W")
			sb.WriteString(fmt.Sprintf("      <action name=\"Execute\" command=\"%s\" />\n", escapeXML(closeTabCmd)))
		}
		sb.WriteString(fmt.Sprintf("      <action name=\"Execute\" command=\"%s\" />\n", escapeXML(command)))
		sb.WriteString("    </keybind>\n")
	}

	sb.WriteString(`  </keyboard>
</labwc_config>
`)

	return sb.String()
}

func main() {
	// Initialize key map
	keyMap = getKeyMap()

	// Read config file
	configPath := "keybinds.yml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("Failed to parse YAML: %v", err)
	}

	// Set default browser config if not specified
	if config.Browser.Command == "" {
		config.Browser.Command = "chromium"
		config.Browser.Args = []string{"--kiosk", "--noerrdialogs", "--disable-infobars", "--password-store=basic"}
	}

	// Generate and output rc.xml
	output := generateRcXML(config)

	// Write to file or stdout
	outputPath := "rc.xml"
	if len(os.Args) > 2 {
		outputPath = os.Args[2]
	}

	if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}

	fmt.Printf("Generated %s from %s\n", outputPath, configPath)
}
