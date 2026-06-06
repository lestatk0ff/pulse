package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// probeFile runs ffprobe on a single file and returns normalized media metadata.
// This stays separate from scanDir so directory walking logic does not carry probe concerns.
func probeFile(path string) (format string, bitrate int, duration string, tags map[string]string) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams", "-show_format",
		path,
	)
	out, err := cmd.Output()
	if err != nil {
		return "", 0, "", nil
	}

	var probe ffprobeOutput
	if err := json.Unmarshal(out, &probe); err != nil {
		return "", 0, "", nil
	}

	// Prefer container-level bitrate and use stream bitrate only as a fallback.
	if probe.Format.BitRate != "" {
		if br, err := strconv.ParseInt(probe.Format.BitRate, 10, 64); err == nil {
			bitrate = int(br / 1000)
		}
	}

	if probe.Format.Duration != "" {
		if secs, err := strconv.ParseFloat(probe.Format.Duration, 64); err == nil {
			s := int(secs)
			duration = fmt.Sprintf("%d:%02d", s/60, s%60)
		}
	}

	tags = make(map[string]string)
	for k, v := range probe.Format.Tags {
		tags[strings.ToLower(k)] = v
	}

	for _, s := range probe.Streams {
		if s.CodecType == "audio" {
			format = strings.ToUpper(s.CodecName)
			if bitrate == 0 && s.BitRate != "" {
				if br, err := strconv.ParseInt(s.BitRate, 10, 64); err == nil {
					bitrate = int(br / 1000)
				}
			}
		}
	}
	return format, bitrate, duration, tags
}
