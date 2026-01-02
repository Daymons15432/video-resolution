package probe

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"
)

type Resolution struct {
	W int
	H int
}

func ResolutionOf(path string) (Resolution, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=p=0:s=x",
		path,
	)

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return Resolution{}, err
	}

	p := strings.Split(strings.TrimSpace(out.String()), "x")
	w, _ := strconv.Atoi(p[0])
	h, _ := strconv.Atoi(p[1])

	return Resolution{W: w, H: h}, nil
}

func Duration(path string) (float64, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	)

	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
}
