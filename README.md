# Dagger Azure module

Known to work with Dagger v0.11.0.


## Push Image to Private Azure Container Registry (ACR) Repo

This module lets you push container images from Dagger to ACR without having to manually configure docker with ACR credentials. It will fetch them automatically behind the scenes.

From CLI, to push `ubuntu:latest` to a given ACR repo, by way of example:

```
dagger call -m github.com/daggerverse/dagger-azure \
    acr-push-example --acr-name daggertest --repo test --azure-credentials ~/.azure/
```

Check `acr-name` and update it to match the top-level name of your ACR registry. Update `repo` to the name of the repo you want the image to appear under within the ACR registry.

Make sure you've created the registry in your Azure account.

## From Dagger Code

Call the AcrPush method on this module with the azureCredentials *Directory (e.g. ~/.azure/) as the first argument, then the acrName and repo as strings, then finally the container you wish to push as the final argument.

For example:

```go
func (y *YourThing) PushYourThings(ctx context.Context, awsCredentials *File) {
    ctr := dag.Container()
        .From("yourbase:image")
        .YourThings()
    // get acrName, repo
    out, err := m.EcrPush(ctx, acrName, repo, ctr)
}
```

See `AcrPushExample` for a concrete example.