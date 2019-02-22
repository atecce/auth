provider "cloudflare" {
    email = "root@atec.pub"
    token = "${file("/keybase/private/atec/etc/cloudflare/token")}"
}

resource "cloudflare_record" "subdomain" {
    name = "auth"
    domain = "atec.pub"
    type = "A"
    value = "${google_compute_address.static.address}"
    proxied = true
}

provider "google" {
    credentials = "${file("/keybase/private/atec/etc/gcp/telos.json")}"
    project = "telos-162721"
    region = "us-east1"
    zone = "us-east1-b"
}

resource "google_compute_firewall" "auth" {

    name = "auth"
    network = "default"
    target_tags = ["auth"]

    source_ranges = ["0.0.0.0/0"]
    allow = {
        protocol = "tcp"
        ports = ["22","443"]
    }
}

resource "google_compute_address" "static" {
    name = "auth"
}

resource "google_compute_instance" "default" {
 
    name = "auth"
    zone = "us-east1-b"

    network_interface = {
        network = "default"
        access_config = {
            nat_ip = "${google_compute_address.static.address}"
        }
    }
    machine_type = "g1-small"
    boot_disk = {
        initialize_params = {
            image = "centos-cloud/centos-7"
        }
    }

    tags = ["auth"]

    provisioner "remote-exec" {
        connection = {
            type = "ssh"
            user = "atec"
            private_key = "${file("~/.ssh/google_compute_engine")}"
            timeout = "120s"
        }
        inline = [
            "sudo mkdir -p /etc/auth",
            "sudo mkdir -p /etc/cloudkit",
            "sudo mkdir -p /etc/musickitjs",
            "sudo mkdir -p /etc/mapkitjs",

            "sudo yum install -y wget",

            "wget https://atec.keybase.pub/etc/sshd_config",
            "sudo mv sshd_config /etc/ssh/sshd_config",
            "sudo systemctl restart sshd.service",
        ]
    }
  
    provisioner "file" {
        connection = {
            type = "ssh"
            user = "root"
            private_key = "${file("~/.ssh/google_compute_engine")}"
            timeout = "120s"
        }
        source = "/keybase/private/atec/etc/aapl/cloudkit/eckey.pem"
        destination = "/etc/cloudkit/eckey.pem"
    }

     provisioner "file" {
        connection = {
            type = "ssh"
            user = "root"
            private_key = "${file("~/.ssh/google_compute_engine")}"
            timeout = "120s"
        }
        source = "/keybase/private/atec/etc/aapl/musickitjs/AuthKey_CUG44HA5T5.p8"
        destination = "/etc/musickitjs/AuthKey_CUG44HA5T5.p8"
    }

     provisioner "file" {
        connection = {
            type = "ssh"
            user = "root"
            private_key = "${file("~/.ssh/google_compute_engine")}"
            timeout = "120s"
        }
        source = "/keybase/private/atec/etc/aapl/mapkitjs/AuthKey_YKVC29UG5H.p8"
        destination = "/etc/mapkitjs/AuthKey_YKVC29UG5H.p8"
    }

    provisioner "file" {
        connection = {
            type = "ssh"
            user = "root"
            private_key = "${file("~/.ssh/google_compute_engine")}"
            timeout = "120s"
        }
        source = "/keybase/private/atec/etc/server.crt"
        destination = "/etc/auth/server.crt"
    }

    provisioner "file" {
        connection = {
            type = "ssh"
            user = "root"
            private_key = "${file("~/.ssh/google_compute_engine")}"
            timeout = "120s"
        }
        source = "/keybase/private/atec/etc/server.key"
        destination = "/etc/auth/server.key"
    }

    provisioner "file" {
        connection = {
            type = "ssh"
            user = "root"
            private_key = "${file("~/.ssh/google_compute_engine")}"
            timeout = "120s"
        }
        source = "/keybase/public/atec/bin/auth"
        destination = "/usr/sbin/auth"
    }

    provisioner "file" {
        connection = {
            type = "ssh"
            user = "root"
            private_key = "${file("~/.ssh/google_compute_engine")}"
            timeout = "120s"
        }
        source = "auth.service"
        destination = "/etc/systemd/system/auth.service"
    }

    provisioner "remote-exec" {
        connection = {
            type = "ssh"
            user = "root"
            private_key = "${file("~/.ssh/google_compute_engine")}"
            timeout = "120s"
        }
        inline = [
            "chmod 755 /usr/sbin/auth",
            "systemctl start auth.service",
        ]
    }

    depends_on = ["google_compute_firewall.auth"]
}
