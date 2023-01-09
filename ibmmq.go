package ibmmq

import (
	"fmt"
	"time"

	"go.k6.io/k6/js/modules"
)

func init() {
	modules.Register("k6/x/ibmq", new(IbmMQ))
}

type IbmMQ struct{}

func (IbmMQ) Connect() {
	fmt.Println("New connect")
}

func (IbmMQ) Write() {
	time.Sleep(100 * time.Millisecond)
	fmt.Println("Write")
}

func (IbmMQ) Read() {
	time.Sleep(100 * time.Millisecond)
	fmt.Println("Read")
}