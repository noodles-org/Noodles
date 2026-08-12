# Noodles Quest
The Noodles Foundry server, run on our TrueNAS cluster.

We use the Node.js version of FoundryVTT. The Foundry version can be updated by rebuilding the Dockerfile with the following command:
`make update-foundry-version URL=<timed_url>`

### New Developer Setup
1. From the `infra/` directory, run:
   ```sh
   make setup-dev
   ```
   Or, to specify a custom kubeconfig path (default is `~/.kube/foundry_config`):
   ```sh
   make setup-dev KUBECONFIG=~/.kube/my_custom_config
   ```
   This will:
   * Install required tools (AWS CLI, sops, kubectl, kubelogin) via Homebrew
   * Verify your AWS authentication (if not yet configured, the script will prompt you to run `aws configure` or `aws sso login` — your IAM user will be preconfigured by an admin)
   * Decrypt all sops secrets
   * Generate a kubeconfig with the OIDC context
2. Export the kubeconfig:
   ```sh
   export KUBECONFIG=~/.kube/foundry_config
   ```
   Add this to your shell profile (`.zshrc` / `.bashrc`) to persist it.

### Initial Cluster Setup (Admin Only)
1. Set up the TrueNAS VM.
2. Use the supplied VNC shell to enable SSH.
3. Install K3s on the VM and grab the config file for your local machine.
   * Swap the IP to match the VM.
4. Update your host file for the `noodles.local` VM IP.
5. Sops decrypt the secrets via `make sops-decrypt-all`.
6. Run the ansible setup script via `make setup-cluster`, or specify a custom kubeconfig path (default is `~/.kube/foundry_config`)
   ```sh
   make setup-cluster KUBECONFIG=~/.kube/my_custom_config
   ```

### Ansible Collections
The playbooks depend on the `community.general` and `ansible.posix` collections, pinned in `infra/requirements.yml`. Both `make setup-cluster` and `make setup-dev` install them first, so there is nothing extra to run. To install them on their own:
```sh
make install-collections
```

### Node Disk Sizing
The VM's root disk backs every `local-path` PersistentVolume, so if it fills up the kubelet reports `NodeHasDiskPressure`, taints the node and evicts pods.

To grow the VM beyond its current disk:
1. In TrueNAS, edit the VM's zvol and raise its size. Keep it under the pool's free space unless it is sparse — and note the UI only allows growing a zvol, never shrinking it.
2. On the node, pick up the new capacity:
   ```sh
   echo 1 | sudo tee /sys/block/sda/device/rescan
   sudo partprobe /dev/sda
   lsblk   # sda should report the new size
   ```
   If `lsblk` still shows the old size, the guest never saw the change — reboot the VM, and confirm the zvol edit was actually saved.
3. Re-run `make setup-cluster` to extend the filesystem.

### Recovering from Disk Pressure
If the node is already under disk pressure:
```sh
df -h /
sudo du -xh --max-depth=2 /var/lib/rancher/k3s/storage | sort -rh | head -20
sudo journalctl --vacuum-size=200M
sudo k3s crictl rmi --prune
```
`local-path` does not enforce PVC sizes, so an oversized claim is usually the cause. Cross-check `kubectl get pv` before deleting any directory under `/var/lib/rancher/k3s/storage`.

### Directly Access the Foundry Files
1. `k get pod -n foundry` to find the pod
2. `k exec {POD NAME} -it -n foundry -- sh` to exec into it with a shell terminal
3. `cd /foundrydata` to directly access the files.
