package camera

import (
	"github.com/cyrilix/robocar-base/mqttdevice"
	"gocv.io/x/gocv"
	"io"
	"log"
	"sync"
	"time"
)

type VideoSource interface {
	Read(m *gocv.Mat) bool
	io.Closer
}

type OpencvCameraPart struct {
	vc               VideoSource
	pub              mqttdevice.Publisher
	topic            string
	publishFrequency int
	muImgBuffered    sync.Mutex
	imgBuffered      *gocv.Mat
}

func New(topic string, publisher mqttdevice.Publisher, publishFrequency int, videoProperties map[gocv.VideoCaptureProperties]float64) *OpencvCameraPart {
	log.Printf("Run camera part")

	vc, err := gocv.OpenVideoCapture(0)
	if err != nil {
		log.Fatalf("unable to open video device: %v", err)
	}
	for k, v := range videoProperties {
		vc.Set(k, v)
	}

	img := gocv.NewMat()
	o := OpencvCameraPart{
		vc:               vc,
		pub:              publisher,
		topic:            topic,
		publishFrequency: publishFrequency,
		imgBuffered:      &img,
	}
	return &o
}

func (o *OpencvCameraPart) Start() error {
	log.Printf("start camera")
	ticker := time.NewTicker(1 * time.Second / time.Duration(o.publishFrequency))
	defer ticker.Stop()

	for {
		go o.publishFrame()
		<-ticker.C
	}
}

func (o *OpencvCameraPart) Stop() {
	log.Print("close video device")
	if err := o.vc.Close(); err != nil {
		log.Printf("unexpected error while VideoCapture is closed: %v", err)
	}
	if err := o.imgBuffered.Close(); err != nil {
		log.Printf("unexpected error while VideoCapture is closed: %v", err)
	}
}

func (o *OpencvCameraPart) publishFrame() {
	o.muImgBuffered.Lock()
	defer o.muImgBuffered.Unlock()

	o.vc.Read(o.imgBuffered)
	img, err := gocv.IMEncode(gocv.JPEGFileExt, *o.imgBuffered)
	if err != nil {
		log.Printf("unable to convert image to jpeg: %v", err)
		return
	}

	o.pub.Publish(o.topic, img)
}
