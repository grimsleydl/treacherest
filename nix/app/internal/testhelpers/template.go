package testhelpers

import (
	"bytes"
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

// TemplateRenderer provides utilities for testing templ components
type TemplateRenderer struct {
	t      *testing.T
	buffer *bytes.Buffer
	html   string
}

// NewTemplateRenderer creates a new template renderer for testing
func NewTemplateRenderer(t *testing.T) *TemplateRenderer {
	return &TemplateRenderer{
		t:      t,
		buffer: &bytes.Buffer{},
	}
}

// Render renders a templ component and stores the HTML
func (r *TemplateRenderer) Render(component templ.Component) *TemplateRenderer {
	r.buffer.Reset()
	err := component.Render(context.Background(), r.buffer)
	if err != nil {
		r.t.Fatalf("Failed to render template: %v", err)
	}
	r.html = r.buffer.String()
	return r
}

// GetHTML returns the rendered HTML
func (r *TemplateRenderer) GetHTML() string {
	return r.html
}

// AssertContains checks if the rendered HTML contains a substring
func (r *TemplateRenderer) AssertContains(substring string) *TemplateRenderer {
	if !strings.Contains(r.html, substring) {
		r.t.Errorf("Expected HTML to contain %q, but it didn't.\nHTML: %s", substring, r.html)
	}
	return r
}

// AssertNotContains checks if the rendered HTML does not contain a substring
func (r *TemplateRenderer) AssertNotContains(substring string) *TemplateRenderer {
	if strings.Contains(r.html, substring) {
		r.t.Errorf("Expected HTML not to contain %q, but it did.\nHTML: %s", substring, r.html)
	}
	return r
}

// AssertMatches checks if the rendered HTML matches a regex pattern
func (r *TemplateRenderer) AssertMatches(pattern string) *TemplateRenderer {
	matched, err := regexp.MatchString(pattern, r.html)
	if err != nil {
		r.t.Fatalf("Invalid regex pattern %q: %v", pattern, err)
	}
	if !matched {
		r.t.Errorf("Expected HTML to match pattern %q, but it didn't.\nHTML: %s", pattern, r.html)
	}
	return r
}

// AssertHasDatastarAttribute checks if an element has a specific data-* attribute
func (r *TemplateRenderer) AssertHasDatastarAttribute(attribute, value string) *TemplateRenderer {
	// Datastar attributes start with data-
	attrName := "data-" + attribute
	pattern := attrName + `="` + regexp.QuoteMeta(value) + `"`
	if !strings.Contains(r.html, pattern) {
		// Also check for single quotes
		pattern2 := attrName + `='` + regexp.QuoteMeta(value) + `'`
		if !strings.Contains(r.html, pattern2) {
			r.t.Errorf("Expected to find attribute %s with value %q, but didn't find it.\nHTML: %s", attrName, value, r.html)
		}
	}
	return r
}

// AssertHasElement checks if the HTML contains a specific element
func (r *TemplateRenderer) AssertHasElement(tagName string) *TemplateRenderer {
	pattern := `<` + tagName + `[\s>]`
	matched, _ := regexp.MatchString(pattern, r.html)
	if !matched {
		r.t.Errorf("Expected to find element <%s>, but didn't find it.\nHTML: %s", tagName, r.html)
	}
	return r
}

// AssertHasElementWithID checks if the HTML contains an element with a specific ID
func (r *TemplateRenderer) AssertHasElementWithID(id string) *TemplateRenderer {
	pattern := `id="` + regexp.QuoteMeta(id) + `"`
	if !strings.Contains(r.html, pattern) {
		// Also check for single quotes
		pattern2 := `id='` + regexp.QuoteMeta(id) + `'`
		if !strings.Contains(r.html, pattern2) {
			r.t.Errorf("Expected to find element with id=%q, but didn't find it.\nHTML: %s", id, r.html)
		}
	}
	return r
}

// AssertHasClass checks if any element has a specific CSS class
func (r *TemplateRenderer) AssertHasClass(className string) *TemplateRenderer {
	// Check for class in various formats
	patterns := []string{
		`class="[^"]*\b` + regexp.QuoteMeta(className) + `\b[^"]*"`,
		`class='[^']*\b` + regexp.QuoteMeta(className) + `\b[^']*'`,
	}

	found := false
	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, r.html)
		if matched {
			found = true
			break
		}
	}

	if !found {
		r.t.Errorf("Expected to find element with class %q, but didn't find it.\nHTML: %s", className, r.html)
	}
	return r
}

// AssertFormAction checks if a form has a specific action attribute
func (r *TemplateRenderer) AssertFormAction(action string) *TemplateRenderer {
	pattern := `<form[^>]*action="` + regexp.QuoteMeta(action) + `"`
	matched, _ := regexp.MatchString(pattern, r.html)
	if !matched {
		// Also check for single quotes
		pattern2 := `<form[^>]*action='` + regexp.QuoteMeta(action) + `'`
		matched2, _ := regexp.MatchString(pattern2, r.html)
		if !matched2 {
			r.t.Errorf("Expected to find form with action=%q, but didn't find it.\nHTML: %s", action, r.html)
		}
	}
	return r
}

// AssertInputValue checks if an input has a specific value
func (r *TemplateRenderer) AssertInputValue(name, value string) *TemplateRenderer {
	// Look for input with name and value
	pattern := `<input[^>]*name="` + regexp.QuoteMeta(name) + `"[^>]*value="` + regexp.QuoteMeta(value) + `"`
	matched, _ := regexp.MatchString(pattern, r.html)
	if !matched {
		// Try different order
		pattern2 := `<input[^>]*value="` + regexp.QuoteMeta(value) + `"[^>]*name="` + regexp.QuoteMeta(name) + `"`
		matched2, _ := regexp.MatchString(pattern2, r.html)
		if !matched2 {
			r.t.Errorf("Expected to find input with name=%q and value=%q, but didn't find it.\nHTML: %s", name, value, r.html)
		}
	}
	return r
}

// CountElements counts how many times an element appears
func (r *TemplateRenderer) CountElements(tagName string) int {
	pattern := `<` + tagName + `[\s>]`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllString(r.html, -1)
	return len(matches)
}

// AssertElementCount checks if an element appears a specific number of times
func (r *TemplateRenderer) AssertElementCount(tagName string, expectedCount int) *TemplateRenderer {
	count := r.CountElements(tagName)
	if count != expectedCount {
		r.t.Errorf("Expected %d <%s> elements, but found %d.\nHTML: %s", expectedCount, tagName, count, r.html)
	}
	return r
}

// AssertNotEmpty checks that the rendered HTML is not empty
func (r *TemplateRenderer) AssertNotEmpty() *TemplateRenderer {
	if len(strings.TrimSpace(r.html)) == 0 {
		r.t.Error("Expected non-empty HTML, but got empty content")
	}
	return r
}

// AssertValid performs basic HTML validation checks
func (r *TemplateRenderer) AssertValid() *TemplateRenderer {
	// Check for basic HTML structure issues
	openTags := regexp.MustCompile(`<(\w+)(?:\s[^>]*)?>`)
	closeTags := regexp.MustCompile(`</(\w+)>`)

	openMatches := openTags.FindAllStringSubmatch(r.html, -1)
	closeMatches := closeTags.FindAllStringSubmatch(r.html, -1)

	// Count tags (basic check, not a full parser)
	tagCounts := make(map[string]int)
	for _, match := range openMatches {
		if len(match) > 1 {
			tag := match[1]
			// Skip void elements
			if !isVoidElement(tag) {
				tagCounts[tag]++
			}
		}
	}

	for _, match := range closeMatches {
		if len(match) > 1 {
			tag := match[1]
			tagCounts[tag]--
		}
	}

	// Check for mismatched tags
	for tag, count := range tagCounts {
		if count != 0 {
			r.t.Errorf("Mismatched tags: <%s> opened %d more times than closed", tag, count)
		}
	}

	return r
}

// isVoidElement checks if an HTML element is self-closing
func isVoidElement(tag string) bool {
	voidElements := []string{
		"area", "base", "br", "col", "embed", "hr", "img", "input",
		"link", "meta", "param", "source", "track", "wbr",
	}
	for _, void := range voidElements {
		if tag == void {
			return true
		}
	}
	return false
}
