package k6ibmmq

import (
	"io"
	"log"
	
	"os"
	"strconv"
	"time"

	"github.com/ibm-messaging/mq-golang-jms20/mqjms"

	"go.k6.io/k6/js/modules"
)

var request []byte

func init() {
	modules.Register("k6/x/K6ibmq", new(K6ibmmq))
	request = getMsg("request.100kb.xml")
}

func check(msg string, e error) {
	if e != nil {
		log.Fatalf("%v: %v", msg, e)
	}
}

func getMsg(path string) []byte {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Error reading file %v with request: %v", path, err)
	}
	defer file.Close() 

	data := make([]byte, 64)

	msg := make([]byte, 0)
	for{
        n, err := file.Read(data)
		
        if err == io.EOF{   
            break           
        }
		msg = append(msg, data[:n]...)
		
    }	
	return msg
}

type K6ibmmq struct {
}

type MQconnect struct {
	context  *mqjms.ContextImpl
	factory  *mqjms.ConnectionFactoryImpl
	producer *mqjms.ProducerImpl
	consumer *mqjms.ConsumerImpl
	qInName  string
	qOutName string
	qIn      *mqjms.QueueImpl
	qOut     *mqjms.QueueImpl
}

func (i *K6ibmmq) New(QM, Host, Port, Channel, qIn, qOut string) *MQconnect {
	p, err := strconv.Atoi(Port)
	check("Error casting string 'Port' to int", err)
	factory := &mqjms.ConnectionFactoryImpl{
		QMName:      QM,
		Hostname:    Host,
		PortNumber:  p,
		ChannelName: Channel,
		UserName:    os.Getenv("IBMMQ_USER"),
		Password:    os.Getenv("IBMMQ_PASS"),
		ApplName:    "load_test",
	}
	factory.ReceiveBufferSize = len(request) + 50

	return &MQconnect{factory: factory, qOutName: qOut, qInName: qIn}
}

func (mq *MQconnect) Setcredentials(user, pass string) *MQconnect {
	mq.factory.UserName = user
	mq.factory.Password = pass
	return mq
}

func (mq *MQconnect) Connect() *MQconnect {
	if len(mq.factory.UserName) == 0 {
		log.Fatalf("Set login to IBMMQ_USER environment variable or use function setcredentials(user, pass).")
	}
	if len(mq.factory.Password) == 0 {
		log.Fatalf("Set password to IBMMQ_PASS environment variable or use function setcredentials(user, pass).")
	}
	context, err := mq.factory.CreateContext()
	context_ := context.(mqjms.ContextImpl)
	mq.context = &context_
	check("Context creation error: ", err)

	producer := context.CreateProducer()
	producer.SetTimeToLive(3000)
	mq.producer = producer.(*mqjms.ProducerImpl)

	consumer, err := context.CreateConsumer(context.CreateQueue(mq.qOutName))
	if err != nil {
		log.Fatalf("Consumer creation error:%v", err)
		return nil
	}
	consumer_ := consumer.(mqjms.ConsumerImpl)
	mq.consumer = &consumer_

	q_ := context.CreateQueue(mq.qInName)
	qIn_ := q_.(mqjms.QueueImpl)
	mq.qIn = &qIn_
	return mq
}

func (mq *MQconnect) Checkmsg() int {
	t1 := time.Now()
	err := mq.producer.SendBytes(mq.qIn, request)
	if err != nil {
		log.Printf("Message sending error: %v", err)
		return -1
	}

	for conter := 0; conter <= 100; conter++ {
		answer, err := mq.consumer.ReceiveNoWait()
		if err != nil {
			log.Printf("Error reading message from queue: %v", err)
			return -1
		}
		if answer != nil {
			// if (answer.GetJMSExpiration() - answer.GetJMSTimestamp()) < 800 {
			// 	return 0, fmt.Errorf("Messages take a long time to be delivered %vms.", (int64(1000) - answer.GetJMSExpiration() + answer.GetJMSTimestamp()))
			// }

			msg := answer.(*mqjms.BytesMessageImpl)
			b := *(msg.ReadBytes())

			if string(b) != string(request) {
				log.Printf("Messages are different.")
				return -1
			}
			diff := int(time.Now().Sub(t1).Milliseconds())
			return diff

		}
		time.Sleep(5 * time.Millisecond)
	}
	log.Printf("Messages are not in queue")
	return -1
}

func (mq *MQconnect) Close() {
	if mq.consumer != nil {
		mq.consumer.Close()
	}
	if mq.context != nil {
		mq.context.Close()
	}
}
