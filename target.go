package snack

// TargetNames extracts just the package names from a slice of targets.
func TargetNames(targets []Target) []string {
	names := make([]string, len(targets))
	for i, t := range targets {
		names[i] = t.Name
	}
	return names
}
