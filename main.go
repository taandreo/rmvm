package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/streaming"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

type vm struct {
	Name               string
	ResoruceGroup      string
	subscriptionId     string
	Nics           	   []string
	Disks          	   []string
	PubIps		   	   []string
	ComputeClient      *armcompute.VirtualMachinesClient
	NetInterfaceClient *armnetwork.InterfacesClient
	DiskClient         *armcompute.DisksClient
	PubIpClient        *armnetwork.PublicIPAddressesClient
}

type nic struct {
	Name string
	Configs []nicConfig
}

type nicConfig struct {
	Name string
	Primary bool
	PrivateIp string
}

var (
	subscriptionId string
	resourceGroup  string
	vmName         string
	backup	       bool
	err            error
	storageUri     string
	ctx            = context.Background()
)

func main() {
	var machine vm

	flag.StringVar(&vmName, "vmName", "", "Virual Machine Name")
	flag.StringVar(&resourceGroup, "resourceGroup", "", "Resource Group")
	flag.StringVar(&subscriptionId, "subscriptionId", "", "SubscriptionId")
	flag.BoolVar(&backup, "JsonBackup", false, "Enables json object for the informed vm")
	flag.Parse()

	if vmName == "" || resourceGroup == "" || subscriptionId == "" {
		log.Fatalf("Error getting variables ...")
	}

	getClients(&machine, subscriptionId)
	parseVM(&machine, vmName, resourceGroup)
	if backup {
		connectionString := os.Getenv("ARM_CONNECTION_STRING")
		if connectionString != "" { 
			OBJBackup(&machine, connectionString)
		} else {
			usage := "To use de backup option, is necessary to inform the connection in the variables"
			log.Fatalf("Error: ARM_CONNECTION_STRING variables not set\n%s", usage)
		}
	}
	
	// startBackup(&machine, vaultName)
	deleteVM(&machine)
}

func deleteVM(machine *vm) {
	fmt.Printf("Removing vm %s ... ", machine.Name)
	poller, _ := machine.ComputeClient.BeginDelete(ctx, machine.ResoruceGroup, machine.Name, nil)
	poller.PollUntilDone(ctx, nil)
	fmt.Println("ok")
	for _, nicId := range machine.Nics {
		rg := strings.Split(nicId, "/")[4]
		name := strings.Split(nicId, "/")[8]
		fmt.Printf("Deleting nic %s ... ", name)
		poller, _ := machine.NetInterfaceClient.BeginDelete(ctx, rg, name, nil)
		poller.PollUntilDone(ctx, nil)
		fmt.Println("ok")
	}
	for _, diskId := range machine.Disks {
		rg := strings.Split(diskId, "/")[4]
		name := strings.Split(diskId, "/")[8]
		fmt.Printf("Deleting disk %s ... ", name)
		poller, _ := machine.DiskClient.BeginDelete(ctx, rg, name, nil)
		poller.PollUntilDone(ctx, nil)
		fmt.Println("ok")
	}
	for _, pubIpId := range machine.PubIps {
		rg := strings.Split(pubIpId, "/")[4]
		name := strings.Split(pubIpId, "/")[8]
		fmt.Printf("Deleting publicIp %s ... ", name)
		poller, _ := machine.PubIpClient.BeginDelete(ctx, rg, name, nil)
		poller.PollUntilDone(ctx, nil)
		fmt.Println("ok")
	}
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
	machine.NetInterfaceClient, err = armnetwork.NewInterfacesClient(subscriptionId, cred, nil)
	if err != nil {
		log.Fatalf("Error getting network client: %v", err)
	}
	machine.DiskClient, err = armcompute.NewDisksClient(subscriptionId, cred, nil)
	if err != nil {
		log.Fatalf("Error getting disk client: %v", err)
	}
	machine.PubIpClient, err = armnetwork.NewPublicIPAddressesClient(subscriptionId, cred, nil)
	if err != nil {
		log.Fatalf("Error getting pubip client: %v", err)
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
	getPubIp(machine)
}

func getPubIpFromCOnfig(machine *vm, name string, rg string) []string {
	var pl []string

	nic, _ := machine.NetInterfaceClient.Get(ctx, rg, name, nil)
	for _, c := range nic.Properties.IPConfigurations {
		if (c.Properties.PublicIPAddress != nil){
			pip := *c.Properties.PublicIPAddress.ID
			pl = append(pl, pip)
		}
	}
	return pl
}

func getPubIp(machine *vm){
	var pl []string

	for _, nicId := range  machine.Nics {
		rg := strings.Split(nicId, "/")[4]
		name := strings.Split(nicId, "/")[8]
		pl = append(pl, getPubIpFromCOnfig(machine, name, rg)...)
	}
	machine.PubIps = pl
}


func getIpsConfigs(machine *vm) []nic {
	var nicList []nic

	for _, nicId := range machine.Nics {
		ncard := nic{}
		nicRG := strings.Split(nicId, "/")[4]
		ncard.Name = strings.Split(nicId, "/")[8]
		nicResp, _ := machine.NetInterfaceClient.Get(ctx, nicRG, ncard.Name, nil)		
		for _, config := range nicResp.Properties.IPConfigurations {
			ncard.Configs = append(ncard.Configs, nicConfig{*config.Name, *config.Properties.Primary, *config.Properties.PrivateIPAddress})
		}
		nicList = append(nicList, ncard)
	}
	return nicList
}

func OBJBackup(machine *vm, connectionString string){
	var m map[string]interface{}

	t := time.Now().Format("20060102150405")
	vmObj, _ := machine.ComputeClient.Get(ctx, resourceGroup, vmName, nil)
	jsonBytes, _ := vmObj.MarshalJSON()
	json.Unmarshal(jsonBytes, &m)
	m["ipConfig"] = getIpsConfigs(machine)
	data, err := json.Marshal(m)
	container := "backupvmjson"
	blobName := machine.Name + "_" + t + ".json"
	coninerClient, err := azblob.NewContainerClientFromConnectionString(connectionString, container, nil)
	handleError(err)
	coninerClient.Create(ctx, nil)
	blobClient, _ := coninerClient.NewBlockBlobClient(blobName)
	blobClient.Upload(ctx, streaming.NopCloser(strings.NewReader(printJson(data))), nil)
}

func handleError(err error){
	if err != nil {
		log.Fatalln(err)
	}
}

func printJson(jsbyte []byte) string {
	var prettyJson bytes.Buffer
	_ = json.Indent(&prettyJson, jsbyte, "", "    ")
	return string(prettyJson.Bytes())
}
