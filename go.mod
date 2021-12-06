module github.com/fiwippi/knafeh

go 1.16

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jessevdk/go-flags v1.5.0
	github.com/wk8/go-ordered-map v0.2.0
	gopkg.in/vansante/go-ffprobe.v2 v2.0.2
)

replace gopkg.in/vansante/go-ffprobe.v2 v2.0.2 => ./pkg/forked/ffprobe
