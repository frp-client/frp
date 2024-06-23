// Copyright 2018 fatedier, fatedier@gmail.com
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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	plugin "github.com/frp-client/frp/pkg/plugin/server"
	"github.com/frp-client/frp/pkg/util/client"
	"github.com/frp-client/frp/pkg/util/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/frp-client/frp/pkg/config"
	v1 "github.com/frp-client/frp/pkg/config/v1"
	"github.com/frp-client/frp/pkg/config/v1/validation"
	"github.com/frp-client/frp/pkg/util/log"
	"github.com/frp-client/frp/pkg/util/version"
	"github.com/frp-client/frp/server"
)

var (
	cfgFile          string
	showVersion      bool
	strictConfigMode bool
	apiServer        string

	serverCfg v1.ServerConfig
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file of frps")
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "version of frps")
	rootCmd.PersistentFlags().BoolVarP(&strictConfigMode, "strict_config", "", true, "strict config parsing mode, unknown fields will cause errors")
	rootCmd.PersistentFlags().StringVarP(&apiServer, "server", "s", "https://api.example.com", "config api server")

	config.RegisterServerConfigFlags(rootCmd, &serverCfg)
}

var rootCmd = &cobra.Command{
	Use:   "frps",
	Short: "frps is the server of frp (https://github.com/fatedier/frp)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			fmt.Println(version.Full())
			return nil
		}

		var (
			svrCfg         *v1.ServerConfig
			isLegacyFormat bool
			err            error
		)
		if cfgFile != "" {
			svrCfg, isLegacyFormat, err = config.LoadServerConfig(cfgFile, strictConfigMode)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if isLegacyFormat {
				fmt.Printf("WARNING: ini format is deprecated and the support will be removed in the future, " +
					"please use yaml/json/toml format instead!\n")
			}
		} else {
			serverCfg.Complete()
			svrCfg = &serverCfg
		}
		if apiServer != "" {
			apiServer = strings.TrimRight(apiServer, "/")
			var endpoint = fmt.Sprintf("%s/api/frps/config", apiServer)
			buf, err := http.HttpJsonGet(endpoint, map[string]string{"X_CLIENT_ID": client.ClientId()})
			if err != nil {
				log.Errorf("fetch remote config error: %s", err.Error())
				os.Exit(1)
			}

			type Resp struct {
				Code uint              `json:"code"`
				Msg  string            `json:"msg"`
				Data plugin.FrpsConfig `json:"data"`
			}
			var resp Resp
			if err = json.Unmarshal(buf, &resp); err != nil {
				log.Errorf("fetch config from [%s] error: %s", endpoint, err.Error())
				os.Exit(1)
			}
			if resp.Data.BindPort <= 0 {
				log.Errorf("remote config error: %s", "bindPort<=0")
				os.Exit(1)
			}
			svrCfg.BindAddr = resp.Data.BindAddr
			svrCfg.BindPort = int(resp.Data.BindPort)
			if resp.Data.VhostHTTPPort > 0 {
				svrCfg.VhostHTTPPort = resp.Data.VhostHTTPPort
			}
			if resp.Data.VhostHTTPSPort > 0 {
				svrCfg.VhostHTTPSPort = resp.Data.VhostHTTPSPort
			}
			if len(resp.Data.HttpPlugins) > 0 {
				svrCfg.HTTPPlugins = make([]v1.HTTPPluginOptions, 0)
				for _, httpPlugin := range resp.Data.HttpPlugins {
					svrCfg.HTTPPlugins = append(svrCfg.HTTPPlugins, v1.HTTPPluginOptions{
						Name:      httpPlugin.Name,
						Addr:      httpPlugin.Addr,
						Path:      httpPlugin.Path,
						Ops:       httpPlugin.Ops,
						TLSVerify: false,
					})
				}
			}
		}

		warning, err := validation.ValidateServerConfig(svrCfg)
		if warning != nil {
			fmt.Printf("WARNING: %v\n", warning)
		}
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if err := runServer(svrCfg); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return nil
	},
}

func Execute() {
	rootCmd.SetGlobalNormalizationFunc(config.WordSepNormalizeFunc)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runServer(cfg *v1.ServerConfig) (err error) {
	log.InitLogger(cfg.Log.To, cfg.Log.Level, int(cfg.Log.MaxDays), cfg.Log.DisablePrintColor)

	if cfgFile != "" {
		log.Infof("frps uses config file: %s", cfgFile)
	} else {
		log.Infof("frps uses command line arguments for config")
	}

	svr, err := server.NewService(cfg)
	if err != nil {
		return err
	}
	log.Infof("frps started successfully")
	svr.Run(context.Background())
	return
}
