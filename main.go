package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

type vm struct {
	Name          string
	ResoruceGroup string
	Nics          []string
	Disks         []string
	ComputeClient *armcompute.VirtualMachinesClient
}

var (
	subscriptionId string
	resourceGroup  string
	vmName         string
	err            error
	ctx            = context.Background()
)

func main() {
	var machine vm

	flag.StringVar(&vmName, "vmName", "", "Virual Machine Name")
	flag.StringVar(&resourceGroup, "resourceGroup", "", "Resource Group")
	flag.StringVar(&subscriptionId, "subscriptionId", "", "SubscriptionId")
	flag.Parse()

	if vmName == "" || resourceGroup == "" || subscriptionId == "" {
		log.Fatalf("Error getting variables ...")
	}
	getClients(&machine, subscriptionId)
	parseVM(&machine, vmName, resourceGroup)
	deleteVM(&machine)
}

func deleteVM(machine *vm) {
	fmt.Printf("Removing vm %s ...\n", machine.Name)
	poller, _ := machine.ComputeClient.BeginDelete(ctx, machine.ResoruceGroup, machine.Name, nil)
	poller.PollUntilDone(ctx, nil)
	fmt.Println("End!")
}

func getClients(machine *vm, subscriptionId string) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("Failed to obtain a credential: %v", err)
	}
	machine.ComputeClient, err = armcompute.NewVirtualMachinesClient(subscriptionId, cred, nil)
	if err != nil {
		log.Fatalf("Error getting compute client: %v", err)
	}
}

func parseVM(machine *vm, vmName string, resourceGroup string) {
	vmObj, err := machine.ComputeClient.Get(ctx, resourceGroup, vmName, nil)
	if err != nil {
		log.Fatalf("Failed to obtain a credential: %v", err)
	}
	machine.Name = *vmObj.Name
	machine.ResoruceGroup = resourceGroup
	for _, nic := range vmObj.Properties.NetworkProfile.NetworkInterfaces {
		machine.Nics = append(machine.Nics, *nic.ID)
	}
	machine.Disks = append(machine.Disks, *vmObj.Properties.StorageProfile.OSDisk.ManagedDisk.ID)
	for _, disk := range vmObj.Properties.StorageProfile.DataDisks {
		machine.Disks = append(machine.Disks, *disk.ManagedDisk.ID)
	}
}

func printJson(jsbyte []byte) {
	var prettyJson bytes.Buffer
	_ = json.Indent(&prettyJson, jsbyte, "", "    ")
	fmt.Printf("%s\n", string(prettyJson.Bytes()))
}
