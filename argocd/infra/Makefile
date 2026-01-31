# sops key is ~/.config/sops/age/keys.txt by default
sops-encrypt-all:
	sops --encrypt k8s/.secret.yaml > .secrets/argocd.enc.yaml

sops-decrypt-all:
	sops --decrypt .secrets/argocd.enc.yaml > k8s/.secret.yaml

# Argo lives in the foundry cluster
switch-context-argocd:
	export KUBECONFIG=~/.kube/foundry_config