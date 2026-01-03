package encoder

import (
	"kiourin-studio/video-resolution/internal/ffmpeg"
	"strings"
)

type Config struct {
	Codec  string
	Params []string
}

func Auto(profile Profile) Config {
	gpu := ffmpeg.DetectGPU()

	switch gpu {
	case ffmpeg.NVIDIA:
		return nvidiaConfig(profile)
	case ffmpeg.INTEL:
		return intelConfig(profile)
	case ffmpeg.AMD:
		return amdConfig(profile)
	default:
		return cpuConfig(profile)
	}
}

func ApplyCompression(config Config, gpu ffmpeg.GPU, profile Profile) Config {
	newConfig := Config{
		Codec:  config.Codec,
		Params: make([]string, len(config.Params)),
	}
	copy(newConfig.Params, config.Params)

	switch gpu {
	case ffmpeg.NVIDIA:
		compressionValue := "28"
		if profile == Low {
			compressionValue = "32"
		} else if profile == High {
			compressionValue = "24"
		}

		for i, param := range newConfig.Params {
			if param == "-cq" && i+1 < len(newConfig.Params) {
				newConfig.Params[i+1] = compressionValue
				return newConfig
			}
		}
		newConfig.Params = append(newConfig.Params, "-cq", compressionValue)

	case ffmpeg.INTEL:
		if strings.Contains(config.Codec, "qsv") {
			qualityValue := "28"
			if profile == Low {
				qualityValue = "32"
			} else if profile == High {
				qualityValue = "24"
			}

			for i, param := range newConfig.Params {
				if param == "-global_quality" && i+1 < len(newConfig.Params) {
					newConfig.Params[i+1] = qualityValue
					return newConfig
				}
			}
			newConfig.Params = append(newConfig.Params, "-global_quality", qualityValue)
		} else if strings.Contains(config.Codec, "vaapi") {
			qpValue := "28"
			if profile == Low {
				qpValue = "32"
			} else if profile == High {
				qpValue = "24"
			}

			for i, param := range newConfig.Params {
				if param == "-qp" && i+1 < len(newConfig.Params) {
					newConfig.Params[i+1] = qpValue
					return newConfig
				}
			}
			newConfig.Params = append(newConfig.Params, "-qp", qpValue)
		}

	case ffmpeg.AMD:
		if strings.Contains(config.Codec, "amf") {
			qpValue := "28"
			if profile == Low {
				qpValue = "32"
			} else if profile == High {
				qpValue = "24"
			}

			for i := 0; i < len(newConfig.Params); i++ {
				if (newConfig.Params[i] == "-qp_i" || newConfig.Params[i] == "-qp_p") &&
					i+1 < len(newConfig.Params) {
					newConfig.Params[i+1] = qpValue
				}
			}
		} else if strings.Contains(config.Codec, "vaapi") {
			qpValue := "28"
			if profile == Low {
				qpValue = "32"
			} else if profile == High {
				qpValue = "24"
			}

			for i, param := range newConfig.Params {
				if param == "-qp" && i+1 < len(newConfig.Params) {
					newConfig.Params[i+1] = qpValue
					return newConfig
				}
			}
			newConfig.Params = append(newConfig.Params, "-qp", qpValue)
		}

	default:
		crfValue := "28"
		if profile == Low {
			crfValue = "32"
		} else if profile == High {
			crfValue = "24"
		}

		for i, param := range newConfig.Params {
			if param == "-crf" && i+1 < len(newConfig.Params) {
				newConfig.Params[i+1] = crfValue
				return newConfig
			}
		}
		newConfig.Params = append(newConfig.Params, "-crf", crfValue)
	}

	return newConfig
}

func nvidiaConfig(profile Profile) Config {
	switch profile {
	case Low:
		return Config{"h264_nvenc", []string{
			"-preset", "p3",
			"-rc", "vbr",
			"-cq", "23",
			"-b_ref_mode", "0",
		}}
	case High:
		return Config{"h264_nvenc", []string{
			"-preset", "p7",
			"-rc", "vbr",
			"-cq", "14",
			"-tune", "hq",
			"-multipass", "fullres",
			"-b:v", "0",
			"-b_ref_mode", "2",
		}}
	default:
		return Config{"h264_nvenc", []string{
			"-preset", "p5",
			"-rc", "vbr",
			"-cq", "18",
			"-tune", "hq",
			"-b_ref_mode", "1",
		}}
	}
}

func intelConfig(profile Profile) Config {
	if ffmpeg.HasEncoder("h264_qsv") {
		switch profile {
		case Low:
			return Config{"h264_qsv", []string{
				"-preset", "fast",
				"-global_quality", "23",
				"-look_ahead", "0",
			}}
		case High:
			return Config{"h264_qsv", []string{
				"-preset", "slow",
				"-global_quality", "16",
				"-look_ahead", "1",
				"-extbrc", "1",
			}}
		default:
			return Config{"h264_qsv", []string{
				"-preset", "medium",
				"-global_quality", "20",
				"-look_ahead", "1",
			}}
		}
	}

	switch profile {
	case Low:
		return Config{"h264_vaapi", []string{
			"-compression_level", "1",
			"-qp", "23",
			"-quality", "speed",
		}}
	case High:
		return Config{"h264_vaapi", []string{
			"-compression_level", "7",
			"-qp", "16",
			"-quality", "quality",
		}}
	default:
		return Config{"h264_vaapi", []string{
			"-compression_level", "3",
			"-qp", "20",
			"-quality", "balanced",
		}}
	}
}

func amdConfig(profile Profile) Config {
	if ffmpeg.HasEncoder("h264_amf") {
		switch profile {
		case Low:
			return Config{"h264_amf", []string{
				"-usage", "ultralowlatency",
				"-quality", "speed",
				"-qp_i", "23",
				"-qp_p", "23",
			}}
		case High:
			return Config{"h264_amf", []string{
				"-usage", "transcoding",
				"-quality", "quality",
				"-qp_i", "16",
				"-qp_p", "16",
				"-preanalysis", "1",
			}}
		default:
			return Config{"h264_amf", []string{
				"-usage", "transcoding",
				"-quality", "balanced",
				"-qp_i", "20",
				"-qp_p", "20",
			}}
		}
	}

	switch profile {
	case Low:
		return Config{"h264_vaapi", []string{
			"-compression_level", "1",
			"-qp", "23",
			"-quality", "speed",
		}}
	case High:
		return Config{"h264_vaapi", []string{
			"-compression_level", "7",
			"-qp", "16",
			"-quality", "quality",
		}}
	default:
		return Config{"h264_vaapi", []string{
			"-compression_level", "3",
			"-qp", "20",
			"-quality", "balanced",
		}}
	}
}

func cpuConfig(profile Profile) Config {
	switch profile {
	case Low:
		return Config{"libx264", []string{
			"-preset", "fast",
			"-crf", "23",
			"-tune", "fastdecode",
		}}
	case High:
		return Config{"libx264", []string{
			"-preset", "veryslow",
			"-crf", "14",
			"-tune", "film",
			"-x264-params", "ref=6:bframes=8",
		}}
	default:
		return Config{"libx264", []string{
			"-preset", "slow",
			"-crf", "16",
			"-tune", "film",
		}}
	}
}
