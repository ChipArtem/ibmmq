package ibmq

import (
	"log"

	"go.k6.io/k6/js/modules"
)

func init() {
	modules.Register("k6/x/ibmq", new(IbmMQ))
}

type IbmMQ struct{}

func (IbmMQ) Connect() {
	log.Println("New connect")
}

func (IbmMQ) Write() {
	log.Println("Write")
}

func (IbmMQ) Read() {
	log.Println("Read")
}