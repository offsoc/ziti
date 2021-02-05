/*
	Copyright NetFoundry, Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package network

import (
	"github.com/openziti/fabric/pb/ctrl_pb"
	cmap "github.com/orcaman/concurrent-map"
	"time"
)

type routeSenderController struct {
	senders cmap.ConcurrentMap // map[string]*routeSender
}

func newRouteSenderController() *routeSenderController {
	return &routeSenderController{}
}

func (self *routeSenderController) forwardRouteResult(r *Router, sessionId string, success bool) bool {
	return false
}

type routeSender struct {
	sessionId string
	circuit   *Circuit
	routeMsgs []*ctrl_pb.Route
	timeout   time.Duration
	maxTries  int
}

func newRouteSender(sessionId string, timeout time.Duration, maxTries int) *routeSender {
	return &routeSender{
		sessionId: sessionId,
		timeout:   timeout,
		maxTries:  maxTries,
	}
}

func (self *routeSender) send(circuit *Circuit, routeMsgs []*ctrl_pb.Route) error {
	return nil
}
