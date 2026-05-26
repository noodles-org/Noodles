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


### Directly Access the Foundry Files
1. `k get pod -n foundry` to find the pod
2. `k exec {POD NAME} -it -n foundry -- sh` to exec into it with a shell terminal
3. `cd /foundrydata` to directly access the files.
