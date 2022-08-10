#!/bin/bash
resourceGroup='GO-REMOVE'
vmName='go-remove-vm'
subscriptionId='51ac041c-510c-47de-808d-95a6c0c0a19d'

go run rmvm -vmName $vmName -resourceGroup $resourceGroup -subscriptionId $subscriptionId