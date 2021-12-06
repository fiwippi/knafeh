package ffmpeg

import (
	"math"
	"strconv"
)

type Codec int

const (
	VP8 Codec = iota
	VP9
	AV1
)

func (c Codec) Int() int {
	return int(c)
}

func (c Codec) String() string {
	return [...]string{"VP8", "VP9", "AV1"}[c]
}

func (c Codec) ArgAudioCodec() (string, string) {
	if c == VP8 {
		return "-c:a", "libvorbis"
	}
	return "-c:a", "libopus"
}

func (c Codec) ArgVideoCodec() (string, string) {
	switch c {
	case VP8:
		return "-c:v", "libvpx"
	case VP9:
		return "-c:v", "libvpx-vp9"
	case AV1:
		return "-c:v", "libaom-av1"
	}

	return "", ""
}

func (c Codec) ArgVideoCodecSpecific() [][]string {
	args := [][]string{
		{"-auto-alt-ref", "1"},
		{"-g", "128"},
	}

	if c == VP8 || c == VP9 {
		args = append(args, []string{"-lag-in-frames", "25"})
	}

	if c == VP9 {
		// –aq-mode=0 for most clean content (animation and video games). (0 means no quantisation)
		// -aq-mode=2 is recommended when you want to give more detail to the complex parts
		args = append(args, []string{"-aq-mode", "0"})
		// Improves efficiency
		args = append(args, []string{"-enable-tpl", "1"})
		// Frame parallel doesnt provide more speed so disable
		args = append(args, []string{"-frame-parallel", "0"})
	}

	if c == AV1 {
		args = append(args, []string{"-lag-in-frames", "35"})
		args = append(args, []string{"-strict", "experimental"})

		// TODO enable these av1 options
		// https://www.reddit.com/r/AV1/comments/lfheh9/encoder_tuning_part_2_making_aomencav1libaomav1/
		//args = append(args, "-aom-params enable-fwd-kf=1:enable-chroma-deltaq=1:quant-b-adapt=1")
		//args = append(args, "--enable-fwd-kf 1")
		//args = append(args, "--enable-qm=1")
		//args = append(args, "--enable-chroma-deltaq 1")
		//args = append(args, "--quant-b-adapt 1")
	}

	return args
}

func (c Codec) ArgSlices(slices, width, height, threads int) [][]string {
	// For VP8 `-slices` converts tp `--token-parts` in the libvpx encoder
	// For VP9 we just specify `-tile-columns` instead, it is more efficient than using `-row-columns`
	// With VP9, you can set `-tile-columns 6` to automatically select the largest possible amount of slices
	// Slices splits the frame up into separate partitions which can be encoded/decoded independently with threads
	// Recommended:
	//   Resolution >= 1080p, slices = 4
	//   Resolution >= 720p, slices = 3
	//   Resolution >= 480p, slices = 2
	//   Resolution < 480p, slices = 1
	switch c {
	case VP8:
		return [][]string{
			{"-slices", strconv.Itoa(slices)},
		}
	case VP9:
		return [][]string{
			{"-tile-columns", strconv.Itoa(int(math.Log2(float64(slices))))},
		}
	case AV1:
		maxCols := math.Floor(math.Log2((float64(width) + 63) / 64))
		maxRows := math.Floor(math.Log2((float64(height) + 63) / 64))

		tiles := math.Ceil(math.Log2(float64(threads)) / 2)
		// Ensure minimum value is 1, this is needed because
		// width and height could be unspecified
		cols := int(math.Max(math.Min(tiles, maxCols), 1))
		rows := int(math.Max(math.Min(tiles, maxRows), 1))

		return [][]string{
			{"-tile-columns", strconv.Itoa(cols)},
			{"-tile-rows", strconv.Itoa(rows)},
		}

	}

	return nil
}

func (c Codec) ArgRowMT(mt bool) [][]string {
	// Row multithreading speeds up the encode with negligible quality loss
	// -cpu-used is the speed of the encode
	if mt {
		switch c {
		case VP8:
			return [][]string{
				{"-cpu-used", "1"},
			}
		case VP9:
			return [][]string{
				{"-cpu-used", "1"},
				{"-row-mt", "1"},
			}
		case AV1:
			return [][]string{
				{"-cpu-used", "4"},
				{"-row-mt", "1"},
			}
		}
	} else {
		switch c {
		case VP8:
			return [][]string{
				{"-cpu-used", "0"},
			}
		case VP9:
			return [][]string{
				{"-cpu-used", "0"},
				{"-row-mt", "0"},
			}
		case AV1:
			return [][]string{
				{"-cpu-used", "1"},
				{"-row-mt", "0"},
			}
		}
	}

	return nil
}
