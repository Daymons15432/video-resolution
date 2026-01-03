package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"kiourin-studio/video-resolution/internal/encoder"
	"kiourin-studio/video-resolution/internal/ffmpeg"
	"kiourin-studio/video-resolution/internal/logger"
	"kiourin-studio/video-resolution/internal/probe"
	"kiourin-studio/video-resolution/internal/scaler"
)

const Version = "1.1"

func showVersion() {
	fmt.Printf("Video Resolution (vr) - Version %s\n", Version)
}

func showHelp() {
	fmt.Println("Video Resolution (vr) - Kiourin Studio")
	fmt.Println("====================================")
	fmt.Println("\nUsage: vr [OPTIONS] <input-file> [profile]")
	fmt.Println("       vr [OPTIONS] <scale-mode> <input-file> [profile]")
	fmt.Println("\nOptions:")
	fmt.Println("  -cpu                Force CPU encoding")
	fmt.Println("  -nvidia, -nv        Force NVIDIA GPU encoding")
	fmt.Println("  -intel, -qsv        Force Intel Quick Sync (iGPU)")
	fmt.Println("  -amd                Force AMD GPU encoding")
	fmt.Println("  -gpu                Force any available GPU (auto-detect)")
	fmt.Println("  -igpu               Force integrated GPU (Intel/AMD)")
	fmt.Println("  -compress           Compress video (reduce bitrate)")
	fmt.Println("  -list-gpus          List available GPU encoders")
	fmt.Println("  -v, -version        Show version information")
	fmt.Println("  -h, -help           Show this help message")
	fmt.Println("\nScale Modes (optional):")
	fmt.Println("  -ds                 Downscale video")
	fmt.Println("  -us                 Upscale video")
	fmt.Println("\nProfiles (optional, default: med):")
	fmt.Println("  low                 Fast encoding, lower quality")
	fmt.Println("  med                 Balanced encoding")
	fmt.Println("  high                Slow encoding, highest quality")
	fmt.Println("\nExamples:")
	fmt.Println("  vr video.mp4                     # Compress only (no scaling)")
	fmt.Println("  vr -compress video.mp4           # Compress with auto-detect encoder")
	fmt.Println("  vr -ds video.mp4 high            # Downscale with high profile")
	fmt.Println("  vr -cpu -ds video.mp4            # Force CPU encoding")
	fmt.Println("  vr -nvidia -us video.mp4 low     # Force NVIDIA encoding")
	fmt.Println("  vr -intel -compress video.mp4    # Compress using Intel iGPU")
	fmt.Println("  vr -amd video.mp4                # Compress using AMD GPU")
	fmt.Println("  vr -list-gpus                    # Show available GPUs")
	fmt.Println("  vr -v                            # Show version")
	fmt.Println("  vr -h                            # Show help")
}

func parseArgs() (scaleMode, input, profileStr string, gpuMode string, compress bool, showVersionFlag, showHelpFlag, listGpusFlag bool) {
	args := os.Args[1:]
	gpuMode = "auto"
	compress = false
	showVersionFlag = false
	showHelpFlag = false
	listGpusFlag = false

	foundInput := false
	foundScaleMode := false

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "-h", "-help":
			showHelpFlag = true
			return
		case "-v", "-version":
			showVersionFlag = true
			return
		case "-list-gpus":
			listGpusFlag = true
			return
		case "-cpu":
			gpuMode = "cpu"
		case "-nvidia", "-nv":
			gpuMode = "nvidia"
		case "-intel", "-qsv":
			gpuMode = "intel"
		case "-amd":
			gpuMode = "amd"
		case "-gpu":
			gpuMode = "gpu"
		case "-igpu":
			gpuMode = "igpu"
		case "-compress":
			compress = true
		case "-ds", "-us":
			if !foundScaleMode {
				scaleMode = arg
				foundScaleMode = true
			}
		default:
			if !strings.HasPrefix(arg, "-") && !foundInput {
				if arg == "low" || arg == "med" || arg == "high" {
					if profileStr == "" {
						profileStr = arg
					}
				} else {
					input = arg
					foundInput = true
				}
			}
		}
	}

	return scaleMode, input, profileStr, gpuMode, compress, showVersionFlag, showHelpFlag, listGpusFlag
}

func listAvailableGPUs() {
	fmt.Println("Available GPU Encoders:")
	fmt.Println("=======================")
	fmt.Printf("Video Resizer v%s\n\n", Version)

	gpus := ffmpeg.GetAvailableGPUs()
	for _, gpu := range gpus {
		fmt.Printf("  - %s\n", strings.ToUpper(string(gpu)))
	}

	fmt.Println("\nEncoders detected:")
	cmd := exec.Command("ffmpeg", "-encoders")
	out, _ := cmd.Output()

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "264") || strings.Contains(line, "265") {
			if strings.Contains(line, "nvenc") ||
				strings.Contains(line, "qsv") ||
				strings.Contains(line, "amf") ||
				strings.Contains(line, "vaapi") {
				fmt.Println("  " + strings.TrimSpace(line))
			}
		}
	}
}

func main() {
	scaleModeFlag, input, profileStr, gpuMode, compress, showVersionFlag, showHelpFlag, listGpusFlag := parseArgs()
	if showVersionFlag {
		showVersion()
		return
	}
	if showHelpFlag {
		showHelp()
		return
	}
	if listGpusFlag {
		if err := ffmpeg.Init(); err != nil {
			fmt.Println("Error: Failed to find FFmpeg in PATH. Please install FFmpeg first.")
			return
		}
		listAvailableGPUs()
		return
	}

	if input == "" {
		if len(os.Args) == 1 {
			showHelp()
		} else {
			fmt.Println("\nError: Missing input file")
			showHelp()
		}
		return
	}

	logger.Info("Init", "Preparing engine...")
	time.Sleep(300 * time.Millisecond)

	if err := ffmpeg.Init(); err != nil {
		logger.Info("Error", "Failed to find FFmpeg in PATH. Please install FFmpeg first.")
		return
	}
	logger.Info("Init", "FFmpeg ready")

	if _, err := os.Stat(input); os.IsNotExist(err) {
		logger.Info("Error", fmt.Sprintf("File not found: %s", input))
		return
	}

	profile := encoder.Med
	if profileStr != "" {
		profile = encoder.ParseProfile(profileStr)
	}

	var detectedGPU ffmpeg.GPU
	if gpuMode != "auto" {
		logger.Info("Mode", fmt.Sprintf("Forcing %s encoding...", gpuMode))
		detectedGPU = ffmpeg.SetForcedGPU(gpuMode)

		if detectedGPU == ffmpeg.CPU && gpuMode != "cpu" {
			logger.Info("Warning",
				fmt.Sprintf("%s encoder not available, falling back to CPU", gpuMode))

			available := ffmpeg.GetAvailableGPUs()
			if len(available) > 1 {
				logger.Info("Info", "Available encoders:")
				for _, gpu := range available {
					logger.Info("      ", string(gpu))
				}
			}
		}
	} else {
		logger.Info("Mode", "Auto-detecting best encoder...")
		detectedGPU = ffmpeg.DetectGPU()
	}

	logger.Info("Scan", "Reading video info...")
	res, err := probe.ResolutionOf(input)
	if err != nil {
		logger.Info("Error", "Cannot read video")
		return
	}
	dur, _ := probe.Duration(input)

	logger.Info("Scan", fmt.Sprintf("Resolution: %dx%d", res.W, res.H))
	if dur > 0 {
		logger.Info("Scan", fmt.Sprintf("Duration: %.0f sec", dur))
	}

	mode := "none"
	if scaleModeFlag != "" {
		mode = map[string]string{"-ds": "down", "-us": "up"}[scaleModeFlag]
	}

	var target scaler.Resolution
	if mode != "none" {
		target = scaler.Auto(
			scaler.Resolution{W: res.W, H: res.H},
			mode,
		)
	} else {
		target = scaler.Resolution{W: res.W, H: res.H}
		logger.Info("Plan", "Mode: No scaling (compress only)")
	}

	gpuName := string(detectedGPU)
	if detectedGPU == ffmpeg.CPU {
		gpuName = "CPU (software)"
	}
	logger.Info("GPU ", strings.ToUpper(gpuName)+" detected")

	if mode != "none" {
		logger.Info("Plan", "Mode: "+map[string]string{"up": "Upscale", "down": "Downscale"}[mode])
	}
	logger.Info("Plan", "Profile: "+string(profile))
	if compress {
		logger.Info("Plan", "Compression: ON")
	}
	logger.Info("Plan", fmt.Sprintf("Target: %dx%d", target.W, target.H))

	enc := encoder.Auto(profile)

	if compress {
		enc = encoder.ApplyCompression(enc, detectedGPU, profile)
	}

	baseName := input
	extensions := []string{".mp4", ".mov", ".avi", ".mkv", ".webm", ".flv", ".wmv"}
	for _, ext := range extensions {
		if strings.HasSuffix(strings.ToLower(input), ext) {
			baseName = strings.TrimSuffix(input, ext)
			break
		}
	}

	suffixes := []string{}

	if mode != "none" {
		suffixes = append(suffixes, fmt.Sprintf("%dx%d", target.W, target.H))
	}
	if compress {
		suffixes = append(suffixes, "compressed")
	}

	output := baseName
	if len(suffixes) > 0 {
		output += "-" + strings.Join(suffixes, "-")
	}
	output += ".mp4"

	if _, err := os.Stat(output); err == nil {
		logger.Info("Warning", fmt.Sprintf("Output file already exists: %s", output))
		logger.Info("Info", "It will be overwritten automatically")
	}

	logger.Info("Run ", "Encoding started...")

	argsEnc := []string{
		"-y",
		"-i", input,
	}

	if mode != "none" {
		argsEnc = append(argsEnc, "-vf", fmt.Sprintf("scale=%d:%d:flags=lanczos", target.W, target.H))
	}

	argsEnc = append(argsEnc, "-c:v", enc.Codec)
	argsEnc = append(argsEnc, enc.Params...)
	argsEnc = append(argsEnc,
		"-pix_fmt", "yuv420p",
		"-c:a", "copy",
		"-movflags", "+faststart",
		"-progress", "pipe:1",
		"-nostats",
		"-loglevel", "error",
		output,
	)

	logger.Info("Debug", fmt.Sprintf("Codec: %s", enc.Codec))
	logger.Info("Debug", fmt.Sprintf("Params: %v", enc.Params))

	cmd := exec.Command("ffmpeg", argsEnc...)
	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		logger.Info("Error", fmt.Sprintf("Failed to start FFmpeg: %v", err))

		if detectedGPU != ffmpeg.CPU {
			logger.Info("Warning", "GPU encoding failed, trying CPU fallback...")
			ffmpeg.SetForcedGPU("cpu")
			enc = encoder.Auto(profile)

			if compress {
				enc = encoder.ApplyCompression(enc, ffmpeg.CPU, profile)
			}

			argsEnc = []string{
				"-y", "-i", input,
			}
			if mode != "none" {
				argsEnc = append(argsEnc, "-vf", fmt.Sprintf("scale=%d:%d:flags=lanczos", target.W, target.H))
			}
			argsEnc = append(argsEnc, "-c:v", enc.Codec)
			argsEnc = append(argsEnc, enc.Params...)
			argsEnc = append(argsEnc,
				"-pix_fmt", "yuv420p",
				"-c:a", "copy",
				"-movflags", "+faststart",
				"-progress", "pipe:1",
				"-nostats",
				"-loglevel", "error",
				output,
			)

			cmd = exec.Command("ffmpeg", argsEnc...)
			stdout, _ = cmd.StdoutPipe()
			cmd.Stderr = os.Stderr

			if err := cmd.Start(); err != nil {
				logger.Info("Error", fmt.Sprintf("CPU fallback also failed: %v", err))
				return
			}
			logger.Info("Info", "Using CPU encoder as fallback")
		} else {
			return
		}
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "out_time_ms=") {
			ms, err := strconv.ParseFloat(strings.TrimPrefix(line, "out_time_ms="), 64)
			if err == nil && dur > 0 {
				p := (ms / 1_000_000) / dur * 100
				if p > 100 {
					p = 100
				}
				logger.Inline(fmt.Sprintf("Progress: %.1f%%", p))
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		logger.Info("Error", fmt.Sprintf("Encoding failed: %v", err))
		fmt.Println()
		return
	}

	fmt.Println()
	logger.Info("Done", fmt.Sprintf("Saved as %s", output))

	operation := "Compressed"
	if mode != "none" {
		operation = map[string]string{"up": "Upscaled", "down": "Downscaled"}[mode]
		if compress {
			operation += " and compressed"
		}
	}

	logger.Info("Info", fmt.Sprintf("Operation: %s", operation))
	logger.Info("Info", fmt.Sprintf("Original: %dx%d → Target: %dx%d",
		res.W, res.H, target.W, target.H))
	logger.Info("Info", fmt.Sprintf("Encoder: %s", enc.Codec))
	if compress {
		logger.Info("Info", "Compression: Applied")
	}

	ffmpeg.ResetForcedGPU()
}
