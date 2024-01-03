// Copyright © 2023 OpenIM. All rights reserved.
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

package cmd

import (
	"github.com/OpenIMSDK/protocol/constant"
	"github.com/openimsdk/open-im-server/v3/internal/tools"
	v3config "github.com/openimsdk/open-im-server/v3/pkg/common/config"
	"github.com/spf13/cobra"
)

type CronTaskCmd struct {
	*RootCmd
}

func NewCronTaskCmd() *CronTaskCmd {
	ret := &CronTaskCmd{NewRootCmd("cronTask", WithCronTaskLogName())}
	ret.SetRootCmdPt(ret)
	return ret
}

func (c *CronTaskCmd) addRunE() {
	c.Command.RunE = func(cmd *cobra.Command, args []string) error {
		return tools.StartTask(c.getPortFlag(cmd), c.getPrometheusPortFlag(cmd))
	}
}

func (c *CronTaskCmd) Exec() error {
	c.addRunE()
	return c.Execute()
}

func (c *CronTaskCmd) GetPortFromConfig(portType string) int {
	if portType == constant.FlagPort {
		return v3config.Config.RpcPort.OpenImCronPort[0]
	} else if portType == constant.FlagPrometheusPort {
		return v3config.Config.Prometheus.CronPrometheusPort[0]
	} else {
		return 0
	}
}
