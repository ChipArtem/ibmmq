package ibmmq

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"
	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"go.k6.io/k6/js/modules"
)

var i int32 = 0

func init() {
	modules.Register("k6/x/ibmq", new(IbmMQ))
}

type IbmMQ struct {
	mng *MngIBMMQ
}

func (i *IbmMQ) Connect(QM, Host, Port, Channel, User, Pass, AppName, qIn, qOut string) *IbmMQ {
	p, err := strconv.Atoi(Port)
	if err != nil {
		log.Fatalf("convert port string to Int: %v", err)
	}
	mng, err := NewIbmMqMng(QM, Host, p, Channel, User, Pass, AppName, qIn, qOut)
	if err != nil {
		log.Fatalf("NewIbmMqMng: %v", err)
	}
	i.mng = mng
	return i
}

func (i *IbmMQ) Write() error {
	err := i.mng.producer.SendBytes(i.mng.queueIn, []byte("body"))
	if err == nil {
		return fmt.Errorf("sendMessage: %v", err)
	}
	return nil
}

func (i *IbmMQ) Read() error {
	for c := 0; c <= 50; c++ {
		rcvBody, err := i.mng.consumer.ReceiveBytesBodyNoWait()
		if err != nil {
			return fmt.Errorf("ReadMessage: %v", err)
		}
		if rcvBody != nil {
			return nil
		}
		time.Sleep(10*time.Millisecond)
	}
	return fmt.Errorf("No message received")
}

func (i *IbmMQ) Close(){
	i.Close()
}
