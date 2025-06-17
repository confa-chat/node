package proto

import (
	"fmt"

	channelv1 "github.com/confa-chat/node/src/proto/confa/channel/v1"
	chatv1 "github.com/confa-chat/node/src/proto/confa/chat/v1"
	userv1 "github.com/confa-chat/node/src/proto/confa/user/v1"
	"github.com/confa-chat/node/src/store"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapMessage(msg store.Message) *chatv1.Message {
	protoMsg := &chatv1.Message{
		MessageId: msg.ID.String(),
		SenderId:  msg.SenderID.String(),
		Content:   msg.Content,
		Timestamp: timestamppb.New(msg.Timestamp),
	}

	// Map attachments if any exist
	if len(msg.Attachments) > 0 {
		protoMsg.Attachments = make([]*chatv1.Attachment, len(msg.Attachments))

		for i, attachment := range msg.Attachments {
			protoMsg.Attachments[i] = &chatv1.Attachment{
				AttachmentId: attachment.AttachmentID.String(),
				Name:         attachment.Name,
				Url:          fmt.Sprintf("/attachments/%s/%s", attachment.AttachmentID.String(), attachment.Name),
			}
		}
	}

	return protoMsg
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
