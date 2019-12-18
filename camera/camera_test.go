package camera

import (
	"bytes"
	"github.com/cyrilix/robocar-base/testtools"
	"gocv.io/x/gocv"
	"image/jpeg"
	"io"
	"log"
	"testing"
	"time"
)

type fakeVideoSource struct {
	io.Closer
}

func (f fakeVideoSource) Read(dest *gocv.Mat) bool {
	img := gocv.IMRead("testdata/img.jpg", gocv.IMReadUnchanged)
	if img.Total() == 0 {
		log.Print("image read is empty")
		return false
	}
	img.CopyTo(dest)
	return true
}

func TestOpencvCameraPart(t *testing.T) {
	p := testtools.NewFakePublisher()
	const topic = "topic/test/camera"
	imgBuffer := gocv.NewMat()

	part := OpencvCameraPart{
		vc:               fakeVideoSource{},
		pub:              p,
		topic:            topic,
		publishFrequency: 1000,
		imgBuffered:      &imgBuffer,
	}

	go part.Start()
	time.Sleep(1 * time.Millisecond)

	img := p.PublishedEvent(topic)
	if img == nil {
		t.Fatalf("event %s has not been published", topic)
	}
	content, err := img.ByteSliceValue()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = jpeg.Decode(bytes.NewReader(content))
	if err != nil {
		t.Errorf("image published can't be decoded: %v", err)
	}
}
