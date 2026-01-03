package ffmpeg

import (
	"os/exec"
	"strings"
)

type GPU string

const (
	NVIDIA GPU = "nvidia"
	INTEL  GPU = "intel"
	AMD    GPU = "amd"
	CPU    GPU = "cpu"
)

var (
	forcedGPU GPU = ""
)

func DetectGPU() GPU {
	if forcedGPU != "" {
		return forcedGPU
	}

	cmd := exec.Command("ffmpeg", "-encoders")
	out, err := cmd.Output()
	if err != nil {
		return CPU
	}

	output := string(out)

	if strings.Contains(output, "h264_nvenc") ||
		strings.Contains(output, "hevc_nvenc") {
		return NVIDIA
	}

	if strings.Contains(output, "h264_qsv") ||
		strings.Contains(output, "hevc_qsv") {
		return INTEL
	}

	if strings.Contains(output, "h264_amf") ||
		strings.Contains(output, "hevc_amf") {
		return AMD
	}

	if strings.Contains(output, "h264_vaapi") ||
		strings.Contains(output, "hevc_vaapi") {
		vendorCmd := exec.Command("lspci", "|", "grep", "-i", "vga")
		vendorOut, _ := vendorCmd.Output()
		vendorStr := string(vendorOut)

		if strings.Contains(vendorStr, "Intel") {
			return INTEL
		} else if strings.Contains(vendorStr, "AMD") ||
			strings.Contains(vendorStr, "ATI") {
			return AMD
		}
	}

	return CPU
}

func HasEncoder(encoderName string) bool {
	cmd := exec.Command("ffmpeg", "-encoders")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), encoderName)
}

func SetForcedGPU(mode string) GPU {
	switch mode {
	case "nvidia":
		if HasEncoder("h264_nvenc") || HasEncoder("hevc_nvenc") {
			forcedGPU = NVIDIA
			return NVIDIA
		}
		forcedGPU = CPU
		return CPU

	case "intel", "qsv":
		if HasEncoder("h264_qsv") || HasEncoder("hevc_qsv") {
			forcedGPU = INTEL
			return INTEL
		}
		if HasEncoder("h264_vaapi") || HasEncoder("hevc_vaapi") {
			forcedGPU = INTEL
			return INTEL
		}
		forcedGPU = CPU
		return CPU

	case "amd":
		if HasEncoder("h264_amf") || HasEncoder("hevc_amf") {
			forcedGPU = AMD
			return AMD
		}
		if HasEncoder("h264_vaapi") || HasEncoder("hevc_vaapi") {
			forcedGPU = AMD
			return AMD
		}
		forcedGPU = CPU
		return CPU

	case "gpu", "igpu":
		if HasEncoder("h264_nvenc") || HasEncoder("hevc_nvenc") {
			forcedGPU = NVIDIA
			return NVIDIA
		}
		if HasEncoder("h264_qsv") || HasEncoder("hevc_qsv") {
			forcedGPU = INTEL
			return INTEL
		}
		if HasEncoder("h264_amf") || HasEncoder("hevc_amf") {
			forcedGPU = AMD
			return AMD
		}
		if HasEncoder("h264_vaapi") || HasEncoder("hevc_vaapi") {
			forcedGPU = INTEL
			return INTEL
		}
		forcedGPU = CPU
		return CPU

	case "cpu":
		forcedGPU = CPU
		return CPU

	default:
		forcedGPU = ""
		return DetectGPU()
	}
}

func ResetForcedGPU() {
	forcedGPU = ""
}

func GetAvailableGPUs() []GPU {
	var gpus []GPU

	cmd := exec.Command("ffmpeg", "-encoders")
	out, err := cmd.Output()
	if err != nil {
		return []GPU{CPU}
	}

	output := string(out)

	if strings.Contains(output, "h264_nvenc") ||
		strings.Contains(output, "hevc_nvenc") {
		gpus = append(gpus, NVIDIA)
	}

	if strings.Contains(output, "h264_qsv") ||
		strings.Contains(output, "hevc_qsv") {
		gpus = append(gpus, INTEL)
	}

	if strings.Contains(output, "h264_amf") ||
		strings.Contains(output, "hevc_amf") {
		gpus = append(gpus, AMD)
	}

	if strings.Contains(output, "h264_vaapi") ||
		strings.Contains(output, "hevc_vaapi") {
		vaapiAdded := false
		for _, g := range gpus {
			if g == INTEL || g == AMD {
				vaapiAdded = true
				break
			}
		}
		if !vaapiAdded {
			gpus = append(gpus, INTEL)
		}
	}

	gpus = append(gpus, CPU)

	return gpus
}
