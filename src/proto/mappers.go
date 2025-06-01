package proto

import (
	channelv1 "github.com/konfa-chat/hub/src/proto/konfa/channel/v1"
	chatv1 "github.com/konfa-chat/hub/src/proto/konfa/chat/v1"
	userv1 "github.com/konfa-chat/hub/src/proto/konfa/user/v1"
	"github.com/konfa-chat/hub/src/store"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapMessage(msg store.Message) *chatv1.Message {
	return &chatv1.Message{
		MessageId: msg.ID.String(),
		SenderId:  msg.SenderID.String(),
		Content:   msg.Content,
		Timestamp: timestamppb.New(msg.Timestamp),
	}
}

func mapTextChannelToChannel(c store.TextChannel) *channelv1.Channel {
	return &channelv1.Channel{
		Channel: &channelv1.Channel_TextChannel{
			TextChannel: mapTextChannel(c),
		},
	}
}

func mapVoiceChannelToChannel(c store.VoiceChannel) *channelv1.Channel {
	return &channelv1.Channel{
		Channel: &channelv1.Channel_VoiceChannel{
			VoiceChannel: mapVoiceChannel(c),
		},
	}
}

func mapTextChannel(c store.TextChannel) *channelv1.TextChannel {
	return &channelv1.TextChannel{
		ServerId:  c.ServerID.String(),
		ChannelId: c.ID.String(),
		Name:      c.Name,
	}
}

func mapVoiceChannel(c store.VoiceChannel) *channelv1.VoiceChannel {
	return &channelv1.VoiceChannel{
		ServerId:     c.ServerID.String(),
		ChannelId:    c.ID.String(),
		Name:         c.Name,
		VoiceRelayId: []string{c.RelayID},
	}
}

func mapUser(c store.User) *userv1.User {
	return &userv1.User{
		Id:       c.ID.String(),
		Username: c.Username,
	}
}
