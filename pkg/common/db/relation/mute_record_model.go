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

	"github.com/OpenIMSDK/tools/ormutil"

	"gorm.io/gorm"

	"github.com/OpenIMSDK/tools/utils"

	"github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

type MuteRecordGorm struct {
	*MetaDB
}

func NewMuteRecordDB(db *gorm.DB) relation.MuteRecordModelInterface {
	return &MuteRecordGorm{NewMetaDB(db, &relation.MuteRecordModel{})}
}

func (b *MuteRecordGorm) NewTx(tx any) relation.MuteRecordModelInterface {
	return &MuteRecordGorm{NewMetaDB(tx.(*gorm.DB), &relation.MuteRecordModel{})}
}

func (b *MuteRecordGorm) Create(ctx context.Context, mute_records []*relation.MuteRecordModel) (err error) {
	return utils.Wrap(b.db(ctx).Create(&mute_records).Error, "")
}

func (b *MuteRecordGorm) Delete(ctx context.Context, mute_records []*relation.MuteRecordModel) (err error) {
	return utils.Wrap(b.db(ctx).Delete(mute_records).Error, "")
}

func (b *MuteRecordGorm) Take(ctx context.Context, blockUserID string, serverID string) (muteRecord *relation.MuteRecordModel, err error) {
	return muteRecord, utils.Wrap(b.db(ctx).Where("server_id = ? and block_user_id = ?", serverID, blockUserID).Take(&muteRecord).Error, "")
}

func (b *MuteRecordGorm) FindServerMuteRecords(
	ctx context.Context,
	serverID string,
	pageNumber, showNumber int32,
) (mute_records []*relation.MuteRecordModel, total int64, err error) {
	err = b.db(ctx).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Wrap(err, "")
	}
	totalUint32, mute_records, err := ormutil.GormPage[relation.MuteRecordModel](
		b.db(ctx).Where("server_id = ?", serverID),
		pageNumber,
		showNumber,
	)
	total = int64(totalUint32)
	return
}
