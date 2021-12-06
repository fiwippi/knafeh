package ffmpeg

// RecommendedSlices calculates the recommended number
// of slices for a video given its width and height
func RecommendedSlices(width, height int) int {
	// 1080p (1920*1080) - 4 slices
	// 720p (1280*720) - 3 slices
	// 480p (640*480) - 2 slices

	s := 1
	if width*height >= 1920*1080 {
		s = 4
	} else if width*height >= 1280*720 {
		s = 3
	} else if width*height >= 640*480 {
		s = 2
	}

	return s
}
