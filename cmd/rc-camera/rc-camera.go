package main

import (
	"flag"
	"github.com/cyrilix/robocar-base/cli"
	"github.com/cyrilix/robocar-camera/camera"
	"go.uber.org/zap"
	"gocv.io/x/gocv"
	"log"
	"os"
)

const DefaultClientId = "robocar-camera"

func main() {
	var mqttBroker, username, password, clientId, topicBase string
	var pubFrequency int
	var device, videoWidth, videoHeight int
	var debug bool

	mqttQos := cli.InitIntFlag("MQTT_QOS", 0)
	_, mqttRetain := os.LookupEnv("MQTT_RETAIN")

	cli.InitMqttFlags(DefaultClientId, &mqttBroker, &username, &password, &clientId, &mqttQos, &mqttRetain)

	flag.StringVar(&topicBase, "mqtt-topic", os.Getenv("MQTT_TOPIC"), "Mqtt topic to publish camera frames, use MQTT_TOPIC if args not set")
	flag.IntVar(&pubFrequency, "mqtt-pub-frequency", 25., "Number of messages to publish per second")

	flag.IntVar(&device, "video-device", 0, "Video device number")
	flag.IntVar(&videoWidth, "video-width", 160, "Video pixels width")
	flag.IntVar(&videoHeight, "video-height", 128, "Video pixels height")

	flag.BoolVar(&debug, "debug", false, "Display raw value to debug")

	flag.Parse()
	if len(os.Args) <= 1 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	config := zap.NewDevelopmentConfig()
	if debug {
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	lgr, err := config.Build()
	if err != nil {
		log.Fatalf("unable to init logger: %v", err)
	}
	defer func() {
		if err := lgr.Sync(); err != nil {
			log.Printf("unable to Sync logger: %v\n", err)
		}
	}()
	zap.ReplaceGlobals(lgr)

	client, err := cli.Connect(mqttBroker, username, password, clientId)
	if err != nil {
		zap.S().Fatalf("unable to connect to mqtt broker: %v", err)
	}
	defer client.Disconnect(10)

	videoProperties := make(map[gocv.VideoCaptureProperties]float64)
	videoProperties[gocv.VideoCaptureFrameWidth] = float64(videoWidth)
	videoProperties[gocv.VideoCaptureFrameHeight] = float64(videoHeight)

	c := camera.New(client, topicBase, pubFrequency, videoProperties)
	defer c.Stop()

	cli.HandleExit(c)

	err = c.Start()
	if err != nil {
		zap.S().Fatalf("unable to start service: %v", err)
	}
}
