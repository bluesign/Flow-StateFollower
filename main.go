package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	gcloud "cloud.google.com/go/storage"
	"github.com/onflow/flow-go-sdk/client"

	"github.com/fxamacker/cbor/v2"

	"github.com/onflow/flow-go/engine/execution/computation/computer/uploader"
	"github.com/onflow/flow-go/model/flow"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func DownloadAndDumpEvents(blockID string) error {

	//google cloud bucket
	gcloudClient, _ := gcloud.NewClient(context.Background(), option.WithoutAuthentication())
	bucket := gcloudClient.Bucket("flow_public_mainnet18_execution_state")
	object := bucket.Object(fmt.Sprintf("%s.cbor", blockID))

	reader, err := object.NewReader(context.Background())
	if err != nil {
		return fmt.Errorf("could not create object reader: %w", err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("could not read execution record: %w", err)
	}
	var record uploader.BlockData
	decOptions := cbor.DecOptions{
		ExtraReturnErrors: cbor.ExtraDecErrorUnknownField,
	}
	decoder, err := decOptions.DecMode()
	if err != nil {
		panic(err)
	}

	err = decoder.Unmarshal(data, &record)
	if err != nil {
		return fmt.Errorf("could not decode execution record: %w", err)
	}

	if record.FinalStateCommitment == flow.DummyStateCommitment {
		return fmt.Errorf("execution record contains empty state commitment")
	}

	if record.Block.Header.Height == 0 {
		return fmt.Errorf("execution record contains empty block data")
	}

	for _, event := range record.Events {
		fmt.Println(event)
	}
	return nil
}

func main() {
	flow, err := client.New("access-001.mainnet18.nodes.onflow.org:9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("Failed to establish connection with the Access API")
	}

	last_block_height_processed := uint64(31735955 + 10000) //spork start + 10000 to skip empty ones

	for {
		block, _ := flow.GetLatestBlock(context.Background(), true)

		for i := last_block_height_processed + 1; i <= block.Height; i++ {
			log.Printf("Processing block height: %d\n", i)
			header, _ := flow.GetBlockByHeight(context.Background(), i)

			last_block_height_processed = i
			if len(header.CollectionGuarantees) == 0 {
				//has no collection
				continue
			}
			log.Println(header.ID.String())
			err := DownloadAndDumpEvents(header.ID.String())
			if err != nil {
				log.Fatal(err)
			}
		}

		time.Sleep(10 * time.Millisecond)
	}

}
