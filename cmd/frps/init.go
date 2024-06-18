// Copyright 2021 The frp Authors
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
	"encoding/json"
	"fmt"
	"github.com/frp-client/frp/pkg/util/http"
	"github.com/spf13/cobra"
	"log"
)

func init() {
	rootCmd.AddCommand(initCmd)

}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init proxy server(生成服务器数据到数据库)",
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(serverCfg.Auth.Token) == 0 {
			// token值从api服务器获取，每隔10分钟随机更新一次（存储在内存）
			fmt.Println("frps: 使用--token参数指定token，用于认证")
			return nil
		}

		log.Println("[apiConfig]", apiConfig)
		if len(apiConfig) == 0 {
			fmt.Println("frps: 使用--apiConfig参数指定节点上报接口")
			return nil
		}

		buf, err := http.HttpJsonPost(apiConfig, []byte(fmt.Sprintf(`{"token":"%s"}`, serverCfg.Auth.Token)))
		if err != nil {
			fmt.Println("初始化数据提交失败：", err.Error())
			return nil
		}
		type Resp struct {
			Code uint        `json:"code"`
			Msg  string      `json:"msg"`
			Data interface{} `json:"data"`
		}
		var resp Resp
		if err = json.Unmarshal(buf, &resp); err != nil {
			fmt.Println("初始化数据提交失败：", err.Error())
			return nil
		}
		if resp.Code != 200 {
			fmt.Println("初始化数据提交失败：", resp.Msg)
			return nil
		}

		fmt.Println("初始化数据提交完成：", resp.Code)
		return nil
	},
}
