variable "prefix" {
  default = "go"
}

resource "azurerm_resource_group" "example" {
  name     = "${var.prefix}-group"
  location = "eastus"
}

resource "azurerm_virtual_network" "example" {
  name                = "${var.prefix}-network"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
}

resource "azurerm_subnet" "internal" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.example.name
  virtual_network_name = azurerm_virtual_network.example.name
  address_prefixes     = ["10.0.2.0/24"]
}

resource "azurerm_network_interface" "main" {
  name                = "${var.prefix}-nic"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name

  ip_configuration {
    primary = true
    name                          = "testconfiguration1"
    subnet_id                     = azurerm_subnet.internal.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id = azurerm_public_ip.pubIp.id
  }
}

resource "azurerm_network_interface" "main2" {
  name                = "${var.prefix}-nic2"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name

  ip_configuration {
    primary = false
    name                          = "testconfiguration2"
    subnet_id                     = azurerm_subnet.internal.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id = azurerm_public_ip.pubip2.id
  }
}

resource "azurerm_public_ip" "pubIp" {
  resource_group_name = azurerm_resource_group.example.name
  location = azurerm_resource_group.example.location
  name = "${var.prefix}-pubIp"
  allocation_method = "Dynamic"
}

resource "azurerm_storage_account" "sta" {
  name = "sad898s67s8mj"
  account_replication_type = "LRS"
  location = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  account_tier = "Standard"
}

resource "azurerm_virtual_machine" "main" {
  name                  = "${var.prefix}-vm"
  location              = azurerm_resource_group.example.location
  resource_group_name   = azurerm_resource_group.example.name
  network_interface_ids = [azurerm_network_interface.main.id, azurerm_network_interface.main2.id]
  primary_network_interface_id = azurerm_network_interface.main.id
  vm_size               = "Standard_DS1_v2"

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "18.04-LTS"
    version   = "latest"
  }
  storage_os_disk {
    name              = "myosdisk1"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    computer_name  = "ubut"
    admin_username = "tairan"
    admin_password = "@a7&&s9o%%d12*4!a"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }

  tags = {
    environment = "sandbox"
  }
}

resource "azurerm_public_ip" "pubip2" {
  name                = "asdhfy7672"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  allocation_method   = "Dynamic"

  tags = {
    environment = "staging"
  }
}

output "storage_string" {
  value = azurerm_storage_account.sta.primary_connection_string
  sensitive = true
}
