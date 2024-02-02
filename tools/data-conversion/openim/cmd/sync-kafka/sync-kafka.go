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

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/IBM/sarama"
	"github.com/OpenIMSDK/protocol/common"
	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/tools/errs"

	"github.com/OpenIMSDK/tools/utils"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
	"github.com/openimsdk/open-im-server/v3/pkg/common/kafka"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	mysqlUsername  = "root"
	mysqlPasswword = "openIM123"
	mysqlAddr      = "172.31.40.48:13306"
	// mysqlAddr      = "192.168.2.222:3306"
	// mysqlPasswword = "123456"
	mysqlDatabase = "openIM_v3"
)

var (
	topic     = "openim-to-mimo"
	kafkaAddr = "172.31.40.48:9092"
	//kafkaAddr     = "192.168.2.222:9092"
	kafkaUsername = ""
	kafkaPassword = ""
)

func checkMysql() (*gorm.DB, error) {
	var sqlDB *sql.DB
	defer func() {
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		mysqlUsername, mysqlPasswword, mysqlAddr, mysqlDatabase)
	db, err := gorm.Open(mysql.Open(dsn), nil)
	if err != nil {
		return nil, errs.Wrap(err)
	} else {
		// sqlDB, err = db.DB()
		// err = sqlDB.Ping()
		// if err != nil {
		// 	return errs.Wrap(err)
		// }
		return db, nil
	}
}

func checkKafka() (sarama.SyncProducer, error) {
	cfg := sarama.NewConfig()
	if kafkaUsername != "" && kafkaPassword != "" {
		cfg.Net.SASL.Enable = true
		cfg.Net.SASL.User = kafkaUsername
		cfg.Net.SASL.Password = kafkaPassword
	}

	kafka.SetupTLSConfig(cfg)
	_, err := sarama.NewClient([]string{kafkaAddr}, cfg)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	cfg.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer([]string{kafkaAddr}, cfg)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	return producer, nil
}

// writeServersToKafka reads records from the 'servers' table in MySQL and writes them to Kafka.gi
func handleSyncServer() error {
	// 检查MySQL连接
	db, err := checkMysql()
	if err != nil {
		fmt.Fprintln(os.Stdout, errs.Wrap(err))
	}

	// Calculate the total number of records
	var totalRecords int64
	if err := db.Model(&relation.ServerModel{}).Count(&totalRecords).Error; err != nil {
		return errs.Wrap(err)
	}

	const chunkSize = 100
	numChunks := (totalRecords + chunkSize - 1) / chunkSize
	for i := int64(0); i < numChunks; i++ {
		offset := i * chunkSize
		limit := chunkSize

		var servers []relation.ServerModel
		if err := db.Order("create_time desc").Offset(int(offset)).Limit(int(limit)).Find(&servers).Error; err != nil {
			fmt.Fprintln(os.Stdout, errs.Wrap(err))
			continue
		}

		// Process the chunk of data
		if err := sendServersToMimoBusiness(servers); err != nil {
			fmt.Fprintln(os.Stdout, errs.Wrap(err))
		}
	}

	return nil
}

// processServersChunk processes a chunk of server records
func sendServersToMimoBusiness(servers []relation.ServerModel) error {
	// 与Kafka建立连接
	producer, err := checkKafka()
	if err != nil {
		fmt.Fprintln(os.Stdout, errs.Wrap(err))
	}
	//defer kafkaClient.Close()
	defer producer.Close()

	for _, server := range servers {
		// 将server结构体转为JSON
		serverJSON, err := json.Marshal(server)
		if err != nil {
			fmt.Fprintln(os.Stdout, errs.Wrap(err))
		}
		fmt.Fprintln(os.Stdout, string(serverJSON))

		// 创建Kafka消息
		msg := &sarama.ProducerMessage{
			Topic: topic, // 使用默认的或者提供的topic
			Value: sarama.StringEncoder(createClubServerEvent(server.ServerID, server.ServerName, server.Banner, server.Icon, server.CommunityViewMode == 1)),
		}

		// // 发送消息到Kafka
		if _, _, err := producer.SendMessage(msg); err != nil {
			return errs.Wrap(err)
		}
	}
	return nil
}

func sendServerMembersToMimoBusiness(serverMembers []relation.ServerMemberModel) error {
	// 与Kafka建立连接
	producer, err := checkKafka()
	if err != nil {
		fmt.Fprintln(os.Stdout, errs.Wrap(err))
	}
	//defer kafkaClient.Close()
	defer producer.Close()

	for _, serverMember := range serverMembers {
		// 将server结构体转为JSON
		serverMemberJSON, err := json.Marshal(serverMember)
		if err != nil {
			fmt.Fprintln(os.Stdout, errs.Wrap(err))
		}
		fmt.Fprintln(os.Stdout, string(serverMemberJSON))

		// 创建Kafka消息
		msg := &sarama.ProducerMessage{
			Topic: topic, // 使用默认的或者提供的topic
			Value: sarama.StringEncoder(createClubServerMemberEvent(serverMember.ServerID, serverMember.UserID, serverMember.Nickname)),
		}

		// // 发送消息到Kafka
		if _, _, err := producer.SendMessage(msg); err != nil {
			return errs.Wrap(err)
		}
	}
	return nil
}

func createClubServerEvent(serverID, name, banner, icon string, isPublic bool) string {
	return utils.StructToJsonString(&common.CommonBusinessMQEvent{
		ClubServer: &common.ClubServer{
			ClubServerId: serverID,
			Name:         name,
			Icon:         icon,
			Banner:       banner,
			IsPublic:     isPublic,
		},
		EventType: constant.ClubServerMQEventType,
	})
}

func createClubServerMemberEvent(serverID, userID, nickname string) string {
	return utils.StructToJsonString(&common.CommonBusinessMQEvent{
		ClubServerUser: &common.ClubServerUser{
			ServerId: serverID,
			UserId:   userID,
			Nickname: nickname,
		},
		EventType: constant.ClubServerUserMQEventType,
	})
}

func handleSyncServerUser() error {
	// 检查MySQL连接
	db, err := checkMysql()
	if err != nil {
		fmt.Fprintln(os.Stdout, errs.Wrap(err))
	}

	// Calculate the total number of records
	var totalRecords int64
	if err := db.Model(&relation.ServerMemberModel{}).Count(&totalRecords).Error; err != nil {
		return errs.Wrap(err)
	}

	const chunkSize = 100
	numChunks := (totalRecords + chunkSize - 1) / chunkSize
	for i := int64(0); i < numChunks; i++ {
		offset := i * chunkSize
		limit := chunkSize

		var serverMembers []relation.ServerMemberModel
		if err := db.Offset(int(offset)).Limit(int(limit)).Find(&serverMembers).Error; err != nil {
			fmt.Fprintln(os.Stdout, errs.Wrap(err))
			continue
		}

		// Process the chunk of data
		if err := sendServerMembersToMimoBusiness(serverMembers); err != nil {
			fmt.Fprintln(os.Stdout, errs.Wrap(err))
		}
	}

	return nil
}

func main() {
	//handleSyncServer()
	handleSyncServerUser()
}
