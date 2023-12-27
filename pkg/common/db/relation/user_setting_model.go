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

type UserSettingGorm struct {
	*MetaDB
}

func NewUserSettingGorm(db *gorm.DB) relation.UserSettingModelInterface {
	return &UserSettingGorm{NewMetaDB(db, &relation.UserSettingModel{})}
}

// 插入多条.
func (u *UserSettingGorm) Create(ctx context.Context, settings []*relation.UserSettingModel) (err error) {
	return utils.Wrap(u.db(ctx).Create(&settings).Error, "")
}

// 更新用户信息 零值.
func (u *UserSettingGorm) UpdateByMap(ctx context.Context, userID string, args map[string]interface{}) (err error) {
	return utils.Wrap(u.db(ctx).Model(&relation.UserSettingModel{}).Where("user_id = ?", userID).Updates(args).Error, "")
}

// 更新多个用户信息 非零值.
func (u *UserSettingGorm) Update(ctx context.Context, setting *relation.UserSettingModel) (err error) {
	return utils.Wrap(u.db(ctx).Model(setting).Updates(setting).Error, "")
}

// 获取指定用户信息  不存在，也不返回错误.
func (u *UserSettingGorm) Find(ctx context.Context, userIDs []string) (settings []*relation.UserSettingModel, err error) {
	err = utils.Wrap(u.db(ctx).Where("user_id in (?)", userIDs).Find(&settings).Error, "")
	return settings, err
}

// 获取某个用户信息  不存在，则返回错误.
func (u *UserSettingGorm) Take(ctx context.Context, userID string) (setting *relation.UserSettingModel, err error) {
	setting = &relation.UserSettingModel{}
	err = utils.Wrap(u.db(ctx).Where("user_id = ?", userID).Take(&setting).Error, "")
	return setting, err
}
