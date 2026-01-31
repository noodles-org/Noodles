# sops key is ~/.config/sops/age/keys.txt by default
sops-decrypt-all:
	sops --decrypt playbooks/.secrets/restic-ansible-vault.enc.yaml > playbooks/.secrets/restic-ansible-vault.yaml && \
	sops --decrypt .secrets/argocd.enc.yaml > .secrets/argocd.yaml && \
	sops --decrypt .secrets/cloudflare.enc.yaml > .secrets/cloudflare.yaml && \
	sops --decrypt .secrets/github.enc.yaml > .secrets/github.yaml && \
	sops --decrypt .secrets/grafana.enc.yaml > .secrets/grafana.yaml && \
	sops --decrypt .secrets/restic.enc.yaml > .secrets/restic.yaml && \
	sops --decrypt .secrets/escrow.enc > .secrets/escrow

sops-encrypt-all:
	sops --encrypt playbooks/.secrets/restic-ansible-vault.yaml > playbooks/.secrets/restic-ansible-vault.enc.yaml && \
	sops --encrypt .secrets/argocd.yaml > .secrets/argocd.enc.yaml && \
	sops --encrypt .secrets/cloudflare.yaml > .secrets/cloudflare.enc.yaml && \
	sops --encrypt .secrets/github.yaml > .secrets/github.enc.yaml && \
	sops --encrypt .secrets/grafana.yaml > .secrets/grafana.enc.yaml && \
	sops --encrypt .secrets/restic.yaml > .secrets/restic.enc.yaml && \
	sops --encrypt .secrets/escrow > .secrets/escrow.enc

switch-context-foundry:
	export KUBECONFIG=~/.kube/foundry_config

setup-cluster:
	ansible-playbook -i inventory.ini playbooks/setup_cluster.yaml --extra-vars "config=~/.kube/foundry_config"

manual-backup:
	kubectl apply -f deployments/foundry/jobs/manual/manual-backup.yaml
manual-restore-latest:
	kubectl apply -f deployments/foundry/jobs/manual/restore-from-latest.yaml

mount-archive:
	ansible-playbook -i inventory.ini playbooks/mount_archive.yaml
unmount-archive:
	umount ~/mnt && rm -r ~/mnt
manual-download-archive:
	ansible-playbook -i inventory.ini playbooks/download_archive_nas.yaml --extra-vars "destination=~/Downloads"