Vagrant.configure(2) do |config|
  config.vm.box = 'bento/ubuntu-18.04'

  config.vm.synced_folder '.', '/vagrant', type: 'rsync',
                                           rsync__exclude: '.git/'

  config.vm.provision :ansible do |ansible|
    ansible.playbook = 'xen.yml'
  end

  if Vagrant.has_plugin? 'vagrant-vbguest'
    # Don't try and install guest drivers
    config.vbguest.auto_update = false
    config.vbguest.no_remote = true
  end

  config.vm.provider 'virtualbox' do |v|
    v.linked_clone = true if Gem::Version.new(Vagrant::VERSION) >=
                             Gem::Version.new('1.8.0')

    v.customize ['modifyvm', :id, '--memory', '2048']
    v.customize ['modifyvm', :id, '--cpus', '2']
    v.customize ['modifyvm', :id, '--chipset', 'piix3']
    v.customize ['modifyvm', :id, '--ioapic', 'on']
    v.customize ['modifyvm', :id, '--rtcuseutc', 'on']
    v.customize ['modifyvm', :id, '--cpuexecutioncap', '100']
    v.customize ['modifyvm', :id, '--pae', 'off']
    v.customize ['modifyvm', :id, '--nestedpaging', 'on']
    v.customize ['modifyvm', :id, '--vram', '16']
    v.customize ['modifyvm', :id, '--accelerate3d', 'off']
  end
end
