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
	"strconv"
)

// ConnectionFactoryImpl defines a struct that contains attributes for
// each of the key properties required to establish a connection to an IBM MQ
// queue manager.
//
// The fields are defined as Public so that the struct can be initialised
// programmatically using whatever approach the application prefers.
type ConnectionFactoryImpl struct {
	QMName      string
	Hostname    string
	PortNumber  int
	ChannelName string
	UserName    string
	Password    string

	TransportType int // Default to TransportType_CLIENT (0)

	// Equivalent to SSLCipherSpec and SSLClientAuth in the MQI client, however
	// the names have been updated here to reflect that SSL protocols have all
	// been discredited.
	TLSCipherSpec string
	TLSClientAuth string // Default to TLSClientAuth_NONE

	KeyRepository    string
	CertificateLabel string

	// Allthough only available per MQ 9.1.2 it looks like a good idea to have this present in MQ-JMS
	ApplName string

	// Controls the size of the buffer used when receiving a message (default is 32kb if not set)
	ReceiveBufferSize int

	// SetCheckCount defines the number of messages that will be asynchronously put using
	// this Context between checks for errors. For example a value of 10 will cause an error
	// check to be triggered once for every 10 messages.
	//
	// See also Destination#SetPutAsyncAllowed(int)
	//
	// Default of 0 (zero) means that no checks are made for asynchronous put calls.
	SendCheckCount int
}

// CreateContext implements the JMS method to create a connection to an IBM MQ
// queue manager.
func (cf ConnectionFactoryImpl) CreateContext() (JMSContext, JMSException) {
	return cf.CreateContextWithSessionMode(JMSContextAUTOACKNOWLEDGE)
}

// CreateContextWithSessionMode implements the JMS method to create a connection to an IBM MQ
// queue manager using the specified session mode.
func (cf ConnectionFactoryImpl) CreateContextWithSessionMode(sessionMode int) (JMSContext, JMSException) {

	// Allocate the internal structures required to create an connection to IBM MQ.
	cno := NewMQCNO()

	if cf.TransportType == TransportType_CLIENT {

		// Indicate that we want to use a client (TCP) connection.
		cno.Options = MQCNO_CLIENT_BINDING

		// Fill in the required fields in the channel definition structure
		cd := NewMQCD()
		cd.ChannelName = cf.ChannelName
		cd.ConnectionName = cf.Hostname + "(" + strconv.Itoa(cf.PortNumber) + ")"
		cno.ClientConn = cd

		// Fill in the fields relating to TLS channel connections
		if cf.TLSCipherSpec != "" {
			cd.SSLCipherSpec = cf.TLSCipherSpec
		}

		switch cf.TLSClientAuth {
		case TLSClientAuth_REQUIRED:
			cd.SSLClientAuth = MQSCA_REQUIRED
		case TLSClientAuth_NONE:
		case "":
			cd.SSLClientAuth = MQSCA_OPTIONAL
		default:
			cd.SSLClientAuth = -1 // Trigger an error message
		}

		// Set up the reference to the key repository file, if it has been specified.
		if cf.KeyRepository != "" {
			sco := NewMQSCO()
			sco.KeyRepository = cf.KeyRepository

			if cf.CertificateLabel != "" {
				sco.CertificateLabel = cf.CertificateLabel
			}

			cno.SSLConfig = sco

		}

		// Fill in the optional (possible since MQ 9.1.2) application name
		cno.ApplName = cf.ApplName

	} else if cf.TransportType == TransportType_BINDINGS {

		// Indicate to use Bindings connections.
		cno.Options = MQCNO_LOCAL_BINDING

	}

	if cf.UserName != "" {

		// Store the user credentials in an MQCSP, which ensures that long passwords
		// can be used.
		csp := NewMQCSP()
		csp.AuthenticationType = MQCSP_AUTH_USER_ID_AND_PWD
		csp.UserId = cf.UserName
		csp.Password = cf.Password
		cno.SecurityParms = csp

	}

	// Apply options

	var ctx JMSContext
	var retErr JMSException

	// Use the objects that we have configured to create a connection to the
	// queue manager.
	qMgr, err := Connx(cf.QMName, cno)

	if err == nil {

		// Initialize the countInc value to 1 so that if CheckCount is enabled (>0)
		// then an error check will be made after the first message - to catch any
		// failures quickly.
		countInc := new(int)
		*countInc = 1

		// Connection was created successfully, so we wrap the MQI object into
		// a new ContextImpl and return it to the caller.
		ctx = ContextImpl{
			qMgr:              qMgr,
			sessionMode:       sessionMode,
			receiveBufferSize: cf.ReceiveBufferSize,
			sendCheckCount:    cf.SendCheckCount,
			sendCheckCountInc: countInc,
		}

	} else {

		// The underlying MQI call returned an error, so extract the relevant
		// details and pass it back to the caller as a JMSException
		rcInt := int(err.(*MQReturn).MQRC)
		errCode := strconv.Itoa(rcInt)
		reason := MQItoString("RC", rcInt)
		retErr = CreateJMSException(reason, errCode, err)

	}

	return ctx, retErr

}
