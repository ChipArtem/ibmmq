package k6ibmmq

import (
	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"go.k6.io/k6/js/modules"
)

func init() {
	modules.Register("k6/x/k6ibmmq", new(mngMQ))
}

type mngMQ struct {
	forImport mqjms.ConnectionFactoryImpl
}