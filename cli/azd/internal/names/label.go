// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package names

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var rfc1123LabelRegex = regexp.MustCompile(`^(?:[A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9-]{0,61}[A-Za-z0-9])$`)

// ValidateLabelName checks if the given name is a valid RFC 1123 Label name.
func ValidateLabelName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	if len(name) > 63 {
		return errors.New("name must be 63 characters or less")
	}

	if !isLowerAsciiAlphaNumeric(rune(name[0])) {
		return errors.New("name must start with a lower-cased alphanumeric character, i.e. a-z or 0-9")
	}

	if !isLowerAsciiAlphaNumeric(rune(name[len(name)-1])) {
		return errors.New("name must end with a lower-cased alphanumeric character, i.e. a-z or 0-9")
	}

	if !rfc1123LabelRegex.MatchString(name) {
		return errors.New("name must contain only lower-cased alphanumeric characters or '-', i.e. a-z, 0-9, or '-'")
	}

	return nil
}

// ValidateProjectName checks if the given name is a valid project name.
// Only lowercase letters, numbers, and - are allowed.
// And the first and last characters must be letters or numbers.
func ValidateProjectName(name string) error {
	if l := len(name); l < 2 || l > 63 {
		return fmt.Errorf("invalid project name '%s': length must be between 2 and 63 characters.", name)
	}

	matched, err := regexp.MatchString(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`, name)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf(
			"invalid project name '%s': only lowercase letters, numbers, and hyphens (-) are allowed, "+
				"and must start and end with a letter or number.",
			name,
		)
	}
	return nil
}

//cspell:disable

// LabelName cleans up a string to be used as a RFC 1123 Label name.
// It does not enforce the 63 character limit.
//
// RFC 1123 Label name:
//   - contain only lowercase alphanumeric characters or '-'
//   - start with an alphanumeric character
//   - end with an alphanumeric character
//
// Examples:
//   - myproject, MYPROJECT -> myproject
//   - myProject, myProjecT, MyProject, MyProjecT -> my-project
//   - my.project, My.Project, my-project, My-Project -> my-project
func LabelName(name string) string {
	hasSeparator, n := cleanAlphaNumeric(name)
	if hasSeparator {
		return labelNameFromSeparators(n)
	}

	return labelNameFromCasing(name)
}

//cspell:enable

// cleanAlphaNumeric removes non-alphanumeric characters from the name.
//
// It also returns whether the name uses word separators.
func cleanAlphaNumeric(name string) (hasSeparator bool, cleaned string) {
	sb := strings.Builder{}
	hasSeparator = false
	for _, c := range name {
		if isAsciiAlphaNumeric(c) {
			sb.WriteRune(c)
		} else if isSeparator(c) {
			hasSeparator = true
			sb.WriteRune(c)
		}
	}

	return hasSeparator, sb.String()
}

func isLowerAsciiAlphaNumeric(r rune) bool {
	return ('0' <= r && r <= '9') || ('a' <= r && r <= 'z')
}

func isAsciiAlphaNumeric(r rune) bool {
	return ('0' <= r && r <= '9') || ('A' <= r && r <= 'Z') || ('a' <= r && r <= 'z')
}

func isSeparator(r rune) bool {
	return r == '-' || r == '_' || r == '.'
}

func lowerCase(r rune) rune {
	if 'A' <= r && r <= 'Z' {
		r += 'a' - 'A'
	}
	return r
}

// Converts camel-cased or Pascal-cased names into lower-cased dash-separated names.
// Example: MyProject, myProject -> my-project
func labelNameFromCasing(name string) string {
	result := strings.Builder{}
	// previously seen upper-case character
	prevUpperCase := -2 // -2 to avoid matching the first character

	for i, c := range name {
		if 'A' <= c && c <= 'Z' {
			if prevUpperCase == i-1 { // handle runs of upper-case word
				prevUpperCase = i
				result.WriteRune(lowerCase(c))
				continue
			}

			if i > 0 && i != len(name)-1 {
				result.WriteRune('-')
			}

			prevUpperCase = i
		}

		if isAsciiAlphaNumeric(c) {
			result.WriteRune(lowerCase(c))
		}
	}

	return result.String()
}

// Converts all word-separated names into lower-cased dash-separated names.
// Examples: my.project, my_project, My-Project -> my-project
func labelNameFromSeparators(name string) string {
	result := strings.Builder{}
	for i, c := range name {
		if isAsciiAlphaNumeric(c) {
			result.WriteRune(lowerCase(c))
		} else if i > 0 && i != len(name)-1 && isSeparator(c) {
			result.WriteRune('-')
		}
	}

	return strings.Trim(result.String(), "-")
}
