# sops key is ~/.config/sops/age/keys.txt by default
sops-decrypt-all:
	sops --decrypt .secrets/argocd.enc.yaml > k8s/argocd/.secret.yaml
	sops --decrypt .secrets/enshrouded.enc.yaml > k8s/enshrouded/.secret.yaml
	sops --decrypt .secrets/valheim.enc.yaml > k8s/valheim/.secret.yaml
	sops --decrypt .secrets/soba.enc.yaml > k8s/soba/.secret.yaml

sops-encrypt-all:
	sops --encrypt k8s/argocd/.secret.yaml > .secrets/argocd.enc.yaml
	sops --encrypt k8s/enshrouded/.secret.yaml > .secrets/enshrouded.enc.yaml
	sops --encrypt k8s/valheim/.secret.yaml > .secrets/valheim.enc.yaml
	sops --encrypt k8s/soba/.secret.yaml > .secrets/soba.enc.yaml
