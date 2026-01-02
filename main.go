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

func main() {
	logger.Info("Init", "Preparing engine...")
	time.Sleep(300 * time.Millisecond)

	if err := ffmpeg.Init(); err != nil {
		logger.Info("Error", "Failed to find FFmpeg in PATH. Please install FFmpeg first.")
		return
	}
	logger.Info("Init", "FFmpeg ready")

	if len(os.Args) < 3 {
		fmt.Println("vr -ds video.mp4 [low|med|high]")
		fmt.Println("vr -us video.mp4 [low|med|high]")
		return
	}

	mode := map[string]string{"-ds": "down", "-us": "up"}[os.Args[1]]
	if mode == "" {
		logger.Info("Error", "Invalid flag")
		return
	}

	input := os.Args[2]
	profile := encoder.Med
	if len(os.Args) >= 4 {
		profile = encoder.ParseProfile(os.Args[3])
	}

	logger.Info("Scan", "Reading video info...")
	res, err := probe.ResolutionOf(input)
	if err != nil {
		logger.Info("Error", "Cannot read video")
		return
	}
	dur, _ := probe.Duration(input)

	logger.Info("Scan", fmt.Sprintf("Resolution: %dx%d", res.W, res.H))
	logger.Info("Scan", fmt.Sprintf("Duration: %.0f sec", dur))

	target := scaler.Auto(
		scaler.Resolution{W: res.W, H: res.H},
		mode,
	)

	gpu := ffmpeg.DetectGPU()
	logger.Info("GPU ", strings.ToUpper(string(gpu))+" detected")

	logger.Info("Plan", "Mode: "+map[string]string{"up": "Upscale", "down": "Downscale"}[mode])
	logger.Info("Plan", "Profile: "+string(profile))
	logger.Info("Plan", fmt.Sprintf("Target: %dx%d", target.W, target.H))

	enc := encoder.Auto(profile)

	output := strings.TrimSuffix(input, ".mp4") +
		fmt.Sprintf("-%dx%d.mp4", target.W, target.H)

	logger.Info("Run ", "Encoding started...")

	args := []string{
		"-y",
		"-i", input,
		"-vf", fmt.Sprintf("scale=%d:%d:flags=lanczos", target.W, target.H),
		"-c:v", enc.Codec,
	}
	args = append(args, enc.Params...)
	args = append(args,
		"-pix_fmt", "yuv420p",
		"-c:a", "copy",
		"-movflags", "+faststart",
		"-progress", "pipe:1",
		"-nostats",
		output,
	)

	cmd := exec.Command("ffmpeg", args...)
	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = os.Stderr
	cmd.Start()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "out_time_ms=") {
			ms, _ := strconv.ParseFloat(strings.TrimPrefix(scanner.Text(), "out_time_ms="), 64)
			p := (ms / 1_000_000) / dur * 100
			if p > 100 {
				p = 100
			}
			logger.Inline(fmt.Sprintf("Progress: %.1f%%", p))
		}
	}

	cmd.Wait()
	fmt.Println()
	logger.Info("Done", "Saved as "+output)
}
