package ibmmq

import (
	"fmt"

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
	fmt.Println("Write")
}

func (IbmMQ) Read() {
	fmt.Println("Read")
}