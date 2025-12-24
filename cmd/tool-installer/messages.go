// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
)

const ansiReset = "\x1B[0m"

type Color struct {
	red   uint8
	green uint8
	blue  uint8
}

var ErrorRed = Color{220, 50, 47}
var SuccessGreen = Color{50, 205, 50}

func (c *Color) String() string {
	return fmt.Sprintf("\x1B[38;2;%d;%d;%dm", c.red, c.green, c.blue)
}

func isColorEnabled() bool {
	_, exists := os.LookupEnv("NO_COLOR")
	return !exists
}

var colorEnabled = isColorEnabled()

func colorPrintln(color Color, message string) {
	if colorEnabled {
		fmt.Printf("%s%s%s\n", color.String(), message, ansiReset)
	} else {
		fmt.Println(message)
	}
}

type MessageType int

const (
	Success MessageType = iota
	Info
	Error
)

type UserMessage struct {
	Type    MessageType
	Tool    string
	Content string
}

func (m *UserMessage) Print() {
	switch m.Type {
	case Success:
		colorPrintln(SuccessGreen, fmt.Sprintf("%s: %s", m.Tool, m.Content))
	case Info:
		fmt.Printf("%s: info: %s\n", m.Tool, m.Content)
	case Error:
		colorPrintln(ErrorRed, fmt.Sprintf("%s: error: %s", m.Tool, m.Content))
	}
}
