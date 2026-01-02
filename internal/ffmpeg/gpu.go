package ffmpeg

import (
	"os/exec"
	"strings"
)

type GPU string

const (
	NVIDIA GPU = "nvidia"
	CPU    GPU = "cpu"
)

func DetectGPU() GPU {
	cmd := exec.Command("ffmpeg", "-encoders")
	out, err := cmd.Output()
	if err != nil {
		return CPU
	}
	if strings.Contains(string(out), "h264_nvenc") {
		return NVIDIA
	}
	return CPU
}
