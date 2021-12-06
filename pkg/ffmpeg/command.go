package ffmpeg

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"

	orderedmap "github.com/wk8/go-ordered-map"
)

type Command struct {
	// Order or args joined together is as specified by
	// the numbers
	inputArgs          *orderedmap.OrderedMap // #1
	videoCodecArgs     *orderedmap.OrderedMap // #2
	videoFilterArgs    *orderedmap.OrderedMap // #3
	audioCodecArgs     *orderedmap.OrderedMap // #4
	dubAudioFilterArgs *orderedmap.OrderedMap // #5
	mapArgs            *orderedmap.OrderedMap // #6
	generalArgs        *orderedmap.OrderedMap // #7

	audioFilterInput int // Input the audio filter should affect

	// fp of the input and output
	twoPass bool
	inputFp string
	outputFp string
}

// Private
func newCommand() *Command {
	return &Command{
		generalArgs:        orderedmap.New(),
		inputArgs:          orderedmap.New(),
		videoCodecArgs:     orderedmap.New(),
		audioCodecArgs:     orderedmap.New(),
		videoFilterArgs:    orderedmap.New(),
		dubAudioFilterArgs: orderedmap.New(),
		mapArgs:            orderedmap.New(),
	}
}

func (c *Command) addInputArgs(arg string) {
	c.inputArgs.Set(fmt.Sprintf("\"%s\"", arg), true)
}

func (c *Command) addMapArgs(arg string) {
	c.mapArgs.Set(arg, true)
}

func (c *Command) addGeneralArg(k, v string) {
	c.generalArgs.Set(k, v)
}

func (c *Command) addVideoArg(k, v string) {
	c.videoCodecArgs.Set(k, v)
}

func (c *Command) addVideoArgs(args [][]string) {
	for _, pair := range args {
		c.videoCodecArgs.Set(pair[0], pair[1])
	}

}

func (c *Command) addVideoFilterArg(k, v string) {
	c.videoFilterArgs.Set(k, v)
}

func (c *Command) addAudioFilterArg(k, v string) {
	c.dubAudioFilterArgs.Set(k, v)
}

func (c *Command) addAudioArg(k, v string) {
	c.audioCodecArgs.Set(k, v)
}

func (c *Command) VideoFiltersString() string {
	var str string

	// #3
	if c.videoFilterArgs.Len() > 0 {
		var filters string
		for pair := c.videoFilterArgs.Oldest(); pair != nil; pair = pair.Next() {
			filters += fmt.Sprintf(",%s=%s", pair.Key, pair.Value)
		}
		str = fmt.Sprintf("[0:v]%s", strings.TrimPrefix(filters, ","))
	}

	return str
}

func (c *Command) AudioFiltersString() string {
	var str string

	// #5
	if c.dubAudioFilterArgs.Len() > 0 {
		var filters string
		for pair := c.dubAudioFilterArgs.Oldest(); pair != nil; pair = pair.Next() {
			filters += fmt.Sprintf(",%s=%s", pair.Key, pair.Value)
		}
		str = fmt.Sprintf("[%d:a]%s", c.audioFilterInput, strings.TrimPrefix(filters, ","))
	}

	return str
}

func (c *Command) StringSlice() []string {
	str := make([]string, 0)

	// #1
	for pair := c.inputArgs.Oldest(); pair != nil; pair = pair.Next() {
		str = append(str, "-i")
		str = append(str, fmt.Sprintf("%s", pair.Key))
	}

	// #2
	for pair := c.videoCodecArgs.Oldest(); pair != nil; pair = pair.Next() {
		str = append(str, fmt.Sprintf("%s", pair.Key))
		str = append(str, fmt.Sprintf("%s", pair.Value))
	}

	// #3
	if c.videoFilterArgs.Len() > 0 {
		str = append(str, "-filter_complex")
		str = append(str, fmt.Sprintf("%s", c.VideoFiltersString()))
	}

	// #4
	for pair := c.audioCodecArgs.Oldest(); pair != nil; pair = pair.Next() {
		str = append(str, fmt.Sprintf("%s", pair.Key))
		str = append(str, fmt.Sprintf("%s", pair.Value))
	}

	// #5
	if c.dubAudioFilterArgs.Len() > 0 {
		fmt.Println(c.dubAudioFilterArgs.Oldest())
		str = append(str, "-filter_complex")
		str = append(str, fmt.Sprintf("%s", c.AudioFiltersString()))
	}

	// #6
	for pair := c.mapArgs.Oldest(); pair != nil; pair = pair.Next() {
		str = append(str, "-map")
		str = append(str, fmt.Sprintf("%s", pair.Key))
	}

	// #7
	for pair := c.generalArgs.Oldest(); pair != nil; pair = pair.Next() {
		if pair.Value == "" {
			str = append(str, fmt.Sprintf("%s", pair.Key))
		} else {
			str = append(str, fmt.Sprintf("%s", pair.Key))
			str = append(str, fmt.Sprintf("%s", pair.Value))
		}

	}

	return str
}

func (c *Command) String() string {
	return strings.Join(c.StringSlice(), " ")
}

func (c *Command) firstPassArgs(passlogfp string) []string {
	// Setup first pass
	args := make([]string, 0)
	core := c.StringSlice()

	if c.twoPass {
		// Looping filters cause infinite loops on the first pass so we remove them and
		// replace them with null equivalents which simply pass through the stream
		dubbing := false
		for i, v := range core {
			if v == "loop=-1:32767:0" {
				core[i] = "null"
			}
			if v == "aloop=-1:2147483647:0" {
				core[i] = "anull"
			}
			if strings.Contains(v, ":a") {
				dubbing = true
			}
		}

		args = append(args, "-i")
		args = append(args, c.inputFp)
		if !dubbing { // If we're not dubbing audio
			args = append(args, "-an")
		}

		args = append(args, core...)
		args = append(args, "-y")
		args = append(args, "-pass")
		args = append(args, "1")
		args = append(args, "-passlogfile")
		args = append(args, passlogfp)
		if runtime.GOOS == "windows" {
			args = append(args, "NUL")
		} else {
			args = append(args, "/dev/null")
		}
	} else {
		args = append(args, "-i")
		args = append(args, c.inputFp)
		args = append(args, core...)
		args = append(args, "-y")
		args = append(args, c.outputFp)
	}

	return args
}

func (c *Command) secondPassArgs(passlogfp string) []string {
	// Setup first pass
	args := make([]string, 0)
	core := c.StringSlice()

	args = append(args, "-i")
	args = append(args, c.inputFp)
	args = append(args, core...)
	args = append(args, "-y")
	args = append(args, "-pass")
	args = append(args, "2")
	args = append(args, "-passlogfile")
	args = append(args, passlogfp)
	args = append(args, c.outputFp)

	return args
}

func (c *Command) Run() error {
	var passlogfp string
	var p1, p2 *exec.Cmd

	// Get the passlogfp if needed
	if c.twoPass {
		file, err := ioutil.TempFile("", "knafeh")
		if err != nil {
			return err
		}
		defer func() {
			file.Close()
			os.Remove(file.Name())
		}()
		passlogfp = file.Name()
	}

	// Create processes for the first and second pass and run them
	p1 = exec.Command("ffmpeg", c.firstPassArgs(passlogfp)...)
	p1.Stdout = os.Stdout
	p1.Stderr = os.Stderr

	// Run the commands
	fmt.Println("------------STARTING------------")
	fmt.Println("---------RUNNING-PASS-1---------")
	fmt.Println(p1)
	err := p1.Run()
	if err != nil {
		return err
	}
	fmt.Println("-----------PASS-1-DONE----------")
	if c.twoPass {
		p2 = exec.Command("ffmpeg", c.secondPassArgs(passlogfp)...)
		p2.Stdout = os.Stdout
		p2.Stderr = os.Stderr

		fmt.Println("---------RUNNING-PASS-2---------")
		fmt.Println(p2)
		err := p2.Run()
		if err != nil {
			return err
		}
		fmt.Println("-----------PASS-2-DONE----------")
	}
	fmt.Println("--------------DONE--------------")

	return nil
}