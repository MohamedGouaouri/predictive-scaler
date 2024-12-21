## SSH public key variable
variable "ssh_public_key" {
  description = "Public SSH key for accessing the VM"
  type        = string
  default     = "~/.ssh/id_rsa_ubuntu.pub" 
}

variable "vm_name" {
  default = "ubuntu_vm"
}

variable "ubuntu_img_url" {
  description = "Ubuntu image"
#   default     = "https://cloud-images.ubuntu.com/releases/focal/release/ubuntu-20.04-server-cloudimg-amd64.img"
    default = "./iso/ubuntu-18.04.6-live-server-amd64.iso"
}

variable "num_vms" {
  default = 1
}

variable "vm_memory" {
  default = 4096  # 4 GB of RAM
}

variable "vm_cpu" {
  default = 2     # 2 vCPUs
}

variable "vm_disk" {
  default = 8 * 1024 * 1024 * 1024
}
