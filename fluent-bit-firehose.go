// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package main

import (
	"C"
	"fmt"
	"unsafe"

	"github.com/awslabs/amazon-kinesis-firehose-for-fluent-bit/firehose"
	"github.com/fluent/fluent-bit-go/output"
)
import "github.com/Sirupsen/logrus"

var (
	out *firehose.FirehoseOutput
)

// The "export" comments have syntactic meaning
// This is how the compiler knows a function should be callable from the C code

//export FLBPluginRegister
func FLBPluginRegister(ctx unsafe.Pointer) int {
	return output.FLBPluginRegister(ctx, "firehose", "Amazon Kinesis Data Firehose Fluent Bit Plugin.")
}

//export FLBPluginInit
func FLBPluginInit(ctx unsafe.Pointer) int {
	plugins.SetupLogger()

	deliveryStream := output.FLBPluginConfigKey(ctx, "delivery-stream")
	logrus.Infof("[firehose] plugin parameter = '%s'\n", deliveryStream)
	region := output.FLBPluginConfigKey(ctx, "region")
	logrus.Infof("[firehose] plugin parameter = '%s'\n", region)
	dataKeys := output.FLBPluginConfigKey(ctx, "data_keys")
	logrus.Infof("[firehose] plugin parameter = '%s'\n", dataKeys)
	roleARN := output.FLBPluginConfigKey(ctx, "role_arn")
	logrus.Infof("[firehose] plugin parameter = '%s'\n", roleARN)

	if deliveryStream == "" || region == "" {
		return output.FLB_ERROR
	}

	var err error
	out, err = firehose.NewFirehoseOutput(region, deliveryStream, dataKeys, roleARN)
	if err != nil {
		logrus.Debugf("firehose: Failed to initialize plugin: %v\n", err)
		return output.FLB_ERROR
	}
	return output.FLB_OK
}

//export FLBPluginFlush
func FLBPluginFlush(data unsafe.Pointer, length C.int, tag *C.char) int {
	var count int
	var ret int
	var record map[interface{}]interface{}

	// Create Fluent Bit decoder
	dec := output.NewDecoder(data, int(length))

	fluentTag := C.GoString(tag)
	logrus.Debugf("firehose: Found logs with tag: %s\n", fluentTag)

	for {
		// Extract Record
		ret, _, record = output.GetRecord(dec)
		if ret != 0 {
			break
		}

		err := out.AddRecord(record)
		if err != nil {
			return output.FLB_ERROR
		}
		count++
	}
	err := out.Flush()
	if err != nil {
		return output.FLB_ERROR
	}
	fmt.Printf("Processed %d events with tag %s\n", count, fluentTag)

	return output.FLB_OK
}

//export FLBPluginExit
func FLBPluginExit() int {
	return output.FLB_OK
}

func main() {
}
