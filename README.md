# Video Resolution Tool

A command-line tool for automatically upscaling or downscaling video resolutions using FFmpeg. This tool intelligently detects your hardware capabilities and applies optimal encoding settings for the best quality-to-speed ratio.

## Features

- **Automatic Scaling**: Intelligently scales videos up or down based on predefined algorithms
- **Hardware Acceleration**: Automatically detects NVIDIA GPUs and uses NVENC for faster encoding
- **Quality Profiles**: Three encoding profiles (low, medium, high) for different quality/speed trade-offs
- **Real-time Progress**: Shows encoding progress with percentage completion
- **FFmpeg Integration**: Leverages FFmpeg for robust video processing

## Usage

### Basic Syntax

```bash
vr -ds "input.mp4" [profile]
vr -us "input.mp4" [profile]
```

### Parameters

- **Mode Flags**:
  - `-ds`: Downscale video (reduces resolution)
  - `-us`: Upscale video (increases resolution)

- **Input**: Path to the input MP4 video file

- **Profile** (optional, defaults to "med"):
  - `low`: Fast encoding, lower quality
  - `med`: Balanced quality and speed
  - `high`: Slow encoding, highest quality

### Examples

```bash
# Downscale a video with default medium quality
vr -ds "myvideo.mp4"

# Upscale a video with high quality settings
vr -us "myvideo.mp4" high

# Downscale with low quality for faster processing
vr -ds "myvideo.mp4" low
```

## How It Works

### Scaling Algorithm

- **Upscale**: Increases resolution by 1.5x factor
- **Downscale**: Reduces resolution by approximately 33% (2/3 factor)
- Maintains aspect ratio automatically
- Ensures even dimensions (required for most video codecs)
- Minimum width: 320px

### Encoding Profiles

#### CPU Encoding (libx264)
- **Low**: Preset "fast", CRF 23
- **Medium**: Preset "slow", CRF 16
- **High**: Preset "veryslow", CRF 14

#### NVIDIA GPU Encoding (h264_nvenc)
- **Low**: Preset p3, VBR, CQ 23
- **Medium**: Preset p5, VBR, CQ 18, tuned for quality
- **High**: Preset p7, VBR, CQ 14, multi-pass, high quality

### Output

- Output file: `input-filename-{width}x{height}.mp4`
- Video codec: H.264 (hardware accelerated if available)
- Audio: Copied from source (no re-encoding)
- Pixel format: yuv420p
- Optimized for web streaming with faststart flag

## Requirements Check

The tool automatically:
1. Verifies FFmpeg installation
2. Detects GPU capabilities
3. Probes input video resolution and duration
4. Applies appropriate scaling and encoding settings

## Error Handling

- Exits gracefully if FFmpeg is not found
- Validates input parameters
- Reports video reading errors
- Shows encoding progress and completion status

## Building from Source

If you have the Go source code:

```bash
go mod tidy
go build -o vr.exe main.go
```

## License

This project is part of kiourin-studio/video-resolution.