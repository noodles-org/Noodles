# sops key is ~/.config/sops/age/keys.txt by default
sops-decrypt-all:
	sops --decrypt .secrets/argocd.enc.yaml > argocd/.secret.yaml && \
	sops --decrypt .secrets/enshrouded.enc.yaml > enshrouded/.secret.yaml && \
	sops --decrypt .secrets/valheim.enc.yaml > valheim/.secret.yaml && \
	sops --decrypt .secrets/soba.enc.yaml > soba/.secret.yaml

sops-encrypt-all:
	sops --encrypt argocd/.secret.yaml > .secrets/argocd.enc.yaml && \
	sops --encrypt enshrouded/.secret.yaml > .secrets/enshrouded.enc.yaml && \
	sops --encrypt valheim/.secret.yaml > .secrets/valheim.enc.yaml && \
	sops --encrypt soba/.secret.yaml > .secrets/soba.enc.yaml

switch-context-gameserver:
	export KUBECONFIG=~/.kube/gameserver_config