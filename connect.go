// Copyright (c) 2014 Dataence, LLC. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mqtt

import (
	"fmt"
	"io"
)

// After a Network Connection is established by a Client to a Server, the first Packet
// sent from the Client to the Server MUST be a CONNECT Packet [MQTT-3.1.0-1].
//
// A Client can only send the CONNECT Packet once over a Network Connection. The Server
// MUST process a second CONNECT Packet sent from a Client as a protocol violation and
// disconnect the Client [MQTT-3.1.0-2].  See section 4.8 for information about
// handling errors.
type ConnectMessage struct {
	fixedHeader

	// 7: username flag
	// 6: password flag
	// 5: will retain
	// 4-3: will QoS
	// 2: will flag
	// 1: clean session
	// 0: reserved
	connectFlags byte

	version byte

	keepAlive uint16

	protoName,
	clientId,
	willTopic,
	willMessage,
	username,
	password []byte
}

var _ Message = (*ConnectMessage)(nil)

// NewConnectMessage creates a new CONNECT message.
func NewConnectMessage() *ConnectMessage {
	msg := &ConnectMessage{}
	msg.SetType(CONNECT)

	return msg
}

// String returns a string representation of the CONNECT message
func (this ConnectMessage) String() string {
	return fmt.Sprintf("%v\nConnect Flags: %08b\nVersion: %d\nKeepAlive: %d\nClient ID: %s\nWill Topic: %s\nWill Message: %s\nUsername: %s\nPassword: %s\n",
		this.fixedHeader,
		this.connectFlags,
		this.Version(),
		this.KeepAlive(),
		this.ClientId(),
		this.WillTopic(),
		this.WillMessage(),
		this.Username(),
		this.Password(),
	)
}

// Version returns the the 8 bit unsigned value that represents the revision level
// of the protocol used by the Client. The value of the Protocol Level field for
// the version 3.1.1 of the protocol is 4 (0x04).
func (this *ConnectMessage) Version() byte {
	return this.version
}

// SetVersion sets the version value of the CONNECT message
func (this *ConnectMessage) SetVersion(v byte) error {
	if _, ok := SupportedVersions[v]; !ok {
		return fmt.Errorf("connect/SetVersion: Invalid version number %d", v)
	}

	this.version = v
	return nil
}

// CleanSession returns the bit that specifies the handling of the Session state.
// The Client and Server can store Session state to enable reliable messaging to
// continue across a sequence of Network Connections. This bit is used to control
// the lifetime of the Session state.
func (this *ConnectMessage) CleanSession() bool {
	return ((this.connectFlags >> 1) & 0x1) == 1
}

// SetCleanSession sets the bit that specifies the handling of the Session state.
func (this *ConnectMessage) SetCleanSession(v bool) {
	if v {
		this.connectFlags |= 0x2 // 00000010
	} else {
		this.connectFlags &= 253 // 11111101
	}
}

// WillFlag returns the bit that specifies whether a Will Message should be stored
// on the server. If the Will Flag is set to 1 this indicates that, if the Connect
// request is accepted, a Will Message MUST be stored on the Server and associated
// with the Network Connection.
func (this *ConnectMessage) WillFlag() bool {
	return ((this.connectFlags >> 2) & 0x1) == 1
}

// SetWillFlag sets the bit that specifies whether a Will Message should be stored
// on the server.
func (this *ConnectMessage) SetWillFlag(v bool) {
	if v {
		this.connectFlags |= 0x4 // 00000100
	} else {
		this.connectFlags &= 251 // 11111011
	}
}

// WillQos returns the two bits that specify the QoS level to be used when publishing
// the Will Message.
func (this *ConnectMessage) WillQos() byte {
	return (this.connectFlags >> 3) & 0x3
}

// SetWillQos sets the two bits that specify the QoS level to be used when publishing
// the Will Message.
func (this *ConnectMessage) SetWillQos(qos byte) error {
	if qos != QosAtMostOnce && qos != QosAtLeastOnce && qos != QosExactlyOnce {
		return fmt.Errorf("connect/SetWillQos: Invalid QoS level %d", qos)
	}

	this.connectFlags = (this.connectFlags & 231) | (qos << 3) // 231 = 11100111
	return nil
}

// WillRetain returns the bit specifies if the Will Message is to be Retained when it
// is published.
func (this *ConnectMessage) WillRetain() bool {
	return ((this.connectFlags >> 5) & 0x1) == 1
}

// SetWillRetain sets the bit specifies if the Will Message is to be Retained when it
// is published.
func (this *ConnectMessage) SetWillRetain(v bool) {
	if v {
		this.connectFlags |= 32 // 00100000
	} else {
		this.connectFlags &= 223 // 11011111
	}
}

// UsernameFlag returns the bit that specifies whether a user name is present in the
// payload.
func (this *ConnectMessage) UsernameFlag() bool {
	return ((this.connectFlags >> 7) & 0x1) == 1
}

// SetUsernameFlag sets the bit that specifies whether a user name is present in the
// payload.
func (this *ConnectMessage) SetUsernameFlag(v bool) {
	if v {
		this.connectFlags |= 128 // 10000000
	} else {
		this.connectFlags &= 127 // 01111111
	}
}

// PasswordFlag returns the bit that specifies whether a password is present in the
// payload.
func (this *ConnectMessage) PasswordFlag() bool {
	return ((this.connectFlags >> 6) & 0x1) == 1
}

// SetPasswordFlag sets the bit that specifies whether a password is present in the
// payload.
func (this *ConnectMessage) SetPasswordFlag(v bool) {
	if v {
		this.connectFlags |= 64 // 01000000
	} else {
		this.connectFlags &= 191 // 10111111
	}
}

// KeepAlive returns a time interval measured in seconds. Expressed as a 16-bit word,
// it is the maximum time interval that is permitted to elapse between the point at
// which the Client finishes transmitting one Control Packet and the point it starts
// sending the next.
func (this *ConnectMessage) KeepAlive() uint16 {
	return this.keepAlive
}

// SetKeepAlive sets the time interval in which the server should keep the connection
// alive.
func (this *ConnectMessage) SetKeepAlive(v uint16) {
	this.keepAlive = v
}

// ClientId returns an ID that identifies the Client to the Server. Each Client
// connecting to the Server has a unique ClientId. The ClientId MUST be used by
// Clients and by Servers to identify state that they hold relating to this MQTT
// Session between the Client and the Server
func (this *ConnectMessage) ClientId() []byte {
	return this.clientId
}

// SetClientId sets an ID that identifies the Client to the Server.
func (this *ConnectMessage) SetClientId(v []byte) error {
	if len(v) > 0 && !ValidClientId(v) {
		return ErrIdentifierRejected
	}

	this.clientId = v
	return nil
}

// WillTopic returns the topic in which the Will Message should be published to.
// If the Will Flag is set to 1, the Will Topic must be in the payload.
func (this *ConnectMessage) WillTopic() []byte {
	return this.willTopic
}

// SetWillTopic sets the topic in which the Will Message should be published to.
func (this *ConnectMessage) SetWillTopic(v []byte) {
	this.willTopic = v

	if len(v) > 0 {
		this.SetWillFlag(true)
	} else if len(this.willMessage) == 0 {
		this.SetWillFlag(false)
	}
}

// WillMessage returns the Will Message that is to be published to the Will Topic.
func (this *ConnectMessage) WillMessage() []byte {
	return this.willMessage
}

// SetWillMessage sets the Will Message that is to be published to the Will Topic.
func (this *ConnectMessage) SetWillMessage(v []byte) {
	this.willMessage = v

	if len(v) > 0 {
		this.SetWillFlag(true)
	} else if len(this.willTopic) == 0 {
		this.SetWillFlag(false)
	}
}

// Username returns the username from the payload. If the User Name Flag is set to 1,
// this must be in the payload. It can be used by the Server for authentication and
// authorization.
func (this *ConnectMessage) Username() []byte {
	return this.username
}

// SetUsername sets the username for authentication.
func (this *ConnectMessage) SetUsername(v []byte) {
	this.username = v

	if len(v) > 0 {
		this.SetUsernameFlag(true)
	} else {
		this.SetUsernameFlag(false)
	}
}

// Password returns the password from the payload. If the Password Flag is set to 1,
// this must be in the payload. It can be used by the Server for authentication and
// authorization.
func (this *ConnectMessage) Password() []byte {
	return this.password
}

// SetPassword sets the username for authentication.
func (this *ConnectMessage) SetPassword(v []byte) {
	this.password = v

	if len(v) > 0 {
		this.SetPasswordFlag(true)
	} else {
		this.SetPasswordFlag(false)
	}
}

// Decode reads from the io.Reader parameter until a full message is decoded, or
// when io.Reader returns EOF or error. The first return value is the number of
// bytes read from io.Reader. The second is error if Decode encounters any problems.
//
// For the CONNECT message, the error returned could be a ConnackReturnCode, so
// be sure to check that. Otherwise it's a generic error. If a generic error is
// returned, this Message should be considered invalid.
//
// Caller should call ValidConnackError(err) to see if the returned error is
// a Connack error. If so, caller should send the Client back the corresponding
// CONNACK message.
func (this *ConnectMessage) Decode(src io.Reader) (int, error) {
	total := 0

	n, err := this.fixedHeader.Decode(src)
	if err != nil {
		return total + n, err
	}
	total += n

	if n, err = this.decodeMessage(); err != nil {
		return total + n, err
	}
	total += n

	return total, nil
}

// Encode returns an io.Reader in which the encoded bytes can be read. The second
// return value is the number of bytes encoded, so the caller knows how many bytes
// there will be. If Encode returns an error, then the first two return values
// should be considered invalid.
// Any changes to the message after Encode() is called will invalidate the io.Reader.
func (this *ConnectMessage) Encode() (io.Reader, int, error) {
	if this.Type() != CONNECT {
		return nil, 0, fmt.Errorf("connect/Encode: Invalid message type. Expecting %d, got %d", CONNECT, this.Type())
	}

	total := 0
	var n int
	verstr, ok := SupportedVersions[this.version]
	if !ok {
		return nil, 0, fmt.Errorf("connect/Encode: Unsupported protocol version %d", this.version)
	}

	// 2 bytes protocol name length
	// n bytes protocol name
	// 1 byte protocol version
	// 1 byte connect flags
	// 2 bytes keep alive timer
	total += 2 + len(verstr) + 1 + 1 + 2

	// Add the clientID length, 2 is the length prefix
	total += 2 + len(this.clientId)

	// Add the will topic and will message length, and the length prefixes
	if this.WillFlag() {
		total += 2 + len(this.willTopic) + 2 + len(this.willMessage)
	}

	// Add the username length
	// According to the 3.1 spec, it's possible that the usernameFlag is set,
	// but the user name string is missing.
	if this.UsernameFlag() && len(this.username) > 0 {
		total += 2 + len(this.username)
	}

	// Add the password length
	// According to the 3.1 spec, it's possible that the passwordFlag is set,
	// but the password string is missing.
	if this.PasswordFlag() && len(this.password) > 0 {
		total += 2 + len(this.password)
	}

	if err := this.SetRemainingLength(int32(total)); err != nil {
		return nil, 0, err
	}

	total = 0

	_, n, err := this.fixedHeader.Encode()
	if err != nil {
		return nil, total + n, err
	}
	total += n

	if n, err = this.encodeMessage(); err != nil {
		return nil, total + n, err
	}
	total += n

	return this.buf, total, nil
}

func (this *ConnectMessage) encodeMessage() (int, error) {
	total := 0

	verstr, ok := SupportedVersions[this.version]
	if !ok {
		return 0, fmt.Errorf("connect/encodeVariableHeader: Unsupported protocol version %d", this.version)
	}

	n, err := writeLPBytes(this.buf, []byte(verstr))
	if err != nil {
		return 0, err
	}
	total += int(n)

	this.buf.WriteByte(this.version)
	total += 1

	this.buf.WriteByte(this.connectFlags)
	total += 1

	if err = writeUint16(this.buf, this.keepAlive); err != nil {
		return total, err
	}
	total += 2

	if n, err = writeLPBytes(this.buf, this.clientId); err != nil {
		return total + n, err
	}
	total += n

	if this.WillFlag() {
		if n, err = writeLPBytes(this.buf, this.willTopic); err != nil {
			return total + n, err
		}
		total += n

		if n, err = writeLPBytes(this.buf, this.willMessage); err != nil {
			return total + n, err
		}
		total += n
	}

	// According to the 3.1 spec, it's possible that the usernameFlag is set,
	// but the username string is missing.
	if this.UsernameFlag() && len(this.username) > 0 {
		if n, err = writeLPBytes(this.buf, this.username); err != nil {
			return total + n, err
		}
		total += n
	}

	// According to the 3.1 spec, it's possible that the passwordFlag is set,
	// but the password string is missing.
	if this.PasswordFlag() && len(this.password) > 0 {
		if n, err = writeLPBytes(this.buf, this.password); err != nil {
			return total + n, err
		}
		total += n
	}

	return total, nil
}

func (this *ConnectMessage) decodeMessage() (int, error) {
	var n, total int
	var err error

	if this.protoName, n, err = readLPBytes(this.buf); err != nil {
		return total + n, err
	}
	total += n

	if this.version, err = this.buf.ReadByte(); err != nil {
		return total, err
	}
	total += 1

	if verstr, ok := SupportedVersions[this.version]; !ok {
		return total, ErrUnacceptableProtocolVersion
	} else if verstr != string(this.protoName) {
		return total, ErrUnacceptableProtocolVersion
	}

	if this.connectFlags, err = this.buf.ReadByte(); err != nil {
		return total, err
	}
	total += 1

	if this.connectFlags&0x1 != 0 {
		return total, fmt.Errorf("connect/decodeMessage: Connect Flags reserved bit 0 is not 0")
	}

	if this.WillQos() > QosExactlyOnce {
		return total, fmt.Errorf("connect/decodeMessage: Invalid QoS level (%d) for %s message", this.WillQos(), this.Name())
	}

	if !this.WillFlag() && (this.WillRetain() || this.WillQos() != QosAtMostOnce) {
		return total, fmt.Errorf("connect/decodeMessage: Protocol violation: If the Will Flag (%t) is set to 0 the Will QoS (%d) and Will Retain (%t) fields MUST be set to zero", this.WillFlag(), this.WillQos(), this.WillRetain())
	}

	if this.UsernameFlag() && !this.PasswordFlag() {
		return total, fmt.Errorf("connect/decodeMessage: Username flag is set but Password flag is not set")
	}

	if this.keepAlive, err = readUint16(this.buf); err != nil {
		return total, err
	}
	total += 2

	if this.clientId, n, err = readLPBytes(this.buf); err != nil {
		return total + n, err
	}
	total += n

	// If the Client supplies a zero-byte ClientId, the Client MUST also set CleanSession to 1
	if len(this.clientId) == 0 && !this.CleanSession() {
		return total, ErrIdentifierRejected
	}

	// The ClientId must contain only characters 0-9, a-z, and A-Z
	// We also support ClientId longer than 23 encoded bytes
	// We do not support ClientId outside of the above characters
	if len(this.clientId) > 0 && !ValidClientId(this.clientId) {
		return total, ErrIdentifierRejected
	}

	if this.WillFlag() {
		if this.willTopic, n, err = readLPBytes(this.buf); err != nil {
			return total + n, err
		}
		total += n

		if this.willMessage, n, err = readLPBytes(this.buf); err != nil {
			return total + n, err
		}
		total += n
	}

	// According to the 3.1 spec, it's possible that the passwordFlag is set,
	// but the password string is missing.
	if this.UsernameFlag() && this.buf.Len() > 0 {
		if this.username, n, err = readLPBytes(this.buf); err != nil {
			return total + n, err
		}
		total += n
	}

	// According to the 3.1 spec, it's possible that the passwordFlag is set,
	// but the password string is missing.
	if this.PasswordFlag() && this.buf.Len() > 0 {
		if this.password, n, err = readLPBytes(this.buf); err != nil {
			return total + n, err
		}
		total += n
	}

	if this.buf.Len() > 0 {
		return total, fmt.Errorf("connect/decodeMessage: Invalid buffer size. Still has %d bytes at the end.", this.buf.Len())
	}

	return total, nil
}
