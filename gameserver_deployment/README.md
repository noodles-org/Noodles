# Noodle Games
The Noodles dedicated game servers, run on our TrueNAS cluster.

### Servers
1. Soba Discord bot to allow fetching our external IP via `/get-server-ip`
2. Enshrouded dedicated server at `<EXTERNAL_IP>:15637`
3. Satisfactory dedicated server at `<EXTERNAL_IP>:7777`
4. Valheim dedicated server at `<EXTERNAL_IP>:2457`

### New Developer Setup
1. From the `infra/` directory, run:
   ```sh
   make setup-dev
   ```
   Or, to specify a custom kubeconfig path (default is `~/.kube/gameserver_config`):
   ```sh
   make setup-dev KUBECONFIG=~/.kube/my_custom_config
   ```
   This will:
   * Install required tools (AWS CLI, sops, kubectl, kubelogin) via Homebrew
   * Verify your AWS authentication (if not yet configured, the script will prompt you to run `aws configure` or `aws sso login` — your IAM user will be preconfigured by an admin)
   * Decrypt all sops secrets
   * Generate a kubeconfig with the OIDC context (authenticated via Dex on the foundry cluster)
2. Export the kubeconfig:
   ```sh
   export KUBECONFIG=~/.kube/gameserver_config
   ```
   Add this to your shell profile (`.zshrc` / `.bashrc`) to persist it.

### Initial Cluster Setup (Admin Only)
1. Set up the TrueNAS VM.
2. Use the supplied VNC shell to enable SSH.
3. Install K3s on the VM and grab the config file for your local machine.
   * Swap the IP to match the VM.
4. Sops decrypt the secrets via `make sops-decrypt-all`.
5. Run the ansible setup script via `make setup-cluster`, or specify a custom kubeconfig path (default is `~/.kube/gameserver_config`)
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

### Retrieve ArgoCD Manager Token
To display the `argocd-manager` ServiceAccount token (needed for `GAMESERVER_TOKEN` in the dashboard secret), pass `SHOW_TOKEN=true` during cluster setup:
```sh
make setup-cluster SHOW_TOKEN=true
```
This runs the full cluster setup and prints the token at the end. Copy it into the dashboard's `.secret.yaml` under `GAMESERVER_TOKEN`, then re-encrypt with `make sops-encrypt-all` in the foundry `infra/` directory.

### Update Soba Bot
The Soba Discord bot image can be updated by running:
```sh
make update-soba-version
```
This rebuilds the Docker image, pushes it, and restarts the deployment.
