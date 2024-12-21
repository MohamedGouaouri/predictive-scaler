# # Cloud-init configuration
# data "template_file" "cloudinit" {
#   template = <<EOF
# #cloud-config
# ssh_authorized_keys:
#   - ${file(var.ssh_public_key)}
# EOF
# }

# resource "libvirt_pool" "ubuntu" {
#   name = "ubuntu"
#   type = "dir"
#   path = "/tmp/terraform-provider-libvirt-pool-ubuntu"
# }

resource "libvirt_volume" "ubuntu_image" {
  name   = "ubuntu-image"
  pool   = "default"
  format = "qcow2"
  source = var.ubuntu_img_url
}

# resource "libvirt_cloudinit_disk" "commoninit" {
#   name           = "commoninit.iso"
#   pool           = libvirt_pool.ubuntu.name
#   user_data      = data.template_file.cloudinit.rendered
# }

resource "libvirt_volume" "os_disk" {
  count   = var.num_vms
  name    = "os-disk-${count.index}"
  pool    =  "default"
  format  = "qcow2"
  size    = var.vm_disk
}

resource "libvirt_domain" "ubuntu_vm" {
  name   = "${var.vm_name}-${count.index}"
  memory = var.vm_memory
  vcpu   = var.vm_cpu
  count  = var.num_vms

  # Use the Ubuntu image as the boot disk
  disk {
    volume_id = libvirt_volume.ubuntu_image.id
  }

  disk {
    volume_id = libvirt_volume.os_disk[count.index].id
  }

#   cloudinit = libvirt_cloudinit_disk.commoninit.id

  network_interface {
    network_name = "default"
  }

  console {
    type        = "pty"
    target_port = "0"
  }

  autostart = true
}
