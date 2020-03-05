package client

import (
	"context"
	. "github.com/xhebox/chrootd/api/containerpool/protobuf"
	"time"
)

func FindContainer(client ContainerPoolClient, name string, isCreate bool) (*Reply, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.FindContainer(ctx, &Query{
		Name:     name,
		IsCreate: isCreate,
	})

}

func DeleteContainerById(client ContainerPoolClient, id string) (*Reply, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.SetContainer(ctx, &SetRequest{
		State: StateType_Delete,
		Body: &SetRequest_Delete{&DeleteContainer{
			Id: id,
		}},
	})
}
