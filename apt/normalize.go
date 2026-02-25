package apt

import "strings"

func normalizeName(name string) string {
	n, _ := parseArch(name)
	return n
}

func parseArch(name string) (string, string) {
	if idx := strings.LastIndex(name, ":"); idx >= 0 {
		pkg := name[:idx]
		arch := name[idx+1:]
		switch arch {
		case "amd64", "i386", "arm64", "armhf", "armel", "mips", "mipsel",
			"mips64el", "ppc64el", "s390x", "all", "any":
			return pkg, arch
		}
	}
	return name, ""
}
