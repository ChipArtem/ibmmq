package ibmmq

import (
	"fmt"
	"time"

	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"

	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
)

type MngIBMMQ struct {
	context  jms20subset.JMSContext
	producer jms20subset.JMSProducer
	consumer jms20subset.JMSConsumer
	queueIn  jms20subset.Queue
	queueOut jms20subset.Queue
}

func NewIbmMqMng(QM, Host string, Port int, Channel, User, Pass, AppName, qIn, qOut string) (*MngIBMMQ, error) {
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

	mng := &MngIBMMQ{
		context:  context,
		producer: context.CreateProducer(),
		consumer: consumer,
		queueIn:  context.CreateQueue(qIn),
		queueOut: queueOut,
	}
	return mng, nil
}

func (m *MngIBMMQ) ReadMessage(body []byte) error {
	for i := 0; i <= 100; i++ {
		rcvBody, err := m.consumer.ReceiveBytesBodyNoWait()
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

func (m *MngIBMMQ) Close(){
	m.consumer.Close()
	m.context.Close()
}