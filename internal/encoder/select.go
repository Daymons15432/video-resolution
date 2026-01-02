package encoder

import "kiourin-studio/video-resolution/internal/ffmpeg"

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
	// Try QSV first, fallback to VAAPI
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

	// Fallback to VAAPI
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

	// Fallback to VAAPI for AMD
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
