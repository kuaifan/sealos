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

	"github.com/spf13/cobra"

	"github.com/fanux/sealos/install"
	"github.com/wonderivan/logger"
)

// joinCmd represents the join command
var joinCmd = &cobra.Command{
	Use:   "join",
	Short: "Simplest way to join your kubernets HA cluster",
	Long:  `sealos join --node 192.168.0.5`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(install.MasterIPs) == 0 && len(install.NodeIPs) == 0 {
			logger.Error("this command is join feature,master and node is empty at the same time.please check your args in command.")
			cmd.Help()
			os.Exit(0)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		beforeNodes := install.ParseIPs(install.NodeIPs)
		beforeMasters := install.ParseIPs(install.MasterIPs)

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
			logger.Error(err)
			c.ShowDefaultConfig()
			os.Exit(0)
		}
		install.BuildJoin(beforeMasters, beforeNodes)
		c.Dump(cfgFile)
	},
}

func init() {
	rootCmd.AddCommand(joinCmd)
	joinCmd.Flags().StringSliceVar(&install.MasterIPs, "master", []string{}, "kubernetes multi-master ex. 192.168.0.5-192.168.0.5")
	joinCmd.Flags().StringSliceVar(&install.NodeIPs, "node", []string{}, "kubernetes multi-nodes ex. 192.168.0.5-192.168.0.5")
	joinCmd.Flags().IntVar(&install.Vlog, "vlog", 0, "kubeadm log level")

	joinCmd.Flags().StringVar(&install.AssignMaster, "assign-master", "", "appoint master")
	joinCmd.Flags().StringVar(&install.SdwanUrl, "sdwan-url", "", "node publishes to this URL after deployment is complete")
	joinCmd.Flags().StringVar(&install.PkgUrl, "pkg-url", "", "http://store.lameleg.com/kube1.14.1.tar.gz download offline package url, or file location ex. /root/kube1.14.1.tar.gz")
}
