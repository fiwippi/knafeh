package ffmpeg

import (
	"strconv"
	"strings"
)

func (i *Inputs) ParseCodec(c string) error {
	switch strings.ToLower(c) {
	case "vp8":
		i.Codec = VP8
		return nil
	case "vp9":
		i.Codec = VP9
		return nil
	case "av1":
		i.Codec = AV1
		return nil
	}

	return ErrInvalidCodec
}

func (i *Inputs) ParseCRF(crf int) error {
	if crf < 0 || crf > 63 {
		return ErrInvalidCRF
	}

	i.VarArgs.CRF = crf
	return nil
}

func abs(a int) int {
	if a < 0 {
		a *= -1
	}
	return a
}

func (i *Inputs) ParseAudioBitrate(b int) error {
	if b < 6 {
		return ErrAudioBitrate
	}

	// For VP9/AV1
	i.VarArgs.AudioBitrate = b

	// For VP8
	qscale := []int{45, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 500}
	min := 500
	for j, v := range qscale {
		diff := abs(b - v)
		if diff < min {
			min = diff
			i.VarArgs.AudioQualityScale = j - 1
		}
	}

	return nil
}

func (i *Inputs) ParseResize(s string) error {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return ErrResize
	}

	w, err := strconv.Atoi(parts[0])
	if err != nil {
		return ErrResize
	}

	h, err := strconv.Atoi(parts[1])
	if err != nil {
		return ErrResize
	}

	i.Resize.Width = w
	i.Resize.Height = h

	return nil
}

func (i *Inputs) ParseCrop(s string) error {
	parts := strings.Split(s, ":")
	if len(parts) != 4 {
		return ErrCrop
	}

	x, err := strconv.Atoi(parts[0])
	if err != nil {
		return ErrCrop
	}
	y, err := strconv.Atoi(parts[1])
	if err != nil {
		return ErrCrop
	}
	w, err := strconv.Atoi(parts[2])
	if err != nil {
		return ErrCrop
	}
	h, err := strconv.Atoi(parts[3])
	if err != nil {
		return ErrCrop
	}

	i.Crop.X = x
	i.Crop.Y = y
	i.Crop.W = w
	i.Crop.H = h

	return nil
}