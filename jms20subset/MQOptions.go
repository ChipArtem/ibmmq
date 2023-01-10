package jms20subset

import ibmmqv5 "github.com/ChipArtem/k6ibmmq/ibmmq"

type MQOptions func(cno *ibmmqv5.MQCNO)

func WithMaxMsgLength(maxMsgLength int32) MQOptions {
	return func(cno *ibmmqv5.MQCNO) {
		cno.ClientConn.MaxMsgLength = maxMsgLength
	}
}
