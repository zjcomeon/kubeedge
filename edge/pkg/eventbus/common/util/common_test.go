/*
Copyright 2019 The KubeEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/klog/v2"

	eventconfig "github.com/kubeedge/kubeedge/edge/pkg/eventbus/config"
	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/edgecore/v1alpha1"
)

var clientOptions = MQTT.NewClientOptions()

func init() {
	nodeName := "testEdge"
	cfg := v1alpha1.NewDefaultEdgeCoreConfig()
	eventconfig.InitConfigure(cfg.Modules.EventBus, nodeName)
}

//TestCheckKeyExist checks the functionality of CheckKeyExist function
func TestCheckKeyExist(t *testing.T) {
	tests := []struct {
		name          string
		keys          []string
		disinfo       map[string]interface{}
		expectedError error
	}{
		{
			name:          "TestCheckKeyExist: Key exists in passed map",
			keys:          []string{"key1"},
			disinfo:       map[string]interface{}{"key1": "value1"},
			expectedError: nil,
		},
		{
			name:          "TestCheckKeyExist: Key does not exists in passed map",
			keys:          []string{"key1"},
			disinfo:       map[string]interface{}{"key2": "value2"},
			expectedError: errors.New("key not found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckKeyExist(tt.keys, tt.disinfo)
			if !reflect.DeepEqual(tt.expectedError, err) {
				t.Errorf("Expected error contain %s, but error is %v", tt.expectedError, err)
			}
		})
	}
}

//TestCheckClientToken checks client token received
func TestCheckClientToken(t *testing.T) {
	tests := []struct {
		name          string
		token         MQTT.Token
		expectedError string
	}{
		{
			name:          "TestCheckClientToken: Client Token with no error",
			token:         MQTT.NewClient(clientOptions).Connect(),
			expectedError: "",
		},
		{
			name:          "TestCheckClientToken: Client token created with error",
			token:         MQTT.NewClient(HubClientInit("tcp://127.0.0:8000", "12345", "", "")).Connect(),
			expectedError: "Network Error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs, err := CheckClientToken(tt.token)
			fmt.Printf("rs  =  %v", rs)
			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error contain %s, but error is %v", tt.expectedError, err)
			}
		})
	}
}

//TestPathExist checks the functionality of PathExist function
func TestPathExist(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "TestPathExist: Path Exist",
			path: "/",
			want: true,
		},
		{
			name: "TestPathExist: Path does not Exist",
			path: "Wrong_Path",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PathExist(tt.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("common.TestPathExist() got = %v, want =  %v", got, tt.want)
			}
		})
	}
}

//TestHubClientInit checks the HubClientInit method that it returns the same clientOptions object or not
func TestHubClientInit(t *testing.T) {
	tests := []struct {
		name          string
		server        string
		clientID      string
		username      string
		password      string
		want          *MQTT.ClientOptions
		expectedError error
	}{
		{
			name:          "TestHubclientInit: given username and password",
			server:        "tcp://127.0.0.1:1880",
			clientID:      "12345",
			username:      "test_user",
			password:      "123456789",
			want:          clientOptions,
			expectedError: nil,
		},
		{
			name:          "TestHubclientInit: given username and no password",
			server:        "tcp://127.0.0.1:1882",
			clientID:      "12345",
			username:      "test_user",
			password:      "",
			want:          clientOptions,
			expectedError: nil,
		},
		{
			name:          "TestHubclientInit: no username and password",
			server:        "tcp://127.0.0.1:1883",
			clientID:      "12345",
			username:      "",
			password:      "",
			want:          clientOptions,
			expectedError: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			brokerURI, _ := url.Parse(tt.server)
			tt.want.Servers = append([]*url.URL{}, brokerURI)
			tt.want.ClientID = tt.clientID
			tt.want.Username = tt.username
			tt.want.Password = tt.password
			tt.want.TLSConfig = &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
			got := HubClientInit(tt.server, tt.clientID, tt.username, tt.password)
			if !reflect.DeepEqual(tt.want.Servers, got.Servers) || tt.want.ClientID != got.ClientID || tt.want.CleanSession != got.CleanSession ||
				tt.want.Username != got.Username || tt.want.Password != got.Password || !reflect.DeepEqual(tt.want.TLSConfig, got.TLSConfig) {
				t.Errorf("expected %#v, but got %#v", tt.want, got)
			}
		})
	}
}

//TestLoopConnect checks LoopConnect to connect to MQTT broker
func TestLoopConnect(t *testing.T) {
	tests := []struct {
		name          string
		client        MQTT.Client
		clientID      string
		clientOptions *MQTT.ClientOptions
		connect       bool
	}{
		{
			name:          "TestLoopConnect: success in connection",
			clientID:      "12345",
			clientOptions: MQTT.NewClientOptions(),
			connect:       true,
		},
		{
			name:          "TestLoopConnect: Connection error",
			clientID:      "12345",
			clientOptions: HubClientInit("tcp://127.0.0.1:1882", "12345", "test_user", "123456789"),
			connect:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.client = MQTT.NewClient(tt.clientOptions)
			go LoopConnect(tt.clientID, tt.client)
			time.Sleep(5 * time.Millisecond)
			if !tt.client.IsConnected() {
				if len(tt.clientOptions.Servers) != 0 {
					if tt.connect {
						t.Errorf("common.TestLoopConnect() Options.Servers = %v, want connect =  %v", tt.clientOptions.Servers, tt.connect)
					}
				}
				klog.Info("No servers defined to connect to")
			}
		})
	}
}
