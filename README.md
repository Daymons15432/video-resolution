# Video Resolution Tool (vr)

A powerful command-line tool for automatically upscaling or downscaling video resolutions using FFmpeg with intelligent hardware acceleration detection and selection. This tool automatically detects available GPU encoders and applies optimal encoding settings for the best quality-to-speed ratio.

## Features

- **Multi-GPU Support**: Automatic detection and support for NVIDIA NVENC, Intel Quick Sync Video (QSV), AMD AMF/VCE, and VAAPI
- **Force GPU Selection**: Manually specify which GPU or encoder to use
- **Automatic Scaling**: Intelligently scales videos up or down based on predefined algorithms
- **Hardware Acceleration**: Leverages GPU acceleration for faster encoding when available
- **Quality Profiles**: Three encoding profiles (low, medium, high) optimized for each hardware type
- **Real-time Progress**: Shows encoding progress with percentage completion
- **Multiple Input Formats**: Supports MP4, MOV, AVI, MKV, WebM, FLV, and WMV
- **Fallback Mechanisms**: Automatically falls back to CPU encoding if GPU fails
- **FFmpeg Integration**: Uses FFmpeg for robust, professional-grade video processing
- **Compression Mode**: Reduce video file size without changing resolution
- **Flexible Argument Parsing**: Flags can be placed anywhere in the command
- **Version Information**: Check tool version with `-v` or `-version`
- **Windows Installer**: Easy Windows installation with automatic FFmpeg setup

## Usage

### Basic Syntax

```bash
vr [OPTIONS] <input-file> [profile]
vr [OPTIONS] <scale-mode> <input-file> [profile]
```

### Options

#### GPU Selection Flags
- `-cpu`: Force CPU (software) encoding
- `-nvidia`, `-nv`: Force NVIDIA GPU encoding
- `-intel`, `-qsv`: Force Intel Quick Sync Video (iGPU)
- `-amd`: Force AMD GPU encoding
- `-gpu`: Force any available GPU (auto-detect best)
- `-igpu`: Force integrated GPU (Intel/AMD)

#### Utility Flags
- `-list-gpus`: List all available GPU encoders on your system
- `-h`, `-help`: Show detailed help message

#### Scale Modes
- `-ds`: Downscale video (reduce resolution)
- `-us`: Upscale video (increase resolution)

#### Quality Profiles (optional, default: med)
- `low`: Fast encoding, lower quality
- `med`: Balanced quality and speed
- `high`: Slow encoding, highest quality

#### Compression Only (No Scaling)
- `-compress`: Compress video without changing its resolution (no upscaling or downscaling).  

### Examples

#### Basic Usage
```bash
# Auto-detect best encoder, downscale with default quality
vr -ds "video.mp4"

# Upscale with high quality
vr -us "video.mp4" high

# Compress only (no scaling), keep original resolution
vr -compress "video.mp4"

# Compress only with high quality
vr -compress "video.mp4" high
```

#### GPU-Specific Encoding
```bash
# Force CPU encoding
vr -cpu -ds "video.mp4"

# Force NVIDIA GPU
vr -nvidia -us "video.mp4" low

# Force Intel iGPU
vr -intel -ds "video.mp4" med

# Force AMD GPU
vr -amd -us "video.mp4"
```

#### Utility Commands
```bash
# List available GPU encoders
vr -list-gpus

# Show help
vr -h
```

## Supported Encoders

### NVIDIA (NVENC)
- **Low**: Preset p3, VBR, CQ 23
- **Medium**: Preset p5, VBR, CQ 18, tuned for quality
- **High**: Preset p7, VBR, CQ 14, multi-pass, high quality

### Intel Quick Sync Video (QSV)
- **Low**: Preset fast, global_quality 23
- **Medium**: Preset medium, global_quality 20, look-ahead enabled
- **High**: Preset slow, global_quality 16, extended bitrate control

### AMD AMF/VCE
- **Low**: Usage ultralowlatency, quality speed, QP 23
- **Medium**: Usage transcoding, quality balanced, QP 20
- **High**: Usage transcoding, quality quality, QP 16, preanalysis enabled

### Intel/AMD VAAPI (fallback)
- **Low**: Compression level 1, QP 23, quality speed
- **Medium**: Compression level 3, QP 20, quality balanced
- **High**: Compression level 7, QP 16, quality quality

### CPU (libx264 - fallback)
- **Low**: Preset fast, CRF 23, tuned for fast decode
- **Medium**: Preset slow, CRF 16, tuned for film
- **High**: Preset veryslow, CRF 14, 6 reference frames, 8 B-frames

## How It Works

### Scaling Algorithm

- **Upscale**: Increases resolution by 1.5x factor
- **Downscale**: Reduces resolution by approximately 33% (2/3 factor)
- Maintains aspect ratio automatically
- Ensures even dimensions (required for most video codecs)
- Minimum width: 320px

### Hardware Detection Priority

1. **NVIDIA NVENC** (highest performance)
2. **Intel Quick Sync Video** (QSV)
3. **AMD AMF/VCE**
4. **VAAPI** (Linux integrated graphics)
5. **CPU encoding** (software fallback)

### Output Specifications

- **Format**: MP4 (H.264 video)
- **Audio**: Copied from source (no re-encoding)
- **Pixel Format**: yuv420p
- **Optimization**: Faststart flag for web streaming
- **Filename**: `input-filename-{width}x{height}.mp4`

## Building from Source

If you have the Go source code:

```bash
go mod tidy
go build -o vr main.go
```

## Troubleshooting

### Common Issues

1. **"FFmpeg not found"**
   - Ensure FFmpeg is installed and in your system PATH
   - On Windows, you may need to restart your terminal after installation

2. **"GPU encoder not available"**
   - Check that your GPU drivers are up to date
   - Use `vr -list-gpus` to see available encoders
   - Try forcing CPU encoding with `-cpu` flag

3. **"Encoding failed"**
   - Check input file format and integrity
   - Try different quality profile
   - GPU encoding may fail on some systems - CPU fallback should work

4. **Poor quality output**
   - Try higher quality profile (`high`)
   - Ensure input video is not already heavily compressed
   - Check that scaling algorithm produces desired resolution


## Version History

### v1.1 (Current)
- Added compression mode (`-compress` flag)
- Flexible argument parsing (flags anywhere)
- Version information command (`-v`, `-version`)
- Windows installer with update detection
- Improved output filename generation
- Enhanced error handling and fallbacks

### v1.0
- Initial release
- Basic scaling functionality
- GPU detection and selection
- Quality profiles
- Progress tracking