resourceGroup='go-group'
vmName='go-vm'
subscriptionId='51ac041c-510c-47de-808d-95a6c0c0a19d'

list:
	az resource list -g $(resourceGroup) -o table

run:
	go run rmvm -vmName $(vmName) -resourceGroup $(resourceGroup) -subscriptionId $(subscriptionId)