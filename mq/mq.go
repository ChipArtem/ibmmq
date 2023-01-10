package mq

import 	"github.com/ibm-messaging/mq-golang-jms20/mqjms"

type mngMQ struct {
	mq mqjms.ConnectionFactoryImpl
}