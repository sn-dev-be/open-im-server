package permissions

import (
	"encoding/json"
)

type Permissions map[string]bool

const (
	ManageServer        = "manageServer"
	ManageGroupCategory = "ManageGroupCategory"
	ManageGroup         = "manageChannel"
	ManageRole          = "manageRole"
	ManageMember        = "manageMember"
	ManageMsg           = "manageMsg"
	ManageCommunity     = "manageCommunity"
	SendMsg             = "sendMsg"
	ShareServer         = "shareServer"
	PostTweet           = "PostTweet"
	TweetReply          = "TweetReply"
)

func NewPermissions(auths map[string]bool) Permissions {
	return auths
}

func NewBasePermissions() Permissions {
	return NewPermissions(map[string]bool{
		ManageServer:        false,
		ManageGroupCategory: false,
		ManageGroup:         false,
		ManageRole:          false,
		ManageMember:        false,
		ManageMsg:           false,
		ManageCommunity:     false,
		SendMsg:             true,
		ShareServer:         true,
		PostTweet:           true,
		TweetReply:          true,
	})
}

func NewDefaultEveryonePermissions() Permissions {
	base := NewBasePermissions()
	return base
}

func NewDefaultAdminPermissions() Permissions {
	base := NewBasePermissions()
	base.AddOrUpdatePermission(ManageServer, true)
	base.AddOrUpdatePermission(ManageGroupCategory, true)
	base.AddOrUpdatePermission(ManageGroup, true)
	base.AddOrUpdatePermission(ManageRole, true)
	base.AddOrUpdatePermission(ManageMember, true)
	base.AddOrUpdatePermission(ManageCommunity, true)
	base.AddOrUpdatePermission(ManageMsg, true)
	return base
}

func (p Permissions) AddOrUpdatePermission(key string, value bool) {
	p[key] = value
}

func (p Permissions) HasPermission(key string) bool {
	value, ok := p[key]
	if !ok {
		return false
	}
	return value
}

func (p Permissions) CanManageServer() bool {
	return p.HasPermission(ManageServer)
}

func (p Permissions) CanManageGroupCategory() bool {
	return p.HasPermission(ManageGroupCategory)
}

func (p Permissions) CanManageGroup() bool {
	return p.HasPermission(ManageGroup)
}

func (p Permissions) CanManageRole() bool {
	return p.HasPermission(ManageRole)
}

func (p Permissions) CanManageMember() bool {
	return p.HasPermission(ManageMember)
}

func (p Permissions) CanManageMsg() bool {
	return p.HasPermission(ManageMsg)
}

func (p Permissions) CanManageCommunity() bool {
	return p.HasPermission(ManageCommunity)
}

func (p Permissions) CanSendMsg() bool {
	return p.HasPermission(SendMsg)
}

func (p Permissions) CanShareServer() bool {
	return p.HasPermission(ShareServer)
}

func (p Permissions) CanPostTweet() bool {
	return p.HasPermission(PostTweet)
}

func (p Permissions) CanTweetReply() bool {
	return p.HasPermission(TweetReply)
}

func (p Permissions) ToJSON() (string, error) {
	bytes, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func PermissionsFromJSON(jsonStr string) (Permissions, error) {
	var p Permissions
	err := json.Unmarshal([]byte(jsonStr), &p)
	if err != nil {
		return nil, err
	}
	return p, nil
}
