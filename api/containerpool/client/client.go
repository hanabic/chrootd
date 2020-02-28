package client

import (
	"context"
	. "github.com/xhebox/chrootd/api/containerpool/protobuf"
	"time"
)

func FindContainer(client ContainerPoolClient, name string, isCreate bool) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reply, err := client.FindContainer(ctx, &Query{
		Name:     name,
		IsCreate: isCreate,
	})

	if err != nil {
		return "", err
	}

	return reply.Message, nil
}

func DeleteContainerById(client ContainerPoolClient, id string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reply, err := client.SetContainer(ctx, &SetRequest{
		State: StateType_Delete,
		Body: &SetRequest_Delete{&DeleteContainer{
			Id: id,
		}},
	})
	if err != nil {
		return "", err
	}

	return reply.Message, nil
}
