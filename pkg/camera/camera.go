package camera

import (
	"fmt"
	"github.com/cyrilix/robocar-protobuf/go/events"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	"go.uber.org/zap"
	"gocv.io/x/gocv"
	"image"
	"io"
	"sync"
	"time"
)

type VideoSource interface {
	Read(m *gocv.Mat) bool
	io.Closer
}

type OpencvCameraPart struct {
	client mqtt.Client
	vc     VideoSource
	topic  string
	topicRoi string
	publishFrequency int
	muImgBuffered    sync.Mutex
	imgBuffered      *gocv.Mat
	horizon			 int
	cancel           chan interface{}
}

func New(client mqtt.Client, topic string, topicRoi string, publishFrequency int,
	videoProperties map[gocv.VideoCaptureProperties]float64, horizon int) *OpencvCameraPart {
	zap.S().Info("run camera part")

	vc, err := gocv.OpenVideoCapture(0)
	if err != nil {
		zap.S().Fatalf("unable to open video device: %v", err)
	}
	for k, v := range videoProperties {
		vc.Set(k, v)
	}

	img := gocv.NewMat()
	o := OpencvCameraPart{
		client:           client,
		vc:               vc,
		topic:            topic,
		topicRoi: 		  topicRoi,
		publishFrequency: publishFrequency,
		imgBuffered:      &img,
	}
	return &o
}

func (o *OpencvCameraPart) Start() error {
	zap.S().Info("start camera")
	o.cancel = make(chan interface{})
	ticker := time.NewTicker(1 * time.Second / time.Duration(o.publishFrequency))
	defer ticker.Stop()

	for {
		select {

		case tickerTime := <-ticker.C:
			o.publishFrames(tickerTime)
		case <-o.cancel:
			return nil
		}
	}
}

func (o *OpencvCameraPart) Stop() {
	zap.S().Info("close video device")
	close(o.cancel)

	if err := o.vc.Close(); err != nil {
		zap.S().Errorf("unexpected error while VideoCapture is closed: %v", err)
	}
	if err := o.imgBuffered.Close(); err != nil {
		zap.S().Errorf("unexpected error while VideoCapture is closed: %v", err)
	}
}

func (o *OpencvCameraPart) publishFrames(tickerTime time.Time) {
	o.muImgBuffered.Lock()
	defer o.muImgBuffered.Unlock()

	o.vc.Read(o.imgBuffered)

	// Publish raw image
	o.publishFrame(tickerTime, o.topic, o.imgBuffered)

	if o.horizon == 0 {
		return
	}

	// Region of interest
	roi := o.imgBuffered.Region(image.Rect(0, o.horizon, o.imgBuffered.Cols(), o.imgBuffered.Rows()))
	defer roi.Close()
	o.publishFrame(tickerTime, o.topicRoi, &roi)
}

func (o *OpencvCameraPart) publishFrame(tickerTime time.Time, topic string, frame *gocv.Mat) {
	img, err := gocv.IMEncode(gocv.JPEGFileExt, *frame)
	if err != nil {
		zap.S().With("topic", topic).Errorf("unable to convert image to jpeg: %v", err)
		return
	}

	msg := &events.FrameMessage{
		Id: &events.FrameRef{
			Name: "camera",
			Id:   fmt.Sprintf("%d%03d", tickerTime.Unix(), tickerTime.Nanosecond()/1000/1000),
			CreatedAt: &timestamp.Timestamp{
				Seconds: tickerTime.Unix(),
				Nanos:   int32(tickerTime.Nanosecond()),
			},
		},
		Frame: img,
	}

	payload, err := proto.Marshal(msg)
	if err != nil {
		zap.S().Errorf("unable to marshal protobuf message: %v", err)
	}

	publish(o.client, topic, &payload)
}


var publish = func(client mqtt.Client, topic string, payload *[]byte) {
	token := client.Publish(topic, 0, false, *payload)
	token.WaitTimeout(10 * time.Millisecond)
	if err := token.Error(); err != nil {
		zap.S().Errorf("unable to publish frame: %v", err)
	}
}
