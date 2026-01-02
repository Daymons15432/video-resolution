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

func showHelp() {
	fmt.Println("Video Resizer (vr) - Kiourin Studio")
	fmt.Println("====================================")
	fmt.Println("\nUsage: vr [OPTIONS] <scale-mode> <input-file> [profile]")
	fmt.Println("\nOptions:")
	fmt.Println("  -cpu                Force CPU encoding")
	fmt.Println("  -nvidia, -nv        Force NVIDIA GPU encoding")
	fmt.Println("  -intel, -qsv        Force Intel Quick Sync (iGPU)")
	fmt.Println("  -amd                Force AMD GPU encoding")
	fmt.Println("  -gpu                Force any available GPU (auto-detect)")
	fmt.Println("  -igpu               Force integrated GPU (Intel/AMD)")
	fmt.Println("  -list-gpus          List available GPU encoders")
	fmt.Println("  -h, -help           Show this help message")
	fmt.Println("\nScale Modes:")
	fmt.Println("  -ds                 Downscale video")
	fmt.Println("  -us                 Upscale video")
	fmt.Println("\nProfiles (optional, default: med):")
	fmt.Println("  low                 Fast encoding, lower quality")
	fmt.Println("  med                 Balanced encoding")
	fmt.Println("  high                Slow encoding, highest quality")
	fmt.Println("\nExamples:")
	fmt.Println("  vr -ds video.mp4 high              # Auto-detect best encoder")
	fmt.Println("  vr -cpu -ds video.mp4              # Force CPU encoding")
	fmt.Println("  vr -nvidia -us video.mp4 low       # Force NVIDIA encoding")
	fmt.Println("  vr -intel -ds video.mp4 med        # Force Intel iGPU")
	fmt.Println("  vr -amd -us video.mp4              # Force AMD GPU")
	fmt.Println("  vr -list-gpus                      # Show available GPUs")
	fmt.Println("  vr -h                              # Show help")
}

func parseArgs() (scaleMode, input, profileStr string, gpuMode string) {
	args := os.Args[1:]
	gpuMode = "auto" // Default to auto-detect

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "-h", "-help":
			showHelp()
			os.Exit(0)
		case "-list-gpus":
			listAvailableGPUs()
			os.Exit(0)
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
		case "-ds", "-us":
			// This is a scale mode
			if scaleMode == "" && i+1 < len(args) {
				scaleMode = arg
				input = args[i+1]
				i++ // Skip input file

				// Check for profile (next argument if exists and not another flag)
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					profileStr = args[i+1]
					i++
				}
			}
		default:
			// If we haven't found scale mode yet and this doesn't look like a flag
			if scaleMode == "" && !strings.HasPrefix(arg, "-") && i == 0 {
				// Might be scale mode without dash? Show error
				fmt.Printf("Error: Invalid argument '%s'. Scale mode must be -ds or -us\n", arg)
				showHelp()
				os.Exit(1)
			}
		}
	}

	return scaleMode, input, profileStr, gpuMode
}

func listAvailableGPUs() {
	fmt.Println("Available GPU Encoders:")
	fmt.Println("=======================")

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
	logger.Info("Init", "Preparing engine...")
	time.Sleep(300 * time.Millisecond)

	if err := ffmpeg.Init(); err != nil {
		logger.Info("Error", "Failed to find FFmpeg in PATH. Please install FFmpeg first.")
		return
	}
	logger.Info("Init", "FFmpeg ready")

	// Parse arguments
	scaleModeFlag, input, profileStr, gpuMode := parseArgs()

	// Validate required arguments
	if scaleModeFlag == "" || input == "" {
		fmt.Println("\nError: Missing required arguments")
		showHelp()
		return
	}

	// Parse scale mode
	mode := map[string]string{"-ds": "down", "-us": "up"}[scaleModeFlag]
	if mode == "" {
		fmt.Printf("\nError: Invalid scale mode '%s'. Use -ds or -us\n", scaleModeFlag)
		showHelp()
		return
	}

	// Check if input file exists
	if _, err := os.Stat(input); os.IsNotExist(err) {
		logger.Info("Error", fmt.Sprintf("File not found: %s", input))
		return
	}

	// Parse profile
	profile := encoder.Med
	if profileStr != "" {
		profile = encoder.ParseProfile(profileStr)
	}

	// Handle GPU mode
	var detectedGPU ffmpeg.GPU
	if gpuMode != "auto" {
		logger.Info("Mode", fmt.Sprintf("Forcing %s encoding...", gpuMode))
		detectedGPU = ffmpeg.SetForcedGPU(gpuMode)

		// Check if forced GPU is available
		if detectedGPU == ffmpeg.CPU && gpuMode != "cpu" {
			logger.Info("Warning",
				fmt.Sprintf("%s encoder not available, falling back to CPU", gpuMode))

			// List available GPUs
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

	target := scaler.Auto(
		scaler.Resolution{W: res.W, H: res.H},
		mode,
	)

	gpuName := string(detectedGPU)
	if detectedGPU == ffmpeg.CPU {
		gpuName = "CPU (software)"
	}
	logger.Info("GPU ", strings.ToUpper(gpuName)+" detected")

	logger.Info("Plan", "Mode: "+map[string]string{"up": "Upscale", "down": "Downscale"}[mode])
	logger.Info("Plan", "Profile: "+string(profile))
	logger.Info("Plan", fmt.Sprintf("Target: %dx%d", target.W, target.H))

	enc := encoder.Auto(profile)

	// Generate output filename
	baseName := input
	extensions := []string{".mp4", ".mov", ".avi", ".mkv", ".webm", ".flv", ".wmv"}
	for _, ext := range extensions {
		if strings.HasSuffix(strings.ToLower(input), ext) {
			baseName = strings.TrimSuffix(input, ext)
			break
		}
	}

	output := fmt.Sprintf("%s-%dx%d.mp4", baseName, target.W, target.H)

	// Check if output file already exists
	if _, err := os.Stat(output); err == nil {
		logger.Info("Warning", fmt.Sprintf("Output file already exists: %s", output))
		logger.Info("Info", "It will be overwritten automatically")
	}

	logger.Info("Run ", "Encoding started...")

	argsEnc := []string{
		"-y", // Always overwrite output
		"-i", input,
		"-vf", fmt.Sprintf("scale=%d:%d:flags=lanczos", target.W, target.H),
		"-c:v", enc.Codec,
	}
	argsEnc = append(argsEnc, enc.Params...)
	argsEnc = append(argsEnc,
		"-pix_fmt", "yuv420p",
		"-c:a", "copy",
		"-movflags", "+faststart",
		"-progress", "pipe:1",
		"-nostats",
		"-loglevel", "error", // Reduce verbosity
		output,
	)

	// Log the command for debugging
	logger.Info("Debug", fmt.Sprintf("Codec: %s", enc.Codec))
	logger.Info("Debug", fmt.Sprintf("Params: %v", enc.Params))

	cmd := exec.Command("ffmpeg", argsEnc...)
	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = os.Stderr

	// Start the command
	if err := cmd.Start(); err != nil {
		logger.Info("Error", fmt.Sprintf("Failed to start FFmpeg: %v", err))

		// Try fallback to CPU if GPU encoding fails
		if detectedGPU != ffmpeg.CPU {
			logger.Info("Warning", "GPU encoding failed, trying CPU fallback...")
			ffmpeg.SetForcedGPU("cpu")
			enc = encoder.Auto(profile)

			argsEnc[6] = enc.Codec // Update codec
			argsEnc = argsEnc[:7]  // Remove old params
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

	// Monitor progress
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

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		logger.Info("Error", fmt.Sprintf("Encoding failed: %v", err))
		fmt.Println() // New line after progress
		return
	}

	fmt.Println() // New line after progress
	logger.Info("Done", fmt.Sprintf("Saved as %s", output))
	logger.Info("Info", fmt.Sprintf("Original: %dx%d → Target: %dx%d",
		res.W, res.H, target.W, target.H))
	logger.Info("Info", fmt.Sprintf("Encoder: %s", enc.Codec))

	// Reset forced GPU mode for next run
	ffmpeg.ResetForcedGPU()
}
