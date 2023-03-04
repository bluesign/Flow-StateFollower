package main

import (
	"context"
	"fmt"
	"github.com/bluesign/stateFollower/websocket"
	"io"
	"log"
	"net/http"
	"time"

	gcloud "cloud.google.com/go/storage"
	"github.com/fxamacker/cbor/v2"
	client "github.com/onflow/flow-go-sdk/access/grpc"
	"github.com/onflow/flow-go/engine/execution/computation/computer/uploader"
	"github.com/onflow/flow-go/model/flow"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func DownloadAndDumpEvents(bucket *gcloud.BucketHandle, blockID string, height uint64) error {

	//google cloud bucket
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
		fmt.Printf("BlockID\t\t : %s - %d \nTransactionID\t : %s\nTransactionIndex : %d\nEventIndex\t : %d\n",
			blockID, height, event.TransactionID, event.TransactionIndex, event.EventIndex)
		fmt.Printf("Event Type\t: %s\n", event.Type)
		//	fmt.Printf("Event Payload\t: %s\n\n", string(event.Payload))
		server.Publish(string(event.Type), event.Payload)

	}
	return nil
}

var server = &websocket.Server{Subscriptions: make(websocket.Subscription)}

func main() {

	go func() {
		log.SetFlags(log.Lshortfile)

		handler := websocket.NewHandler(server)
		http.HandleFunc("/socket", handler.HandleWS)
		if err := http.ListenAndServeTLS(":443", "fullchain.pem", "privkey.pem", nils); err != nil {
			panic(err)
		}
	}()

	flow, errGrpc := client.NewBaseClient("access-001.mainnet22.nodes.onflow.org:9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if errGrpc != nil {
		log.Fatal("Failed to establish connection with the Access API")
	}

	latestBlock, _ := flow.GetLatestBlock(context.Background(), true)

	lastBlockHeightProcessed := uint64(latestBlock.Height) //spork start + 10000 to skip empty ones
	gcloudClient, _ := gcloud.NewClient(context.Background(), option.WithoutAuthentication())
	bucket := gcloudClient.Bucket("flow_public_mainnet22_execution_state")

	for {
		block, _ := flow.GetLatestBlock(context.Background(), true)

		for i := lastBlockHeightProcessed + 1; i <= block.Height; i++ {
			log.Printf("Processing block height: %d\n", i)
			header, _ := flow.GetBlockByHeight(context.Background(), i)

			lastBlockHeightProcessed = i
			if len(header.CollectionGuarantees) == 0 {
				//has no collection
				continue
			}
			log.Println(header.ID.String())
			err := DownloadAndDumpEvents(bucket, header.ID.String(), i)
			if err != nil {
				log.Fatal(err)
			}
		}

		time.Sleep(10 * time.Millisecond)
	}

}
