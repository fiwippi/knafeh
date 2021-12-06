package ffmpeg

import (
	"context"
	"os"
	"time"

	"gopkg.in/vansante/go-ffprobe.v2"
)

// FileData holds the data knafeh needs from ffprobe
type FileData struct {
	Title           string
	Width, Height   int
	DurationSeconds float64
	VideoStreams    []*ffprobe.Stream
	AudioStreams    []*ffprobe.Stream
	SubtitleStreams []*ffprobe.Stream
}

func (fd *FileData) ValidDimensions() bool {
	return fd.Width > 0 && fd.Height > 0
}

func (fd *FileData) ValidDuration() bool {
	return fd.DurationSeconds > 0
}

func (fd *FileData) HasTitle() bool {
	return fd.Title != ""
}

// Probe returns FileData output for a file on the filesystem
func Probe(fp string) (*FileData, error) {
	fileReader, err := os.Open(fp)
	if err != nil {
		return nil, err
	}
	defer fileReader.Close()

	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	data, err := ffprobe.ProbeReader(ctx, fileReader)
	if err != nil {
		return nil, err
	}

	return probeDataToFileData(data), nil
}

// probeDataToFileData Converts ffprobe.ProbeData to FileData
func probeDataToFileData(data *ffprobe.ProbeData) *FileData {
	var fd = &FileData{
		Width:           -1,
		Height:          -1,
		DurationSeconds: -1,
	}

	// Get the duration and title
	if data.Format != nil {
		fd.DurationSeconds = data.Format.DurationSeconds

		if data.Format.Tags != nil {
			fd.Title = data.Format.Tags.Title
		}

	}

	// Set width and height
	if data.FirstVideoStream() != nil {
		fd.Width = data.FirstVideoStream().Width
		fd.Height = data.FirstVideoStream().Height
	}

	// Retrieve the streams from the probe data
	var vI, aI, sI int
	for _, s := range data.Streams {
		switch s.CodecType {
		case "video":
			fd.VideoStreams = append(fd.VideoStreams, s)
			fd.VideoStreams[vI].Index = vI
			vI += 1
		case "audio":
			fd.AudioStreams = append(fd.AudioStreams, s)
			fd.AudioStreams[aI].Index = aI
			aI += 1
		case "subtitle":
			fd.SubtitleStreams = append(fd.SubtitleStreams, s)
			fd.SubtitleStreams[sI].Index = sI
			sI += 1
		}
	}

	return fd
}
