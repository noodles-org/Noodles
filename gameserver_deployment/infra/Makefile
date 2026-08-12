KUBECONFIG ?= ~/.kube/gameserver_config
SHOW_TOKEN ?= false

install-collections:
	ansible-galaxy collection install -r requirements.yml

setup-cluster: install-collections
	ansible-playbook -i inventory.ini playbooks/setup_cluster.yaml --extra-vars "config=$(KUBECONFIG) show_token=$(SHOW_TOKEN)" -K

setup-dev: install-collections
	ansible-playbook -i inventory.ini playbooks/setup_dev.yaml --extra-vars "kubeconfig=$(KUBECONFIG)"

# sops key is ~/.config/sops/age/keys.txt by default
# export AWS_PROFILE for KMS encryption/decryption.
sops-decrypt-all:
	sops --decrypt .secrets/argocd.enc.yaml > k8s/argocd/.secret.yaml
	sops --decrypt .secrets/auth.enc.yaml > k8s/auth/.secret.yaml
	sops --decrypt .secrets/enshrouded.enc.yaml > k8s/enshrouded/.secret.yaml
	sops --decrypt .secrets/valheim.enc.yaml > k8s/valheim/.secret.yaml
	sops --decrypt .secrets/soba.enc.yaml > k8s/soba/.secret.yaml

sops-encrypt-all:
	sops --encrypt k8s/argocd/.secret.yaml > .secrets/argocd.enc.yaml
	sops --encrypt k8s/auth/.secret.yaml > .secrets/auth.enc.yaml
	sops --encrypt k8s/enshrouded/.secret.yaml > .secrets/enshrouded.enc.yaml
	sops --encrypt k8s/valheim/.secret.yaml > .secrets/valheim.enc.yaml
	sops --encrypt k8s/soba/.secret.yaml > .secrets/soba.enc.yaml

update-soba-version:
	docker buildx build --platform linux/amd64 -t mephalrith/soba-discord-bot k8s/soba/soba-bot/.
	docker push mephalrith/soba-discord-bot:latest
	kubectl rollout restart -n soba soba