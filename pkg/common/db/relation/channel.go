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

package relation

import (
	"context"

	"gorm.io/gorm"

	"github.com/OpenIMSDK/tools/utils"

	"github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

var _ relation.ChannelModelInterface = (*ChannelGorm)(nil)

type ChannelGorm struct {
	*MetaDB
}

func NewChannelDB(db *gorm.DB) relation.ChannelModelInterface {
	return &ChannelGorm{NewMetaDB(db, &relation.ChannelModel{})}
}

func (s *ChannelGorm) NewTx(tx any) relation.ChannelModelInterface {
	return &ChannelGorm{NewMetaDB(tx.(*gorm.DB), &relation.ChannelModel{})}
}

func (s *ChannelGorm) Create(ctx context.Context, channels []*relation.ChannelModel) (err error) {
	return utils.Wrap(s.DB.Create(&channels).Error, "")
}

func (s *ChannelGorm) Take(ctx context.Context, channelID string) (channel *relation.ChannelModel, err error) {
	channel = &relation.ChannelModel{}
	return channel, utils.Wrap(s.DB.Where("channel_id = ?", channelID).Take(channel).Error, "")
}
