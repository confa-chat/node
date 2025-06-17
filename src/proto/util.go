package proto

import (
	"github.com/confa-chat/node/pkg/uuid"
	chatv1 "github.com/confa-chat/node/src/proto/confa/chat/v1"
)

func apply[I any, O any](input []I, f func(I) O) []O {
	res := make([]O, 0, len(input))
	for _, v := range input {
		res = append(res, f(v))
	}
	return res
}

type channelRef struct {
	ServerID  uuid.UUID
	ChannelID uuid.UUID
}

func parseChannelRef(ref *chatv1.TextChannelRef) (channelRef, error) {
	serverID, err := uuid.FromString(ref.ServerId)
	if err != nil {
		return channelRef{}, err
	}
	channelID, err := uuid.FromString(ref.ChannelId)
	if err != nil {
		return channelRef{}, err
	}
	return channelRef{
		ServerID:  serverID,
		ChannelID: channelID,
	}, nil
}
