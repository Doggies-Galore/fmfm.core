package main

import (
	"strings"
	"io/ioutil"
	"time"

	"github.com/but80/fmfm/cmd/fmfm-cli/internal/player"
	"github.com/but80/fmfm/ymf"
	"github.com/but80/smaf825/pb/smaf"
	"github.com/golang/protobuf/proto"
)

func main() {
	info, err := ioutil.ReadDir("voice")
	if err != nil {
		panic(err)
	}
	libs := []*smaf.VM5VoiceLib{}
	for _, i := range info {
		if i.IsDir() || !strings.HasSuffix(i.Name(), ".vm5.pb") {
			continue
		}
		b, err := ioutil.ReadFile("voice/" + i.Name())
		if err != nil {
			panic(err)
		}
		var lib smaf.VM5VoiceLib
		err = proto.Unmarshal(b, &lib)
		if err != nil {
			panic(err)
		}
		libs = append(libs, &lib)
	}

	renderer := player.NewRenderer()
	chip := ymf.NewChip(renderer.Parameters.SampleRate, -12.0)
	seq := player.NewSequencer(chip, libs)
	seq.Reset()
	renderer.Start(chip.Next)
	time.Sleep(24 * time.Hour)
}
