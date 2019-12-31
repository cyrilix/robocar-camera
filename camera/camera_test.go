package camera

import (
	"bytes"
	"github.com/cyrilix/robocar-protobuf/go/events"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/golang/protobuf/proto"
	"gocv.io/x/gocv"
	"image/jpeg"
	"log"
	"sync"
	"testing"
	"time"
)

type fakeVideoSource struct {
}

func (f fakeVideoSource) Close() error {
	return nil
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
	var muPubEvents sync.Mutex
	publishedEvents := make(map[string]*[]byte)
	oldPublish := publish
	defer func() {
		publish = oldPublish}()
	publish = func(_ mqtt.Client, topic string, payload *[]byte){
		muPubEvents.Lock()
		defer muPubEvents.Unlock()
		publishedEvents[topic] = payload
	}

	const topic = "topic/test/camera"
	imgBuffer := gocv.NewMat()

	part := OpencvCameraPart{
		client: nil,
		vc:               fakeVideoSource{},
		topic:            topic,
		publishFrequency: 1000,
		imgBuffered:      &imgBuffer,
	}


	go part.Start()
	time.Sleep(5 * time.Millisecond)

	var frameMsg events.FrameMessage
	muPubEvents.Lock()
	err := proto.Unmarshal(*(publishedEvents[topic]), &frameMsg)
	if err != nil {
		t.Errorf("unable to unmarshal pubblished frame")
	}
	muPubEvents.Unlock()

	if frameMsg.GetId().GetName() != "camera" {
		t.Errorf("bad name frame: %v, wants %v", frameMsg.GetId().GetName(), "camera")
	}
	if frameMsg.GetId().GetId() != "XX" {
		t.Errorf("bad name frame: %v, wants %v", frameMsg.GetId().GetId(), "XX")
	}

	_, err = jpeg.Decode(bytes.NewReader(frameMsg.GetFrame()))
	if err != nil {
		t.Errorf("image published can't be decoded: %v", err)
	}
}
