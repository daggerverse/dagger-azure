// Push a container image into Azure Container Registry (acr)
//
// This module lets you push a container into acr, automating the tedious manual steps of configuring your local docker daemon with the credentials
//
// For more info and sample usage, check the readme: https://github.com/lukemarsden/dagger-azure

package main

import (
	"context"
	"encoding/json"
	"fmt"
)

type Azure struct{}

// example usage: "dagger call get-secret --azure-credentials ~/.azure/"
func (m *Azure) GetSecret(ctx context.Context, azureCredentials *Directory) (string, error) {
	ctr, err := m.WithAzureSecret(ctx, dag.Container().From("ubuntu:latest"), azureCredentials)
	if err != nil {
		return "", err
	}
	return ctr.
		WithExec([]string{"bash", "-c", "cat /root/.azure/msal_token_cache.json |base64"}).
		Stdout(ctx)
}

func (m *Azure) WithAzureSecret(ctx context.Context, ctr *Container, azureCredentials *Directory) (*Container, error) {
	return ctr.WithMountedDirectory("/root/.azure", azureCredentials), nil
}

func (m *Azure) AzureCli(ctx context.Context, azureCredentials *Directory) (*Container, error) {
	ctr := dag.Container().
		From("mcr.microsoft.com/azure-cli:latest")
	ctr, err := m.WithAzureSecret(ctx, ctr, azureCredentials)
	if err != nil {
		return nil, err
	}
	return ctr, nil
}

// example usage: "dagger call acr-get-login-password --acr-name daggertest --azure-credentials ~/.azure/"
func (m *Azure) AcrGetLoginPassword(ctx context.Context, azureCredentials *Directory, acrName string) (string, error) {
	ctr, err := m.AzureCli(ctx, azureCredentials)
	if err != nil {
		return "", err
	}
	return ctr.
		WithExec([]string{"az", "acr", "login", "--name", acrName, "--expose-token"}).
		Stdout(ctx)
}

// Push ubuntu:latest to acr under given repo 'test' (repo must be created first)
// example usage: "dagger call acr-push-example --azure-credentials ~/.azure/credentials --acr-name daggertest --repo test"
func (m *Azure) AcrPushExample(ctx context.Context, azureCredentials *Directory, acrName, repo string) (string, error) {
	ctr := dag.Container().From("ubuntu:latest")
	return m.AcrPush(ctx, azureCredentials, acrName, repo, ctr)
}

func (m *Azure) AcrPush(ctx context.Context, azureCredentials *Directory, acrName, repo string, pushCtr *Container) (string, error) {
	// Get the acr login password so we can authenticate with Publish WithRegistryAuth
	// https://learn.microsoft.com/en-us/azure/container-registry/container-registry-authentication?tabs=azure-cli#az-acr-login-with---expose-token
	ctr, err := m.AzureCli(ctx, azureCredentials)
	if err != nil {
		return "", err
	}
	/// XXX why is this duplicated?
	regCred, err := ctr.
		WithExec([]string{"az", "acr", "login", "--name", acrName, "--expose-token"}).
		Stdout(ctx)
	if err != nil {
		return "", err
	}

	var j struct {
		AccessToken string `json:"accessToken"`
	}

	err = json.Unmarshal([]byte(regCred), &j)
	if err != nil {
		return "", err
	}

	accessToken := j.AccessToken
	secret := dag.SetSecret("azure-reg-cred", accessToken)
	acrHost := fmt.Sprintf("%s.azurecr.io", acrName)
	acrWithRepo := fmt.Sprintf("%s/%s", acrHost, repo)

	return pushCtr.WithRegistryAuth(acrHost, "00000000-0000-0000-0000-000000000000", secret).Publish(ctx, acrWithRepo)
}
