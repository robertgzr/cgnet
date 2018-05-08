# -*- mode: ruby -*-
# vi: set ft=ruby :

ENV["TERM"] = "xterm-256color"
ENV["LC_ALL"] = "en_US.UTF-8"

Vagrant.configure("2") do |config|
  config.vm.box = "kaorimatz/fedora-27-x86_64"

  config.vm.synced_folder ".", "/vagrant", disabled: true
  config.vm.synced_folder ".", "/home/vagrant/go/src/github.com/kinvolk/cgnet", 
      create: true,
      owner: "vagrant",
      group: "vagrant"
  config.vm.network "forwarded_port", guest: 6443, host: 6443

  if Vagrant.has_plugin?("vagrant-vbguest")
    config.vbguest.auto_update = false
  end
  config.vm.provider :virtualbox do |vb|
      vb.check_guest_additions = true
      vb.functional_vboxsf = true
      vb.customize ["modifyvm", :id, "--memory", "8192"]
      vb.customize ["modifyvm", :id, "--cpus", "2"]
  end

  config.vm.provision "shell", inline: "dnf install -y bcc-tools bcc-devel clang llvm git go strace"

  # NOTE: chown is explicitly needed, even when synced_folder is configured
  # with correct owner/group. Maybe a vagrant issue?
  config.vm.provision "shell", inline: "mkdir -p /home/vagrant/go ; chown -R vagrant:vagrant /home/vagrant/go"

  config.vm.provision "shell", env: {"GOPATH" => "/home/vagrant/go"}, privileged: true, path: "vagrant-setup-env.sh"
end
