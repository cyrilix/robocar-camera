package camera

import (
	"bytes"
	"github.com/cyrilix/robocar-protobuf/go/events"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
	"gocv.io/x/gocv"
	"google.golang.org/protobuf/proto"
	"image/jpeg"
	"sync"
	"testing"
)

type fakeVideoSource struct {
}

func (f fakeVideoSource) Close() error {
	return nil
}

func (f fakeVideoSource) Read(dest *gocv.Mat) bool {
	img := gocv.IMRead("testdata/img.jpg", gocv.IMReadUnchanged)
	if img.Total() == 0 {
		zap.S().Info("image read is empty")
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
		publish = oldPublish
	}()
	waitEvent := sync.WaitGroup{}
	waitEvent.Add(2)
	publish = func(_ mqtt.Client, topic string, payload *[]byte) {
		muPubEvents.Lock()
		defer muPubEvents.Unlock()
		publishedEvents[topic] = payload
		waitEvent.Done()
	}

	const topic = "topic/test/camera"
	const topicRoi = "topic/test/camera-roi"
	imgBuffer := gocv.NewMat()

	part := OpencvCameraPart{
		client:           nil,
		vc:               fakeVideoSource{},
		topic:            topic,
		topicRoi:         topicRoi,
		publishFrequency: 2, // Send 2 img/s for tests
		imgBuffered:      &imgBuffer,
		horizon:          30,
	}

	go part.Start()
	waitEvent.Wait()

	for _, tpc := range []string{topic, topicRoi} {

		var frameMsg events.FrameMessage
		muPubEvents.Lock()
		err := proto.Unmarshal(*(publishedEvents[tpc]), &frameMsg)
		if err != nil {
			t.Errorf("unable to unmarshal published frame to topic %v", tpc)
		}
		muPubEvents.Unlock()

		if frameMsg.GetId().GetName() != "camera" {
			t.Errorf("bad name frame: %v, wants %v", frameMsg.GetId().GetName(), "camera")
		}
		if len(frameMsg.GetId().GetId()) != 13 {
			t.Errorf("bad id length: %v, wants %v", len(frameMsg.GetId().GetId()), 13)
		}

		if frameMsg.GetId().GetCreatedAt() == nil {
			t.Errorf("missin CreatedAt field")
		}

		_, err = jpeg.Decode(bytes.NewReader(frameMsg.GetFrame()))
		if err != nil {
			t.Errorf("image published can't be decoded: %v", err)
		}

		// Uncomment to debug image cropping
		/*
			dir, f := path.Split(fmt.Sprintf("/tmp/%s.jpg", tpc))
			 os.MkdirAll(dir, os.FileMode(0755))
			err = ioutil.WriteFile(path.Join(dir, f), frameMsg.GetFrame(), os.FileMode(0644) )
			if err != nil {
				t.Errorf("unable to write image for topic %s: %v", tpc, err)
			}
		*/
	}
}
