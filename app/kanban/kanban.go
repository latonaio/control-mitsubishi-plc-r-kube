package kanban

import (
	"bitbucket.org/latonaio/aion-core/pkg/go-client/msclient"
	"context"
)

func WriteKanban(ctx context.Context,data map[string]interface{}) error {
	client,err := msclient.NewKanbanClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	metadata := msclient.SetMetadata(data)
	req,err := msclient.NewOutputData(metadata)
	if err != nil {
		return err
	}
	err = client.OutputKanban(req)
	if err != nil {
		return err
	}
	return nil
}
