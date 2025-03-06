package services

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/GGP1/atoll"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/samber/lo"
	"github.com/the-clothing-loop/website/server/internal/app"
	"github.com/the-clothing-loop/website/server/internal/models"
	"gorm.io/gorm"
)

// Checks if the user is registered on the matrix server if not they are registered
func ChatPatchUser(db *gorm.DB, ctx context.Context, mmTeamId string, user *models.User) error {
	var mmUser *model.User
	var err error
	// check if user exists properly
	if user.ChatUserID != nil {
		// get the chat user
		mmUser, _, err = app.ChatClient.GetUser(ctx, *user.ChatUserID, "")
		if err == nil {
			slog.Info("chat user exists", "id", mmUser.Id, "err", err)
		}
	}

	if mmUser == nil {
		slog.Info("create Chat user if does not exist")

		username := func() string {
			p, _ := atoll.NewPassword(10, []atoll.Level{atoll.Digit, atoll.Lower, atoll.Level("-_.")})
			return "u" + string(p)
		}()
		password := func() string {
			p, _ := atoll.NewPassword(16, []atoll.Level{atoll.Digit, atoll.Lower})
			return string(p)
		}()
		if user.Email == nil {
			return fmt.Errorf("Email is required")
		}
		mmUser, _, err = app.ChatClient.CreateUser(ctx, &model.User{
			Nickname: user.Name,
			Username: username,
			Password: password,
			Email:    username + "@example.com",
		})
		if err != nil {
			return err
		}

		_, _, err = app.ChatClient.AddTeamMember(ctx, mmTeamId, mmUser.Id)
		if err != nil {
			return err
		}

		// Update database
		user.ChatUserID = &mmUser.Id
		user.ChatPass = &password
		user.ChatUserName = &username
		db.Exec(`UPDATE users SET chat_user_id = ?, chat_pass = ?, chat_user_name = ? WHERE id = ?`,
			*user.ChatUserID,
			*user.ChatPass,
			*user.ChatUserName,
			user.ID)
	}

	return nil
}

func ChatCreateChannel(db *gorm.DB, ctx context.Context, chain *models.Chain, mmUserId, name, color string) (*model.Channel, error) {
	newChannel := &model.Channel{
		TeamId:      app.ChatTeamId,
		Name:        fmt.Sprintf("%dr%d%s", chain.ID, len(chain.ChatRoomIDs)+1, lo.RandomString(5, lo.LowerCaseLettersCharset)),
		DisplayName: name,
		Type:        model.ChannelTypePrivate,
		Header:      color,
	}

	newChannel, _, err := app.ChatClient.CreateChannel(ctx, newChannel)
	if err != nil {
		return nil, err
	}

	err = chatChannelAddUser(ctx, newChannel.Id, mmUserId, true)
	if err != nil {
		return nil, err
	}

	chain.ChatRoomIDs = append(chain.ChatRoomIDs, newChannel.Id)
	err = chain.SaveChannelIDs(db)
	if err != nil {
		return nil, err
	}
	return newChannel, nil
}

func ChatDeleteChannel(db *gorm.DB, ctx context.Context, chain *models.Chain, mmChannelID string) error {
	_, err := app.ChatClient.DeleteChannel(ctx, mmChannelID)
	if err != nil {
		return err
	}

	chain.ChatRoomIDs = lo.Filter(chain.ChatRoomIDs, func(roomID string, _ int) bool {
		return roomID != mmChannelID
	})
	err = chain.SaveChannelIDs(db)
	if err != nil {
		return err
	}
	return nil
}

func chatChannelAddUser(ctx context.Context, mmChannelId string, mmUserId string, setRoleAdmin bool) error {
	member, _, err := app.ChatClient.AddChannelMember(ctx, mmChannelId, mmUserId)
	if err != nil {
		return err
	}
	err = chatChannelSetMemberRole(ctx, mmChannelId, member, setRoleAdmin)
	if err != nil {
		return err
	}

	return nil
}

func chatChannelSetMemberRole(ctx context.Context, mmChannelId string, member *model.ChannelMember, setRoleAdmin bool) error {
	slog.Info("chatChannelSetMemberRole", "roles", member.Roles)
	roles := strings.Split(member.Roles, " ")

	isRolesContainsAdmin := lo.Contains(roles, model.ChannelAdminRoleId)
	shouldUpdateRoles := isRolesContainsAdmin != setRoleAdmin
	if shouldUpdateRoles {
		if setRoleAdmin {
			roles = append(roles, model.ChannelAdminRoleId)
		} else {
			roles = lo.Filter(roles, func(r string, i int) bool {
				return r != model.ChannelAdminRoleId
			})
		}
		_, err := app.ChatClient.UpdateChannelRoles(ctx, mmChannelId, member.UserId, strings.Join(roles, " "))
		if err != nil {
			return err
		}
	}
	return nil
}

func ChatJoinChannel(db *gorm.DB, ctx context.Context, chain *models.Chain, user *models.User, isChainAdmin bool, mmChannelId string) error {
	if user.ChatUserID == nil {
		return fmt.Errorf("You must be registered on our chat server before joining a room")
	}

	if len(chain.ChatRoomIDs) == 0 || !lo.Contains(chain.ChatRoomIDs, mmChannelId) {
		return fmt.Errorf("Channel does not exist in this Loop")
	}

	// Check if room already contains user
	mmChannelMembers, _, _ := app.ChatClient.GetChannelMembersByIds(ctx, mmChannelId, []string{*user.ChatUserID})
	if len(mmChannelMembers) != 0 {
		member, ok := lo.Find(mmChannelMembers, func(member model.ChannelMember) bool {
			return member.UserId == *user.ChatUserID
		})
		if ok {
			chatChannelSetMemberRole(ctx, mmChannelId, &member, isChainAdmin)
			return nil
		}
	}

	// Add user if not already added to chat room
	err := chatChannelAddUser(ctx, mmChannelId, *user.ChatUserID, isChainAdmin)
	if err != nil {
		return err
	}

	return nil
}

var reChatValidateUniqueName = regexp.MustCompile("[^a-z0-9]")

func ChatValidateUniqueName(name string) string {
	return reChatValidateUniqueName.ReplaceAllString(strings.ToLower(name), "")
}
