package winget

import (
	"strconv"
	"strings"

	"github.com/gogrlx/snack"
)

// parseTable parses winget tabular output (list, search, upgrade).
//
// Winget uses fixed-width columns whose positions are determined by a
// header row with dashes (e.g. "----  ------  -------").
// The header names vary by locale, so we detect columns positionally.
//
// Typical `winget list` output:
//
//	Name            Id                     Version   Available Source
//	--------------------------------------------------------------
//	Visual Studio   Microsoft.VisualStudio 17.8.0    17.9.0    winget
//
// Typical `winget search` output:
//
//	Name            Id                     Version  Match        Source
//	--------------------------------------------------------------
//	Visual Studio   Microsoft.VisualStudio 17.9.0                winget
//
// When installed is true, we mark all parsed packages as Installed.
func parseTable(output string, installed bool) []snack.Package {
	lines := strings.Split(output, "\n")

	// Find the separator line (all dashes/spaces) to determine column positions.
	sepIdx := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if isSeparatorLine(trimmed) {
			sepIdx = i
			break
		}
	}
	if sepIdx < 1 {
		return nil
	}

	// Use the header line (just above separator) to determine column starts.
	header := lines[sepIdx-1]
	cols := detectColumns(header)
	if len(cols) < 2 {
		return nil
	}

	var pkgs []snack.Package
	for _, line := range lines[sepIdx+1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Skip footer lines like "X upgrades available."
		if isFooterLine(line) {
			continue
		}
		fields := extractFields(line, cols)
		if len(fields) < 2 {
			continue
		}
		pkg := snack.Package{
			Name:      fields[0],
			Installed: installed,
		}
		// Column order: Name, Id, Version, [Available], [Source]
		// For search: Name, Id, Version, [Match], [Source]
		if len(fields) >= 3 {
			pkg.Version = fields[2]
		}
		// Use the ID as the package name for consistency (winget uses
		// Publisher.Package IDs as the canonical identifier).
		if len(fields) >= 2 && fields[1] != "" {
			pkg.Description = fields[0] // keep display name as description
			pkg.Name = fields[1]
		}
		// If there's a Source column (typically the last), use it.
		if len(fields) >= 5 && fields[len(fields)-1] != "" {
			pkg.Repository = fields[len(fields)-1]
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}

// isSeparatorLine returns true if the line is a column separator
// (composed entirely of dashes and spaces, with at least some dashes).
func isSeparatorLine(line string) bool {
	hasDash := false
	for _, c := range line {
		switch c {
		case '-':
			hasDash = true
		case ' ', '\t':
			// allowed
		default:
			return false
		}
	}
	return hasDash
}

// detectColumns returns the starting index of each column based on
// the header line. Columns are separated by 2+ spaces.
func detectColumns(header string) []int {
	var cols []int
	inWord := false
	for i, c := range header {
		if c != ' ' && c != '\t' {
			if !inWord {
				cols = append(cols, i)
				inWord = true
			}
		} else {
			// Need at least 2 spaces to end a column
			if inWord && i+1 < len(header) && (header[i+1] == ' ' || header[i+1] == '\t') {
				inWord = false
			} else if inWord && i+1 >= len(header) {
				inWord = false
			}
		}
	}
	return cols
}

// extractFields splits a data line according to detected column positions.
func extractFields(line string, cols []int) []string {
	fields := make([]string, len(cols))
	for i, start := range cols {
		if start >= len(line) {
			break
		}
		end := len(line)
		if i+1 < len(cols) {
			end = min(cols[i+1], len(line))
		}
		fields[i] = strings.TrimSpace(line[start:end])
	}
	return fields
}

// isFooterLine returns true for winget output footer lines.
func isFooterLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return true
	}
	lower := strings.ToLower(trimmed)
	if strings.Contains(lower, "upgrades available") ||
		strings.Contains(lower, "package(s)") ||
		strings.Contains(lower, "installed package") ||
		strings.HasPrefix(lower, "the following") {
		return true
	}
	return false
}

// parseShow parses `winget show` key-value output into a Package.
//
// Typical output:
//
//	Found Visual Studio Code [Microsoft.VisualStudioCode]
//	Version: 1.85.0
//	Publisher: Microsoft Corporation
//	Description: Code editing. Redefined.
//	...
func parseShow(output string) *snack.Package {
	pkg := &snack.Package{}
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Parse "Found <Name> [<Id>]" header line.
		if strings.HasPrefix(line, "Found ") {
			if idx := strings.LastIndex(line, "["); idx > 0 {
				endIdx := strings.LastIndex(line, "]")
				if endIdx > idx {
					pkg.Name = strings.TrimSpace(line[idx+1 : endIdx])
					pkg.Description = strings.TrimSpace(line[6:idx])
				}
			}
			continue
		}
		before, after, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key := strings.TrimSpace(before)
		val := strings.TrimSpace(after)
		switch strings.ToLower(key) {
		case "version":
			pkg.Version = val
		case "description":
			if pkg.Description == "" {
				pkg.Description = val
			}
		case "publisher":
			// informational only
		}
	}
	if pkg.Name == "" {
		return nil
	}
	return pkg
}

// parseSourceList parses `winget source list` output into Repositories.
//
// Output format:
//
//	Name    Argument
//	----------------------------------------------
//	winget  https://cdn.winget.microsoft.com/cache
//	msstore https://storeedgefd.dsx.mp.microsoft.com/v9.0
func parseSourceList(output string) []snack.Repository {
	lines := strings.Split(output, "\n")
	sepIdx := -1
	for i, line := range lines {
		if isSeparatorLine(strings.TrimSpace(line)) {
			sepIdx = i
			break
		}
	}
	if sepIdx < 0 {
		return nil
	}
	var repos []snack.Repository
	for _, line := range lines[sepIdx+1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) < 2 {
			continue
		}
		repos = append(repos, snack.Repository{
			ID:      fields[0],
			Name:    fields[0],
			URL:     fields[1],
			Enabled: true,
		})
	}
	return repos
}

// stripVT removes ANSI/VT100 escape sequences from a string.
func stripVT(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			// Skip CSI sequence: ESC [ ... final byte
			j := i + 2
			for j < len(s) && s[j] >= 0x20 && s[j] <= 0x3f {
				j++
			}
			if j < len(s) {
				j++ // skip final byte
			}
			i = j
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// isProgressLine returns true for lines that are only progress indicators.
func isProgressLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}
	// Lines containing block characters and/or percentage
	hasBlock := strings.ContainsAny(trimmed, "█▓░")
	hasPercent := strings.Contains(trimmed, "%")
	if hasBlock || (hasPercent && len(trimmed) < 40) {
		return true
	}
	return false
}

// semverCmp does a basic semver-ish comparison.
// Returns -1 if a < b, 0 if equal, 1 if a > b.
func semverCmp(a, b string) int {
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")

	maxLen := max(len(partsB), len(partsA))

	for i := 0; i < maxLen; i++ {
		var numA, numB int
		if i < len(partsA) {
			numA, _ = strconv.Atoi(stripNonNumeric(partsA[i]))
		}
		if i < len(partsB) {
			numB, _ = strconv.Atoi(stripNonNumeric(partsB[i]))
		}
		if numA < numB {
			return -1
		}
		if numA > numB {
			return 1
		}
	}
	return 0
}

// stripNonNumeric keeps only leading digits from a string.
func stripNonNumeric(s string) string {
	for i, c := range s {
		if c < '0' || c > '9' {
			return s[:i]
		}
	}
	return s
}
