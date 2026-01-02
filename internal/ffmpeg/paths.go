package ffmpeg

import "os/exec"

func Init() error {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return err
	}

	if _, err := exec.LookPath("ffprobe"); err != nil {
		return err
	}

	return nil
}

func Bin(name string) string {
	return name
}
