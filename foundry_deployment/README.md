# Noodles Quest
The Noodles Foundry server, run on our TrueNAS cluster.

We use the Node.js version of FoundryVTT. The Foundry version can be updated by rebuilding the Dockerfile with the following command:
`make update-foundry-version URL=<timed_url>`

### Set Up
1. Set up the TrueNAS VM.
2. Use the supplied VNC shell to enable SSH.
3. Install K3s on the VM and grab the config file for your local machine.
   * Swap the IP to match the VM.
4. Update your host file for the `noodles.local` VM IP.
5. Sops decrypt the secrets via `make sops-decrypt-all`.
6. Run the ansible setup script via `make setup-cluster`
   * Remember to check the location of the config file.


### Kubelogin
1. Install [kubelogin](https://github.com/int128/kubelogin): `brew install kubelogin`
2. Add the following user and context to your `~/.kube/config`:
   ```yaml
   users:
     - name: oidc
       user:
         exec:
           apiVersion: client.authentication.k8s.io/v1beta1
           command: kubectl
           args:
             - oidc-login
             - get-token
             - --oidc-issuer-url=https://dex.noodles.quest
             - --oidc-client-id=kubelogin
             - --oidc-extra-scope=email
             - --oidc-extra-scope=groups
             - --oidc-client-secret=<KUBELOGIN_CLIENT_SECRET>
             - --certificate-authority-data=LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUZGVENDQXYyZ0F3SUJBZ0lVYjZqZjBwYXJ5MDdGbytJamczOXJNYnJOcmlVd0RRWUpLb1pJaHZjTkFRRUwKQlFBd0VqRVFNQTRHQTFVRUF3d0hhM1ZpWlMxallUQWVGdzB5TmpBMU1UZ3lNalEwTlRKYUZ3MHpOakExTVRVeQpNalEwTlRKYU1CSXhFREFPQmdOVkJBTU1CMnQxWW1VdFkyRXdnZ0lpTUEwR0NTcUdTSWIzRFFFQkFRVUFBNElDCkR3QXdnZ0lLQW9JQ0FRQ3BCTHhVNTBFZm5aK1hleUp2M1JXTE1FUmNqcGxxRHRTeEdoR0tGNytSeWFIZC83ZC8KeHI3K2ZRVlBBNlI1aTFZb294Wi94b2dBUFhvTzRLMEp0WjA4cDR2Q2FkaHpqRDI4ZHhuYnZ5RlRGVmUyb0wxOApRY3FwQi91L0EyTG50S2laTVA5dzgyUzhMNmd6K0pKNXJUMjdBU05EdGRaQlhWTDlDWW0xb21HUkF0Sk9lTE51CndSVTM3K2Q1c3FXb0dETExZNlFsOUJsQlE4Zk1maHVDK25lWVZYeTFYb0VmNU40QjNHVjBCaDdsWk5oRlRkRmoKZEJzcUx5bnFRY1hWZVYzN1o5b3NCdzRyWXVobWhtaGJUWkdMSGlzZ1cyQk42U0hyMTZzODNxdU40U2gwa2xzRQpRR0NNSFc5aGFtOHR2a091eHFwM3dYTEtwVVFyQ3dPNWZoMC9Obi8vZ1ludHpHMXVFd0wyL0EwZmU4NkhieWlpClhmd25HTzYvLzRiSDcyR29GTHhERlduMlAwdXdOTGVJd2ZzQ2JSd3UzcVJ0S1ltSXdta0s1S0ozbTRwV1N3dGwKYW5FbTVwTkVlTlZ0ZE5vNk80ME9ibzJ5enBaa3FGcHBMSVMzVENaS09ZaXhWTWQxTk9KQlJuVkFFc0tJZEtxRgpGc3NPT2VGQ2hFOEpxNVlHMkY5TjMyeU1jeG9wSmQ3T2lKTnhwU0pzd0ZNSStTQVExaXAreDRidzRHNm9hSnpNCmNFaEljeSs3YzZWQy9HYWk4cWlGWXNxd0dGR0ZITFNkOXJNWklnRHB0VnhaaW12TnlXWUhrYVpVczYxOFM0YTIKVkxnUVo3dDlTd21lYXRCRlE3cVJlZ1o2Q0Y5RCtlcmF4elRFYVJVckhodS9KUktNeGp4RTdwZnNVd0lEQVFBQgpvMk13WVRBZEJnTlZIUTRFRmdRVVhkalJJQVpzWDVmTW0yYjQzcGFXUERxc1pFWXdId1lEVlIwakJCZ3dGb0FVClhkalJJQVpzWDVmTW0yYjQzcGFXUERxc1pFWXdEd1lEVlIwVEFRSC9CQVV3QXdFQi96QU9CZ05WSFE4QkFmOEUKQkFNQ0FRWXdEUVlKS29aSWh2Y05BUUVMQlFBRGdnSUJBRzhEQWl4aFZ5RUJudlhJZXJMWVVTZ3FOSTdIUlFuTQoxS0NJQ1F3dCtocmswR3piTXJQUFVDcUxQYnptdUFTSm9WN3FJckNja0lEbzZ1ZElZUHYwYjhDYURPaHZaTlBYCmpBcFpXK2RveWI4SUtTQWM2TDR0SitpMFlvVWNvN0oxOXVpVktLd1pzcDhvREhNcjNHenJ0WS9ZZXRRUS8yamYKcWVDNEUzNjRUejMxTUxzM2c2N25CTEQ4SEt6cmlsb01iTjJwdXR1VnFiSERVaU8wbDZjOStsbGt3UXRNSm50Ywo5TVFiRCt3UWkxMk9FQkVMZkYxRnFqTUVrQXFqODVKSWZqMmNuNmNyNllvQUQ4eDBZenNZNk5YUU5teVR6WlFKCnR2dXJvekRFcGZJUTFCSmh0SjA2aW1oVUhCVzZlR3VvelJNWHJDYmllK2VGMlltT2grYk14SnVoZUNaS0NmaWUKcVFGSWpnSWhtUm9TLzlrU3BrL0tTK0l0ZkxFYTl2SUtpV2h1dENHN1dUQUNFVE5ZWWoxbGtkTU8xZkEvZ1Z6bwpBd3V4Yk5Bdy8xeC9pNHVRMXJtR1R1Nkxha3p2eVlrMXFLa0hNMGk0dmtJN3F0RVc1RDFtMTFBOUM2TWR3NXJEClpnRUpqSEhBay9MQm5KOUlIcTFFZ0V2Ny9Ua3VSYUVKdlpBWlNwYzdoODlDRk5mc2pya0daby9VKzFqd2lrWnUKcldWUktMbUdkV0tqSklMTlN2eDVaNjFoeERLVWtpZ3B0S1Z3cUkxUzF0SHRtQWtvSHdVd0lXRUw3eHp0SzE0RwpDMWtNa085MzJkWDloVjVicnE4N3ZKKzhmZlpOUTU5V1VRbU1lamJLajE3cnZSRGsvRkphWDRyVjIrSHNESWFFCmFiSmpKb0l4RjdGWAotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
   contexts:
     - name: noodles-oidc
       context:
         cluster: foundry-cluster
         user: oidc
   ```
3. Switch to the context: `kubectl config use-context noodles-oidc`


### Directly Access the Foundry Files
1. `k get pod -n foundry` to find the pod
2. `k exec {POD NAME} -it -n foundry -- sh` to exec into it with a shell terminal
3. `cd /foundrydata` to directly access the files.
