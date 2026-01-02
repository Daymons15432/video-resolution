package encoder

import "kiourin-studio/video-resolution/internal/ffmpeg"

type Config struct {
	Codec  string
	Params []string
}

func Auto(profile Profile) Config {
	if ffmpeg.DetectGPU() == ffmpeg.NVIDIA {
		switch profile {
		case Low:
			return Config{"h264_nvenc", []string{
				"-preset", "p3", "-rc", "vbr", "-cq", "23",
			}}
		case High:
			return Config{"h264_nvenc", []string{
				"-preset", "p7",
				"-rc", "vbr",
				"-cq", "14",
				"-tune", "hq",
				"-multipass", "fullres",
				"-b:v", "0",
			}}
		default:
			return Config{"h264_nvenc", []string{
				"-preset", "p5", "-rc", "vbr", "-cq", "18", "-tune", "hq",
			}}
		}
	}

	switch profile {
	case Low:
		return Config{"libx264", []string{"-preset", "fast", "-crf", "23"}}
	case High:
		return Config{"libx264", []string{"-preset", "veryslow", "-crf", "14"}}
	default:
		return Config{"libx264", []string{"-preset", "slow", "-crf", "16"}}
	}
}
