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
)

const GroupTreasuryTableName = "group_treasuries"

type GroupTreasuryModel struct {
	GroupID              string `gorm:"column:group_id;primary_key;size:64"`
	TreasuryID           string `gorm:"column:treasury_id"`
	Icon                 string `gorm:"column:icon"`
	Name                 string `gorm:"column:name"`
	WalletType           int32  `gorm:"column:wallet_type"`
	Symbol               string `gorm:"column:symbol"`
	ContractAddress      string `gorm:"column:contract_address"`
	AdministratorAddress string `gorm:"column:administrator_address"`
	Ex                   string `gorm:"column:ex;size:1024"`
}

func (GroupTreasuryModel) TableName() string {
	return GroupTreasuryTableName
}

type GroupTreasuryModelInterface interface {
	Create(ctx context.Context, treasuries []*GroupTreasuryModel) (err error)
	NewTx(tx any) GroupTreasuryModelInterface
	Delete(ctx context.Context, groupID string) (err error)
	UpdateByMap(ctx context.Context, groupID string, args map[string]interface{}) (err error)
	Update(ctx context.Context, treasuries []*GroupTreasuryModel) (err error)
	Find(ctx context.Context, groupID string) (treasury *GroupTreasuryModel, err error)
	Take(ctx context.Context, groupID string) (treasury *GroupTreasuryModel, err error)
}
