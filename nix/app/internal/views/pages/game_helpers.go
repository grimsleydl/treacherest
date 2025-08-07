package pages

import (
	"html"
	"regexp"
	"strings"
	"unicode/utf8"
)

// CardTextLine represents a formatted line with parts and italic sections
type CardTextLine struct {
	Parts []CardTextPart
}

// CardTextPart represents a part of text that may be italic or a mana symbol
type CardTextPart struct {
	Text       string
	Italic     bool
	IsMana     bool
	ManaSymbol string // The symbol inside the braces (e.g., "2", "W", "U", etc.)
	SafeHTML   string // HTML-safe version of Text with Unicode preserved
}

// FormatCardTextToLines splits card text by | and formats each line
func FormatCardTextToLines(text string) []CardTextLine {
	lines := strings.Split(text, "|")
	result := make([]CardTextLine, len(lines))

	for i, line := range lines {
		result[i] = formatLineWithParentheses(line)
	}

	return result
}

// formatLineWithParentheses formats a single line, marking parenthetical text as italic and extracting mana symbols
func formatLineWithParentheses(line string) CardTextLine {
	var result CardTextLine

	// Keep Unicode characters as-is - the issue is in the HTTP response encoding

	// Regular expressions for mana symbols and parentheses
	manaRe := regexp.MustCompile(`\{([0-9WUBRGCXYZ]+)\}`)

	// Process the line to handle both mana symbols and parentheses
	remaining := line
	inParentheses := false
	currentText := ""

	for len(remaining) > 0 {
		// Check for mana symbol at the start
		if manaMatch := manaRe.FindStringSubmatch(remaining); manaMatch != nil && strings.HasPrefix(remaining, manaMatch[0]) {
			// Save any accumulated text first
			if currentText != "" {
				result.Parts = append(result.Parts, CardTextPart{
					Text:   currentText,
					Italic: inParentheses,
				})
				currentText = ""
			}

			// Add the mana symbol
			result.Parts = append(result.Parts, CardTextPart{
				IsMana:     true,
				ManaSymbol: manaMatch[1],
			})

			// Move past the mana symbol
			remaining = remaining[len(manaMatch[0]):]
			continue
		}

		// Handle character by character for parentheses
		// We need to handle multi-byte UTF-8 characters properly
		r, size := utf8.DecodeRuneInString(remaining)
		remaining = remaining[size:]

		if r == '(' && !inParentheses {
			// Save any accumulated text before parenthesis
			if currentText != "" {
				result.Parts = append(result.Parts, CardTextPart{
					Text:   currentText,
					Italic: false,
				})
				currentText = ""
			}
			inParentheses = true
			currentText = string(r)
		} else if r == ')' && inParentheses {
			currentText += string(r)
			// Save the parenthetical text
			result.Parts = append(result.Parts, CardTextPart{
				Text:   currentText,
				Italic: true,
			})
			currentText = ""
			inParentheses = false
		} else {
			currentText += string(r)
		}
	}

	// Add any remaining text
	if currentText != "" {
		result.Parts = append(result.Parts, CardTextPart{
			Text:   currentText,
			Italic: inParentheses,
		})
	}

	return result
}

// escapeHTMLKeepUnicode escapes dangerous HTML but preserves Unicode.
// Safe for @templ.Raw() because input is from static JSON, not user data.
func escapeHTMLKeepUnicode(text string) string {
	// html.EscapeString handles this correctly - it only escapes
	// <, >, &, ", ' but leaves Unicode characters intact
	return html.EscapeString(text)
}
