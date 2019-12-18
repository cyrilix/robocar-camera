package main

import (
	"flag"
	"github.com/cyrilix/robocar-base/cli"
	"github.com/cyrilix/robocar-base/mqttdevice"
	"github.com/cyrilix/robocar-camera/camera"
	"gocv.io/x/gocv"
	"log"
	"os"
)

const DefaultClientId = "robocar-camera"

func main() {
	var mqttBroker, username, password, clientId, topicBase string
	var pubFrequency int
	var device, videoWidth, videoHeight int

	mqttQos := cli.InitIntFlag("MQTT_QOS", 0)
	_, mqttRetain := os.LookupEnv("MQTT_RETAIN")

	cli.InitMqttFlags(DefaultClientId, &mqttBroker, &username, &password, &clientId, &mqttQos, &mqttRetain)

	flag.StringVar(&topicBase, "mqtt-topic", os.Getenv("MQTT_TOPIC"), "Mqtt topic to publish camera frames, use MQTT_TOPIC if args not set")
	flag.IntVar(&pubFrequency, "mqtt-pub-frequency", 25., "Number of messages to publish per second")

	flag.IntVar(&device, "video-device", 0, "Video device number")
	flag.IntVar(&videoWidth, "video-width", 160, "Video pixels width")
	flag.IntVar(&videoHeight, "video-height", 128, "Video pixels height")

	flag.Parse()
	if len(os.Args) <= 1 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	pubSub := mqttdevice.NewPahoMqttPubSub(mqttBroker, username, password, clientId, mqttQos, mqttRetain)
	defer func() {
		err := pubSub.Close()
		if err != nil {
			log.Printf("unable to close mqtt publisher: %v", err)
		}
	}()

	videoProperties := make(map[gocv.VideoCaptureProperties]float64)
	videoProperties[gocv.VideoCaptureFrameWidth] = float64(videoWidth)
	videoProperties[gocv.VideoCaptureFrameHeight] = float64(videoHeight)

	c := camera.New(topicBase, pubSub, pubFrequency, videoProperties)
	defer c.Stop()

	cli.HandleExit(c)

	err := c.Start()
	if err != nil {
		log.Fatalf("unable to start service: %v", err)
	}
}
