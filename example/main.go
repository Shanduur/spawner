package main

import (
	"bufio"
	"bytes"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	buff := bytes.NewBuffer([]byte{})
	c := make(chan bool)

	l := logrus.New()
	l.SetOutput(buff)

	go func() {
		for i := 0; i < 10; i++ {
			l.Info("hello")
			time.Sleep(time.Second)
		}
		c <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(time.Second)
			fmt.Printf("BUFFER %d:\n------\n%s\n------\n", i, buff.String())
		}
		c <- true
	}()

	scanner := bufio.NewScanner(buff)
	i := 0
	for scanner.Scan() {
		fmt.Printf("%d: %s\n", i, scanner.Text())
		i++
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("error: %s\n", err.Error())
	}

	fmt.Printf("scanner done\n")

	<-c
	<-c
}
