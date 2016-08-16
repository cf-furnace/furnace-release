# furnace-release

bosh release for the furnace work

## Bosh-lite changes
- Add the following to the end of your Vagrantfile
```
  config.vm.provision "shell", inline: "modprobe br-netfilter"
```
