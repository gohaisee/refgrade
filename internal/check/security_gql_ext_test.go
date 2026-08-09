package check

import (
	"context"
	"testing"
)

func TestSecG07_flagsAdminResolverWithoutRole(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/graph/resolver/admin.resolvers.go", `package resolver

type mutationResolver struct{}

func (r *mutationResolver) AdminDeleteUser(ctx context.Context, id string) (bool, error) {
	return true, nil
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/graph/resolver/admin.resolvers.go"}}}
	findings, err := NewSecG07().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "sec-g07" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestSecG08_flagsSubscriptionWithoutAuth(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/graph/resolver/subscription.resolvers.go", `package resolver

type subscriptionResolver struct{}

func (r *subscriptionResolver) SubscribeToOrders(ctx context.Context) (<-chan string, error) {
	ch := make(chan string)
	return ch, nil
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/graph/resolver/subscription.resolvers.go"}}}
	findings, err := NewSecG08().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "sec-g08" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestSecG09_flagsShadowGraphQLHandle(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/api/graphql.go", `package api

import "net/http"

func Mount() {
	http.Handle("/graphql", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/api/graphql.go"}}}
	findings, err := NewSecG09().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "sec-g09" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestSecG10_flagsWebhookResolverWithoutHMAC(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/graph/resolver/webhook.resolvers.go", `package resolver

type mutationResolver struct{}

func (r *mutationResolver) StripeWebhook(ctx context.Context, payload string) (bool, error) {
	return true, nil
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/graph/resolver/webhook.resolvers.go"}}}
	findings, err := NewSecG10().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "sec-g10" {
		t.Fatalf("findings = %+v", findings)
	}
}
