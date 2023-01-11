package k6ibmmq

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/ChipArtem/k6ibmmq/mqibm"
	"go.k6.io/k6/js/modules"
)

var i int32 = 0

func init() {
	modules.Register("k6/x/K6ibmq", new(K6ibmmq))
}

type K6ibmmq struct {
	mng *mngMQ
}

type mngMQ struct {
	context  JMSContext
	producer JMSProducer
	consumer JMSConsumer
	queueIn  Queue
	queueOut Queue
}

func NewIbmMqMng(QM, Host string, Port int, Channel, User, Pass, AppName, qIn, qOut string) (*mngMQ, error) {
	connFactory := mqjms.ConnectionFactoryImpl{
		QMName:      QM,
		Hostname:    Host,
		PortNumber:  Port,
		ChannelName: Channel,
		UserName:    User,
		Password:    Pass,
		ApplName:    AppName,
	}
	context, err := connFactory.CreateContext()
	if err != nil {
		return nil, fmt.Errorf("CreateContext: %v", err)
	}

	queueOut := context.CreateQueue(qOut)

	consumer, err := context.CreateConsumer(queueOut)
	if err != nil {
		return nil, fmt.Errorf("CreateConsumer: %v", err)
	}

	mng := &mngMQ{
		context:  context,
		producer: context.CreateProducer(),
		consumer: consumer,
		queueIn:  context.CreateQueue(qIn),
		queueOut: queueOut,
	}
	return mng, nil
}

func (m *mngMQ) Close() {
	m.consumer.Close()
	m.context.Close()
}

func (i *K6ibmmq) Connect(QM, Host, Port, Channel, User, Pass, AppName, qIn, qOut string) *K6ibmmq {
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

func (i *K6ibmmq) Write() error {
	err := i.mng.producer.SendBytes(i.mng.queueIn, []byte("body"))
	if err == nil {
		return fmt.Errorf("sendMessage: %v", err)
	}
	return nil
}

func (i *K6ibmmq) Read() error {
	for c := 0; c <= 50; c++ {
		rcvBody, err := i.mng.consumer.ReceiveBytesBodyNoWait()
		if err != nil {
			return fmt.Errorf("ReadMessage: %v", err)
		}
		if rcvBody != nil {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("No message received")
}

func (i *K6ibmmq) Close() {
	i.Close()
}
