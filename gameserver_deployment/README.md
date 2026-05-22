# Noodle Games
The Noodles dedicated game servers, run on our TrueNAS cluster.

### Set Up
Sops decrypt the secrets via `make sops-decrypt-all`.
Set kube configs via `export KUBECONFIG=~/.kube/gameserver_config:~/.kube/foundry_config`

### Servers
1. Soba Discord bot to allow fetching our external IP via `/get-server-ip`
2. Enshrouded dedicated server at `<EXTERNAL_IP>:15637`
3. Satisfactory dedicated server at `<EXTERNAL_IP>:7777`
4. Valheim dedicated server at `<EXTERNAL_IP>:2457`
