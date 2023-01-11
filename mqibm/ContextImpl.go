// Copyright (c) IBM Corporation 2019.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0, which is available at
// http://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

// package mqibm provides the implementation of the JMS style Golang interfaces to communicate with IBM MQ.
package mqibm

import (
	"fmt"
	"strconv"
)

// ContextImpl encapsulates the objects necessary to maintain an active
// connection to an IBM MQ queue manager.
type ContextImpl struct {
	qMgr              MQQueueManager
	sessionMode       int
	receiveBufferSize int
	sendCheckCount    int
	sendCheckCountInc *int // Internal counter to keep track of async-put messages sent
}

// CreateQueue implements the logic necessary to create a provider-specific
// object representing an IBM MQ queue.
func (ctx ContextImpl) CreateQueue(queueName string) Queue {

	// Store the name of the queue
	queue := QueueImpl{
		queueName:       queueName,
		putAsyncAllowed: Destination_PUT_ASYNC_ALLOWED_AS_DEST,
	}

	return queue
}

// CreateProducer implements the logic necessary to create a JMSProducer object
// that allows messages to be sent to destinations in IBM MQ.
func (ctx ContextImpl) CreateProducer() JMSProducer {

	// Initialise the Producer with the attributes necessary for it to send
	// messages.
	producer := ProducerImpl{
		ctx:          ctx,
		deliveryMode: DeliveryMode_PERSISTENT,
		priority:     Priority_DEFAULT,
	}

	return &producer
}

// CreateConsumer creates a consumer object that allows an application to
// receive messages from the specified Destination.
func (ctx ContextImpl) CreateConsumer(dest Destination) (JMSConsumer, JMSException) {
	return ctx.CreateConsumerWithSelector(dest, "")
}

// CreateConsumerWithSelector creates a consumer object that allows an application to
// receive messages that match the specified selector from the given Destination.
func (ctx ContextImpl) CreateConsumerWithSelector(dest Destination, selector string) (JMSConsumer, JMSException) {

	// First validate the selector string format (we don't make use of it at
	// runtime until the receive is called)
	if selector != "" {
		getmqmd := NewMQMD()
		gmo := NewMQGMO()

		selectorErr := applySelector(selector, getmqmd, gmo)
		if selectorErr != nil {
			return nil, CreateJMSException("Invalid selector syntax", "MQJMS0004", selectorErr)
		}
	}

	// Set up the necessary objects to open the queue
	mqod := NewMQOD()
	var openOptions int32
	openOptions = MQOO_FAIL_IF_QUIESCING
	openOptions |= MQOO_INPUT_AS_Q_DEF
	mqod.ObjectType = MQOT_Q
	mqod.ObjectName = dest.GetDestinationName()

	var retErr JMSException
	var consumer JMSConsumer

	// Invoke the MQ command to open the queue.
	qObject, err := ctx.qMgr.Open(mqod, openOptions)

	if err == nil {

		// Success - store the necessary objects away for later use to receive
		// messages.
		consumer = ConsumerImpl{
			ctx:      ctx,
			qObject:  qObject,
			selector: selector,
		}

	} else {

		// Error occurred - extract the failure details and return to the caller.
		rcInt := int(err.(*MQReturn).MQRC)
		errCode := strconv.Itoa(rcInt)
		reason := MQItoString("RC", rcInt)
		retErr = CreateJMSException(reason, errCode, err)

	}

	return consumer, retErr
}

// CreateBrowser creates a consumer for the specified Destination so that
// an application can look at messages without removing them.
func (ctx ContextImpl) CreateBrowser(dest Destination) (QueueBrowser, JMSException) {

	// Set up the necessary objects to open the queue
	mqod := NewMQOD()
	var openOptions int32
	openOptions = MQOO_FAIL_IF_QUIESCING
	openOptions |= MQOO_INPUT_AS_Q_DEF
	openOptions |= MQOO_BROWSE // This is the important part for browsing!
	mqod.ObjectType = MQOT_Q
	mqod.ObjectName = dest.GetDestinationName()

	var retErr JMSException
	var browser QueueBrowser

	// Invoke the MQ command to open the queue.
	qObject, err := ctx.qMgr.Open(mqod, openOptions)

	if err == nil {

		// Success - store the necessary objects away for later use to receive
		// messages.
		consumer := ConsumerImpl{
			ctx:     ctx,
			qObject: qObject,
		}

		brse := int32(MQGMO_BROWSE_FIRST)

		browser = &BrowserImpl{
			browseOption: &brse,
			ConsumerImpl: consumer,
		}

	} else {

		// Error occurred - extract the failure details and return to the caller.
		rcInt := int(err.(*MQReturn).MQRC)
		errCode := strconv.Itoa(rcInt)
		reason := MQItoString("RC", rcInt)
		retErr = CreateJMSException(reason, errCode, err)

	}

	return browser, retErr
}

// CreateTextMessage is a JMS standard mechanism for creating a TextMessage.
func (ctx ContextImpl) CreateTextMessage() TextMessage {

	var bodyStr *string
	thisMsgHandle := createMsgHandle(ctx.qMgr)

	return &TextMessageImpl{
		bodyStr: bodyStr,
		MessageImpl: MessageImpl{
			msgHandle: &thisMsgHandle,
		},
	}
}

// createMsgHandle creates a new message handle object that can be used to
// store and retrieve message properties.
func createMsgHandle(qMgr MQQueueManager) MQMessageHandle {

	cmho := NewMQCMHO()
	thisMsgHandle, err := qMgr.CrtMH(cmho)

	if err != nil {
		// No easy way to pass this error back to the application without
		// changing the function signature, which could break existing
		// applications.
		fmt.Println(err)
	}

	return thisMsgHandle

}

// CreateTextMessageWithString is a JMS standard mechanism for creating a TextMessage
// and initialise it with the chosen text string.
func (ctx ContextImpl) CreateTextMessageWithString(txt string) TextMessage {

	thisMsgHandle := createMsgHandle(ctx.qMgr)

	msg := &TextMessageImpl{
		bodyStr: &txt,
		MessageImpl: MessageImpl{
			msgHandle: &thisMsgHandle,
		},
	}

	return msg
}

// CreateBytesMessage is a JMS standard mechanism for creating a BytesMessage.
func (ctx ContextImpl) CreateBytesMessage() BytesMessage {

	var thisBodyBytes *[]byte
	thisMsgHandle := createMsgHandle(ctx.qMgr)

	return &BytesMessageImpl{
		bodyBytes: thisBodyBytes,
		MessageImpl: MessageImpl{
			msgHandle: &thisMsgHandle,
		},
	}
}

// CreateBytesMessageWithBytes is a JMS standard mechanism for creating a BytesMessage.
func (ctx ContextImpl) CreateBytesMessageWithBytes(bytes []byte) BytesMessage {

	thisMsgHandle := createMsgHandle(ctx.qMgr)

	return &BytesMessageImpl{
		bodyBytes: &bytes,
		MessageImpl: MessageImpl{
			msgHandle: &thisMsgHandle,
		},
	}
}

// Commit confirms all messages that were sent under this transaction.
func (ctx ContextImpl) Commit() JMSException {

	var retErr JMSException

	if (MQQueueManager{}) != ctx.qMgr {
		err := ctx.qMgr.Cmit()

		if err != nil {

			linkedErr := err

			// Check whether this failure could be due to async put failures
			if *ctx.sendCheckCountInc == ContextImpl_TRANSACTED_ASYNCPUT_ACTIVE {

				// One or more async put messages have been sent under a transaction so we
				// need to check now whether they were successful or not.

				// Invoke the Stat call agains the queue manager to check for errors.
				sts := NewMQSTS()
				statErr := ctx.qMgr.Stat(MQSTAT_TYPE_ASYNC_ERROR, sts)

				if statErr != nil {

					// Problem occurred invoking the Stat call, pass this back to
					// the user.
					err = statErr

				} else {

					// If there are any Warnings or Failures then we have found a problem that
					// needs to be reported to the user.
					if sts.PutWarningCount+sts.PutFailureCount > 0 {

						linkedErr = populateAsyncPutError(sts)

					}

				}

			}

			rcInt := int(err.(*MQReturn).MQRC)
			errCode := strconv.Itoa(rcInt)
			reason := MQItoString("RC", rcInt)
			retErr = CreateJMSException(reason, errCode, linkedErr)

		}

	}

	return retErr
}

// Rollback releases all messages that were sent under this transaction.
func (ctx ContextImpl) Rollback() JMSException {

	var retErr JMSException

	if (MQQueueManager{}) != ctx.qMgr {
		err := ctx.qMgr.Back()

		if err != nil {

			rcInt := int(err.(*MQReturn).MQRC)
			errCode := strconv.Itoa(rcInt)
			reason := MQItoString("RC", rcInt)
			retErr = CreateJMSException(reason, errCode, err)

		}
	}

	return retErr

}

// Close this connection to the MQ queue manager, and release any resources
// that were allocated to support this connection.
func (ctx ContextImpl) Close() {

	// JMS semantics are to roll back an active transaction on Close.
	ctx.Rollback()

	if (MQQueueManager{}) != ctx.qMgr {
		ctx.qMgr.Disc()
	}

}

// ContextImpl_TRANSACTED_ASYNCPUT_ACTIVE is an internal constant that indicates that
// a transacted asynchronous put has taken place.
const ContextImpl_TRANSACTED_ASYNCPUT_ACTIVE int = -100
