package telegram

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func isURL(text string) bool {
	re := regexp.MustCompile(`^https?://[^\s]+$`)
	return re.MatchString(strings.TrimSpace(text))
}

func parseInt(s string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(s))
}

func filepathFromCareerOps(careerOpsPath, subdir string) string {
	abs, err := filepath.Abs(careerOpsPath)
	if err != nil {
		return careerOpsPath
	}
	return filepath.Join(abs, subdir)
}
