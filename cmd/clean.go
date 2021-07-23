// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
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
	"github.com/fanux/sealos/pkg/sshcmd/sshutil"
	"net"
	"os"
	"strings"

	"github.com/fanux/sealos/install"
	"github.com/spf13/cobra"
	"github.com/wonderivan/logger"
)

var exampleCleanCmd = `
	# clean  master
	sealos clean --master 192.168.0.2 \
	--master 192.168.0.3
  
	# clean  node  use --force/-f will be not prompt 
	sealos clean --node 192.168.0.4 \
	--node 192.168.0.5 --force

	# clean master and node
	sealos clean --master 192.168.0.2-192.168.0.3 \
 	--node 192.168.0.4-192.168.0.5
	
	# clean your kubernets HA cluster and use --force/-f will be not prompt (danger)
	sealos clean --all -f
`

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:     "clean",
	Short:   "Simplest way to clean your kubernets HA cluster",
	Long:    `sealos clean`,
	Example: exampleCleanCmd,
	Run:     CleanCmdFunc,
}

func init() {
	rootCmd.AddCommand(cleanCmd)

	// Here you will define your flags and configuration settings.
	cleanCmd.Flags().StringSliceVar(&install.NodeIPs, "node", []string{}, "clean node ips.kubernetes multi-nodes ex. 192.168.0.5-192.168.0.5")
	cleanCmd.Flags().StringSliceVar(&install.MasterIPs, "master", []string{}, "clean master ips.kubernetes multi-nodes ex. 192.168.0.5-192.168.0.5")
	cleanCmd.PersistentFlags().BoolVarP(&install.CleanForce, "force", "f", false, "if this is true, will no prompt")
	cleanCmd.PersistentFlags().BoolVar(&install.CleanAll, "all", false, "if this is true, delete all ")
	cleanCmd.Flags().IntVar(&install.Vlog, "vlog", 0, "kubeadm log level")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cleanCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cleanCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	cleanCmd.Flags().StringVar(&install.AssignMaster, "assign-master", "", "appoint master")
	cleanCmd.Flags().StringVar(&install.SdwanUrl, "sdwan-url", "", "node publishes to this URL after clean is complete")
}

func CleanCmdFunc(cmd *cobra.Command, args []string) {
	deleteNodes := install.ParseIPs(install.NodeIPs)
	deleteMasters := install.ParseIPs(install.MasterIPs)

	if install.AssignMaster != "" {
		var array []string
		array = append(array, install.AssignMaster)
		install.AssignMaster = install.ParseIPs(install.ParsePasss(array))[0]

		extranetIp, _, _ := install.RunShellInSystem("curl ip.sb")
		extranetIp = strings.TrimSpace(extranetIp)
		if net.ParseIP(extranetIp).To4() == nil {
			logger.Error("failed to get external IP")
			os.Exit(1)
		}
		if extranetIp == install.RemoveIpPort(install.AssignMaster) {
			logger.Info("[%s] extranet ip %s is same, sikp copy remote file to local", install.AssignMaster, extranetIp)
		} else {
			logger.Info("[%s] copy remote file to local...", install.AssignMaster)
			config := sshutil.SSH{
				User:     "root",
				Password: install.SSHConfig.UserPass[install.AssignMaster],
			}
			config.CopyRemoteFileToLocal(install.AssignMaster, install.GetConfigPath(cfgFile), install.GetConfigPath(cfgFile))
		}
	}

	c := &install.SealConfig{}
	err := c.Load(cfgFile)
	if err != nil {
		// comment: if cfgFile is not exist; do not use sealos clean something.
		// its danger for sealos do clean nodes without `~/.sealos/config.yaml`
		//// 判断错误是否为配置文件不存在
		//if errors.Is(err, os.ErrNotExist) {
		//	_, err = fmt.Fprint(os.Stdout, "Please enter the password to connect to the node:\n")
		//	if err != nil {
		//		logger.Error("fmt.Fprint err", err)
		//		os.Exit(-1)
		//	}
		//	passwordTmp, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		//	if err != nil {
		//		logger.Error("read password err", err)
		//		os.Exit(-1)
		//	}
		//	install.SSHConfig.Password = string(passwordTmp)
		//} else {
		logger.Error(err)
		os.Exit(-1)
		//}
	}

	// 使用 sealos clean --node   不小心写了 masterip.
	if ok, node := deleteNodeIsExistInNodes(deleteNodes, c.Masters); ok {
		logger.Error(`clean master Use "sealos clean --master %s" to clean it, exit...`, node)
		os.Exit(-1)
	}
	// 使用 sealos clean --master 不小心写了 nodeip.
	if ok, node := deleteNodeIsExistInNodes(deleteMasters, c.Nodes); ok {
		logger.Error(`clean nodes Use "sealos clean --node %s" to clean it, exit...`, node)
		os.Exit(-1)
	}

	install.BuildClean(deleteNodes, deleteMasters)
	c.Dump(cfgFile)

}

// IsExistNodes
func deleteNodeIsExistInNodes(deleteNodes []string, nodes []string) (bool, string) {
	for _, node := range nodes {
		for _, deleteNode := range deleteNodes {
			// 如果ips 相同. 则说明删除错了.
			if node == deleteNode {
				return true, node
			}
		}
	}
	return false, ""
}
