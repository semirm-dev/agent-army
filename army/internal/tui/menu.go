package tui

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// SelectOne displays a numbered menu and returns the selected item.
func SelectOne(p Prompter, w io.Writer, prompt string, items []string) (string, error) {
	printNumberedMenu(w, items)
	for {
		raw, err := p.Prompt(prompt + " ")
		if err != nil {
			return "", err
		}
		choice, ok := parseInt(raw)
		if ok && choice >= 1 && choice <= len(items) {
			return items[choice-1], nil
		}
		fmt.Fprintf(w, "Invalid choice. Enter a number between 1 and %d.\n", len(items))
	}
}

// SelectMulti displays a numbered menu for comma-separated multi-selection.
func SelectMulti(p Prompter, w io.Writer, prompt string, items []string) ([]string, error) {
	printNumberedMenu(w, items)
	for {
		raw, err := p.Prompt(prompt + " (comma-separated, e.g. 1,3,5): ")
		if err != nil {
			return nil, err
		}
		selected, ok := parseMultiChoice(raw, len(items))
		if ok && len(selected) > 0 {
			result := make([]string, len(selected))
			for i, idx := range selected {
				result[i] = items[idx-1]
			}
			return result, nil
		}
	}
}

// PromptWithDefault prompts for text input with a default value.
func PromptWithDefault(p Prompter, prompt, defaultVal string) (string, error) {
	raw, err := p.Prompt(fmt.Sprintf("%s [%s]: ", prompt, defaultVal))
	if err != nil {
		return "", err
	}
	s := strings.TrimSpace(raw)
	if s == "" {
		return defaultVal, nil
	}
	return s, nil
}

// SelectOneWithDefault displays a numbered menu with a default marked by (*).
func SelectOneWithDefault(p Prompter, w io.Writer, prompt string, items []string, defaultVal string) (string, error) {
	printNumberedMenuWithDefault(w, items, defaultVal)
	for {
		raw, err := p.Prompt(prompt + " ")
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(raw) == "" {
			return defaultVal, nil
		}
		choice, ok := parseInt(raw)
		if ok && choice >= 1 && choice <= len(items) {
			return items[choice-1], nil
		}
		fmt.Fprintf(w, "Invalid choice. Enter a number between 1 and %d, or press Enter for default.\n", len(items))
	}
}

// SelectMultiOptional allows empty selection (returns nil on Enter).
func SelectMultiOptional(p Prompter, w io.Writer, prompt string, items []string) ([]string, error) {
	printNumberedMenu(w, items)
	for {
		raw, err := p.Prompt(prompt + " (comma-separated, or Enter to skip): ")
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(raw) == "" {
			return nil, nil
		}
		selected, ok := parseMultiChoice(raw, len(items))
		if ok && len(selected) > 0 {
			result := make([]string, len(selected))
			for i, idx := range selected {
				result[i] = items[idx-1]
			}
			return result, nil
		}
	}
}

func printNumberedMenu(w io.Writer, items []string) {
	fmt.Fprintln(w)
	for i, item := range items {
		fmt.Fprintf(w, "  %d) %s\n", i+1, item)
	}
	fmt.Fprintln(w)
}

func printNumberedMenuWithDefault(w io.Writer, items []string, defaultVal string) {
	fmt.Fprintln(w)
	for i, item := range items {
		marker := ""
		if item == defaultVal {
			marker = " (*)"
		}
		fmt.Fprintf(w, "  %d) %s%s\n", i+1, item, marker)
	}
	fmt.Fprintln(w)
}

func parseInt(raw string) (int, bool) {
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, false
	}
	return n, true
}

func parseMultiChoice(raw string, maxVal int) ([]int, bool) {
	parts := strings.Split(raw, ",")
	var result []int
	for _, p := range parts {
		p = strings.TrimSpace(p)
		n, err := strconv.Atoi(p)
		if err != nil || n < 1 || n > maxVal {
			return nil, false
		}
		result = append(result, n)
	}
	return result, len(result) > 0
}
