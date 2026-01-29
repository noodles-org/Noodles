# Noodles Quest
The Noodles Foundry server, run on our TrueNAS cluster.

### Set Up
1. Set up the TrueNAS VM.
2. Use the supplied VNC shell to enable SSH.
3. Install K3s on the VM and grab the config file for your local machine.
   * Swap the IP to match the VM.
4. Sops decrypt the secrets via `make sops-decrypt-all`.
5. Run the ansible setup script via `make setup-cluster`.
6. Update your hosts file for the `noodles.local` VM IP.

### Directly Access the Foundry Files
1. `k get pod -n foundry` to find the pod
2. `k exec deployment-{POD NAME} -it -n foundry -- sh` to exec into it with a shell terminal
3. `cd /foundrydata` to directly access the files.
