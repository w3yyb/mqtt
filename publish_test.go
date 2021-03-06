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
	"bytes"
	"testing"

	"github.com/dataence/assert"
)

func TestPublishMessageFixedHeaderFields(t *testing.T) {
	msg := NewPublishMessage()
	msg.flags = 11

	assert.True(t, true, msg.Dup(), "Incorrect DUP flag.")

	assert.True(t, true, msg.Retain(), "Incorrect RETAIN flag.")

	assert.Equal(t, true, 1, msg.QoS(), "Incorrect QoS.")

	msg.SetDup(false)
	assert.False(t, true, msg.Dup(), "Incorrect DUP flag.")

	msg.SetRetain(false)
	assert.False(t, true, msg.Retain(), "Incorrect RETAIN flag.")

	err := msg.SetQoS(2)
	assert.NoError(t, true, err, "Error setting QoS.")

	assert.Equal(t, true, 2, msg.QoS(), "Incorrect QoS.")

	err = msg.SetQoS(3)
	assert.Error(t, true, err)

	err = msg.SetQoS(0)
	assert.NoError(t, true, err, "Error setting QoS.")

	assert.Equal(t, true, 0, msg.QoS(), "Incorrect QoS.")

	msg.SetDup(true)
	assert.True(t, true, msg.Dup(), "Incorrect DUP flag.")

	msg.SetRetain(true)
	assert.True(t, true, msg.Retain(), "Incorrect RETAIN flag.")
}

func TestPublishMessageFields(t *testing.T) {
	msg := NewPublishMessage()

	msg.SetTopic([]byte("coolstuff"))
	assert.Equal(t, true, "coolstuff", string(msg.Topic()), "Error setting message topic.")

	err := msg.SetTopic([]byte("coolstuff/#"))
	assert.Error(t, true, err)

	msg.SetPacketId(100)
	assert.Equal(t, true, 100, msg.PacketId(), "Error setting acket ID.")

	msg.SetPayload([]byte("this is a payload to be sent"))
	assert.Equal(t, true, []byte("this is a payload to be sent"), msg.Payload(), "Error setting payload.")
}

func TestPublishMessageDecode1(t *testing.T) {
	msgBytes := []byte{
		byte(PUBLISH<<4) | 2,
		23,
		0, // topic name MSB (0)
		7, // topic name LSB (7)
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0, // packet ID MSB (0)
		7, // packet ID LSB (7)
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	src := bytes.NewBuffer(msgBytes)
	msg := NewPublishMessage()

	n, err := msg.Decode(src)
	assert.NoError(t, true, err, "Error decoding message.")

	assert.Equal(t, true, len(msgBytes), n, "Error decoding message.")

	assert.Equal(t, true, 7, msg.PacketId(), "Error decoding message.")

	assert.Equal(t, true, "surgemq", string(msg.Topic()), "Error deocding topic name.")

	assert.Equal(t, true, []byte{'s', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e'}, msg.Payload(), "Error deocding payload.")
}

// test insufficient bytes
func TestPublishMessageDecode2(t *testing.T) {
	msgBytes := []byte{
		byte(PUBLISH<<4) | 2,
		26,
		0, // topic name MSB (0)
		7, // topic name LSB (7)
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0, // packet ID MSB (0)
		7, // packet ID LSB (7)
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	src := bytes.NewBuffer(msgBytes)
	msg := NewPublishMessage()

	_, err := msg.Decode(src)
	assert.Error(t, true, err)
}

// test qos = 0 and no client id
func TestPublishMessageDecode3(t *testing.T) {
	msgBytes := []byte{
		byte(PUBLISH << 4),
		21,
		0, // topic name MSB (0)
		7, // topic name LSB (7)
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	src := bytes.NewBuffer(msgBytes)
	msg := NewPublishMessage()

	_, err := msg.Decode(src)
	assert.NoError(t, true, err, "Error decoding message.")
}

func TestPublishMessageEncode(t *testing.T) {
	msgBytes := []byte{
		byte(PUBLISH<<4) | 2,
		23,
		0, // topic name MSB (0)
		7, // topic name LSB (7)
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0, // packet ID MSB (0)
		7, // packet ID LSB (7)
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	msg := NewPublishMessage()
	msg.SetTopic([]byte("surgemq"))
	msg.SetQoS(1)
	msg.SetPacketId(7)
	msg.SetPayload([]byte{'s', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e'})

	dst, n, err := msg.Encode()
	assert.NoError(t, true, err, "Error decoding message.")

	assert.Equal(t, true, len(msgBytes), n, "Error decoding message.")

	assert.Equal(t, true, msgBytes, dst.(*bytes.Buffer).Bytes(), "Error decoding message.")
}

// test empty topic name
func TestPublishMessageEncode2(t *testing.T) {
	msg := NewPublishMessage()
	msg.SetTopic([]byte(""))
	msg.SetPacketId(7)
	msg.SetPayload([]byte{'s', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e'})

	_, _, err := msg.Encode()
	assert.Error(t, true, err)
}

// test encoding qos = 0 and no packet id
func TestPublishMessageEncode3(t *testing.T) {
	msgBytes := []byte{
		byte(PUBLISH << 4),
		21,
		0, // topic name MSB (0)
		7, // topic name LSB (7)
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	msg := NewPublishMessage()
	msg.SetTopic([]byte("surgemq"))
	msg.SetQoS(0)
	msg.SetPayload([]byte{'s', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e'})

	dst, n, err := msg.Encode()
	assert.NoError(t, true, err, "Error decoding message.")

	assert.Equal(t, true, len(msgBytes), n, "Error decoding message.")

	assert.Equal(t, true, msgBytes, dst.(*bytes.Buffer).Bytes(), "Error decoding message.")
}
