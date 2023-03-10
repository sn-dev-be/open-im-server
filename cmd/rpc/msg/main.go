package main

import (
	"OpenIM/internal/rpc/msg"
	"OpenIM/pkg/common/cmd"
	"OpenIM/pkg/common/config"
	"fmt"
	"os"
)

func main() {
	rpcCmd := cmd.NewRpcCmd()
	rpcCmd.AddPortFlag()
	rpcCmd.AddPrometheusPortFlag()
	if err := rpcCmd.Exec(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if err := rpcCmd.StartSvr(config.Config.RpcRegisterName.OpenImMsgName, msg.Start); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}