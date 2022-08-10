package main

import (
	"context"
	"flag"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

var (
	subscriptionId string
	resourceGroup  string
	vmName         string
	ctx            context.Context
)

func main() {
	ctx = context.Background()

	flag.StringVar(&vmName, "vmName", "", "Virual Machine Name")
	flag.StringVar(&resourceGroup, "resourceGroup", "", "Resource Group")
	flag.StringVar(&subscriptionId, "subscriptionId", "", "SubscriptionId")
	flag.Parse()

	if vmName == "" || resourceGroup == "" || subscriptionId == "" {
		log.Fatalf("Error getting variables ...")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("Failed to obtain a credential: %v", err)
	}

	client, err := armcompute.NewVirtualMachinesClient(subscriptionId, cred, nil)
	if err != nil {
		log.Fatalf("Error getting client: %v", err)
	}
	vm, err := client.Get(ctx, resourceGroup, vmName, nil)
	if err != nil {
		log.Fatalf("Error getting vm: %v", err)
	}

}

func removeNIC()
