package components

import (
	"html"
	"regexp"
	"strings"
	"unicode/utf8"
)

type roleCardTextLine struct {
	Parts []roleCardTextPart
}

type roleCardTextPart struct {
	Text       string
	Italic     bool
	IsMana     bool
	ManaSymbol string
}

func formatRoleCardTextToLines(text string) []roleCardTextLine {
	lines := strings.Split(text, "|")
	result := make([]roleCardTextLine, len(lines))
	for i, line := range lines {
		result[i] = formatRoleCardLine(line)
	}
	return result
}

func formatRoleCardLine(line string) roleCardTextLine {
	var result roleCardTextLine
	manaRe := regexp.MustCompile(`\{([0-9WUBRGCXYZ]+)\}`)
	remaining := line
	inParentheses := false
	currentText := ""

	for len(remaining) > 0 {
		if manaMatch := manaRe.FindStringSubmatch(remaining); manaMatch != nil && strings.HasPrefix(remaining, manaMatch[0]) {
			if currentText != "" {
				result.Parts = append(result.Parts, roleCardTextPart{
					Text:   currentText,
					Italic: inParentheses,
				})
				currentText = ""
			}
			result.Parts = append(result.Parts, roleCardTextPart{
				IsMana:     true,
				ManaSymbol: manaMatch[1],
			})
			remaining = remaining[len(manaMatch[0]):]
			continue
		}

		r, size := utf8.DecodeRuneInString(remaining)
		remaining = remaining[size:]
		if r == '(' && !inParentheses {
			if currentText != "" {
				result.Parts = append(result.Parts, roleCardTextPart{Text: currentText})
				currentText = ""
			}
			inParentheses = true
			currentText = string(r)
		} else if r == ')' && inParentheses {
			currentText += string(r)
			result.Parts = append(result.Parts, roleCardTextPart{
				Text:   currentText,
				Italic: true,
			})
			currentText = ""
			inParentheses = false
		} else {
			currentText += string(r)
		}
	}

	if currentText != "" {
		result.Parts = append(result.Parts, roleCardTextPart{
			Text:   currentText,
			Italic: inParentheses,
		})
	}
	return result
}

func escapeRoleCardHTML(text string) string {
	return html.EscapeString(text)
}
