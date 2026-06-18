package main

import (
	"os"
	"path/filepath"
	"strings"
)

var audioExts = map[string]bool{
	".mp3":  true,
	".ogg":  true,
	".flac": true,
	".wav":  true,
	".aac":  true,
	".m4a":  true,
}

// scanDir recursively walks dir and returns an AudioFile for every recognised
// audio file. Metadata (bitrate, duration, tags) is NOT populated here; that
// happens lazily in probeAndShowDetails when a file is selected.
func scanDir(dir string) ([]*AudioFile, error) {
	var files []*AudioFile
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if !audioExts[ext] {
			return nil
		}

		relPath, _ := filepath.Rel(dir, path)
		files = append(files, &AudioFile{
			Path:    path,
			RelPath: relPath,
			Name:    info.Name(),
			Size:    info.Size(),
			Format:  strings.ToUpper(strings.TrimPrefix(ext, ".")), // e.g. "MP3"
		})
		return nil
	})
	return files, err
}
