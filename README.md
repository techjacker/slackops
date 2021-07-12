[![build](https://github.com/techjacker/slackops/actions/workflows/build.yml/badge.svg)](https://github.com/techjacker/slackops/actions/workflows/build.yml)
# Slackops
K8 operator that ensures slack mirrors the state of the world in a given kubernetes cluster. Who needs gitops when you can have slackops?

### Bonus
- [x] Container security scan (see [build](.github/workflows/build.yml))
- [x] Does not run as root (see [deployment](config/manager/manager.yaml))
--------------------------------
## Quickstart
### Deploy

### Deploy Operator
```shell
kubectl apply -f templates-exports
```
#### Create K8 ConfigMap
Update the [configmap](fixtures/configmap.yaml) with desired values.
```shell
kubectl apply -f fixtures/configmap.yaml
```
#### Deploy K8 slack bot secret
Follow the [slack setup guide below](#slack-setup) and set `$SLACK_BOT_TOKEN` in your environment. Then run:
```shell
export K8_OP_NS=slackops-system
kubectl -n "$K8_OP_NS" create secret generic slackops \
  --from-literal=slack_bot_token="$SLACK_BOT_TOKEN"
```
### Check Operator Working
```shell
export K8_OP_NS=slackops-system
kubectl config set-context --current --namespace=${K8_OP_NS}
kubectl logs -c manager slackops-controller-manager-xxxx
```
### Test
Ensure the pod name you create contains the string specified in the configmap you deloyed (eg `pod_target_contains: "foo"`).
```shell
# triggers create
kubectl run nginx-foo --image=nginx --restart=Never
# triggers update (only works for label updates currently)
kubectl patch pod nginx-foo --patch "$(cat fixtures/patch.yaml)"
# tiggers delete
kubectl delete pod nginx-foo
# tiggers nothing because does not have "foo" in the name
kubectl run nginx-blah --image=nginx --restart=Never
```
### Remove Operator
```shell
kubectl delete -f templates-exports
```
---------------------------------------------------------
## Slack Setup
### Slack App and Token Creation
1. [Create a slack app](https://api.slack.com/start/overview#creating) in your slack workspace
2. [Request the needed permissions and generate a token for the app](https://api.slack.com/messaging/sending#permissions) which are
   1. `chat:write.public`
   2. `chat:write`
3. Generate an OAuth token for the app via the OAuth & Permissions settings page (you'll need to add this as a k8 secret in the operator namespace)
4. (Re-)Install the app after you've added the permissions so they are applied to your workspace

### Testing token works
Check that you have granted all required permissions by making a test call via `curl`.
Example assumes:
- you have a public channel named `random` in your slack workspace
- you have set the OAuth token as the `SLACK_BOT_TOKEN`  environment variable in your shell
```shell
$ curl -XPOST -H "Authorization: Bearer $SLACK_BOT_TOKEN" -H "Content-type: application/json" -d '{
  "channel": "random",
  "text": "Hello world :tada:"
}' 'https://slack.com/api/chat.postMessage'
```

---------------------------------------------------------
## Local Development
### Install dependencies
```
make deps-lint
pip install pre-commit
pre-commit install
pre-commit install-hooks
```
### Setup of new API
```shell
$ operator-sdk init --domain=techjacker.com --owner "Andrew Griffiths" --repo github.com/techjacker/slackops
$ operator-sdk create api --group=core --version=v1 --kind=Pod --controller=true --resource=false
```
### Running app locally
Create a .`env` file and populate with the env vars listed in `.env.example`.
```shell
source .env.example
source .env
make install
# ensure you have a valid kubeconfig set before running
make run
# need to manually created the slackops-system namespace before can test by creating/deleting pods
```
### Build Docker Image
```
source .env.example
source .env
echo $DOCKERHUB_PWD | docker login --username $DOCKERHUB_ID --password-stdin
make docker-build
make docker-push
```

### Deploy to k8
```shell
#### Create secrets
export K8_OP_NS=slackops-system
kubectl -n "$K8_OP_NS" create secret generic slackops \
  --from-literal=slack_bot_token="$SLACK_BOT_TOKEN"
#### Create ConfigMap
kubectl apply -f fixtures/configmap.yaml
### Deploy Operator option #1
make deploy
### Deploy Operator option #2
make deploy-export
kubectl apply -f templates-exports
#### check it worked
export K8_OP_NS=slackops-system
kubectl config set-context --current --namespace=${K8_OP_NS}
kubectl logs -c manager slackops-controller-manager-56895df697-9tz67
### Delete Operator
make undeploy
```
