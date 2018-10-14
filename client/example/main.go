package main

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wangxianzhuo/iec104/client"
)

func main() {
	address := "127.0.0.1:2404"
	outChan := make(chan map[string]float32)
	testInterval := 5 * time.Second
	c, _, err := client.New(address, outChan, logrus.WithField("client", "iec104"))
	if err != nil {
		panic(err)
	}
	client.SetConnectDeadLine(1 * time.Minute)

	go c.Start(testInterval)
	defer c.Close()

	for d := range outChan {
		fmt.Println(d)
	}
}
