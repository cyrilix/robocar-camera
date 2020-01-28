package camera

import (
	"fmt"
	"github.com/cyrilix/robocar-protobuf/go/events"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
	"io"
	"sync"
	"time"
)

type VideoSource interface {
	Read(m *gocv.Mat) bool
	io.Closer
}

type OpencvCameraPart struct {
	client           mqtt.Client
	vc               VideoSource
	topic            string
	publishFrequency int
	muImgBuffered    sync.Mutex
	imgBuffered      *gocv.Mat
	cancel           chan interface{}
}

func New(client mqtt.Client, topic string, publishFrequency int, videoProperties map[gocv.VideoCaptureProperties]float64) *OpencvCameraPart {
	log.Infof("run camera part")

	vc, err := gocv.OpenVideoCapture(0)
	if err != nil {
		log.Fatalf("unable to open video device: %v", err)
	}
	for k, v := range videoProperties {
		vc.Set(k, v)
	}

	img := gocv.NewMat()
	o := OpencvCameraPart{
		client:           client,
		vc:               vc,
		topic:            topic,
		publishFrequency: publishFrequency,
		imgBuffered:      &img,
	}
	return &o
}

func (o *OpencvCameraPart) Start() error {
	log.Printf("start camera")
	o.cancel = make(chan interface{})
	ticker := time.NewTicker(1 * time.Second / time.Duration(o.publishFrequency))
	defer ticker.Stop()

	for {
		select {

		case tickerTime := <-ticker.C:
			o.publishFrame(tickerTime)
		case <-o.cancel:
			return nil
		}
	}
}

func (o *OpencvCameraPart) Stop() {
	log.Print("close video device")
	close(o.cancel)

	if err := o.vc.Close(); err != nil {
		log.Printf("unexpected error while VideoCapture is closed: %v", err)
	}
	if err := o.imgBuffered.Close(); err != nil {
		log.Printf("unexpected error while VideoCapture is closed: %v", err)
	}
}

func (o *OpencvCameraPart) publishFrame(tickerTime time.Time) {
	o.muImgBuffered.Lock()
	defer o.muImgBuffered.Unlock()

	o.vc.Read(o.imgBuffered)

	img, err := gocv.IMEncode(gocv.JPEGFileExt, *o.imgBuffered)
	if err != nil {
		log.Printf("unable to convert image to jpeg: %v", err)
		return
	}

	msg := &events.FrameMessage{
		Id: &events.FrameRef{
			Name: "camera",
			Id:   fmt.Sprintf("%d%000d", tickerTime.Unix(), tickerTime.Nanosecond()/1000/1000),
			CreatedAt: &timestamp.Timestamp{
				Seconds: tickerTime.Unix(),
				Nanos:   int32(tickerTime.Nanosecond()),
			},
		},
		Frame: img,
	}

	payload, err := proto.Marshal(msg)
	if err != nil {
		log.Errorf("unable to marshal protobuf message: %v", err)
	}

	publish(o.client, o.topic, &payload)
}

var publish = func(client mqtt.Client, topic string, payload *[]byte) {
	token := client.Publish(topic, 0, false, *payload)
	token.WaitTimeout(10 * time.Millisecond)
	if err := token.Error(); err != nil {
		log.Errorf("unable to publish frame: %v", err)
	}
}
