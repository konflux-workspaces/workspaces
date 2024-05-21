package user

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/konflux-workspaces/workspaces/e2e/pkg/auth"
	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

func whenAnUserOnboards(ctx context.Context) (context.Context, error) {
	u, err := OnBoardUserInKubespaceNamespace(ctx, DefaultUserName)
	if err != nil {
		return ctx, err
	}
	return tcontext.InjectUser(ctx, *u), nil
}

func whenUserRequestsTheListOfWorkspaces(ctx context.Context) (context.Context, error) {
	c, err := buildWorkspacesClient(ctx)
	if err != nil {
		return ctx, err
	}
	ww := workspacesv1alpha1.InternalWorkspaceList{}
	if err := c.List(ctx, &ww, &client.ListOptions{}); err != nil {
		u := tcontext.RetrieveUser(ctx)
		k := tcontext.RetrieveUnauthKubeconfig(ctx)
		return ctx, fmt.Errorf("error retrieving workspaces from host %s as user %s: %w", k.Host, u.Status.CompliantUsername, err)
	}

	return tcontext.InjectUserWorkspaces(ctx, ww), nil
}

func whenUserRequestsTheirDefaultWorkspace(ctx context.Context) (context.Context, error) {
	c, err := buildWorkspacesClient(ctx)
	if err != nil {
		return ctx, err
	}

	u := tcontext.RetrieveUser(ctx)
	w := workspacesv1alpha1.InternalWorkspace{}
	wk := types.NamespacedName{Namespace: u.Name, Name: u.Name}
	if err := c.Get(ctx, wk, &w, &client.GetOptions{}); err != nil {
		k := tcontext.RetrieveUnauthKubeconfig(ctx)
		return ctx, fmt.Errorf("error retrieving workspace %v from host %s as user %s: %w", wk, k.Host, u.Status.CompliantUsername, err)
	}
	log.Printf("retrieved workspace: %v", w)
	return tcontext.InjectInternalWorkspace(ctx, w), nil
}

func buildWorkspacesClient(ctx context.Context) (client.Client, error) {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(workspacesv1alpha1.AddToScheme(scheme))
	utilruntime.Must(toolchainv1alpha1.AddToScheme(scheme))

	u := tcontext.RetrieveUser(ctx)
	k := tcontext.RetrieveUnauthKubeconfig(ctx)

	t, err := auth.BuildJwtForUser(ctx, u.Status.CompliantUsername)
	if err != nil {
		return nil, err
	}
	log.Printf("token: %s", t)
	k.BearerToken = t
	k.Host = os.Getenv("PROXY_URL")

	m, err := buildRESTMapper()
	if err != nil {
		return nil, err
	}

	c, err := client.New(k, client.Options{Scheme: scheme, Mapper: m})
	if err != nil {
		return nil, fmt.Errorf("error building client for host %s and user %s: %w", k.Host, u.Status.CompliantUsername, err)
	}

	return c, nil
}

func buildRESTMapper() (meta.RESTMapper, error) {
	p := func() string {
		e := os.Getenv("KUBECONFIG")
		if e != "" {
			return e
		}
		return filepath.Join(homedir.HomeDir(), ".kube", "config")
	}()

	cfg, err := clientcmd.BuildConfigFromFlags("", p)
	if err != nil {
		panic(fmt.Sprintf("error building config: %v", err))
	}

	hc, err := rest.HTTPClientFor(cfg)
	if err != nil {
		return nil, err
	}

	m, err := apiutil.NewDynamicRESTMapper(cfg, hc)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func whenTheUserChangesWorkspaceVisibilityTo(ctx context.Context, visibility string) (context.Context, error) {
	w := tcontext.RetrieveInternalWorkspace(ctx)

	cli, err := buildWorkspacesClient(ctx)
	if err != nil {
		return ctx, err
	}

	w.Spec.Visibility = workspacesv1alpha1.InternalWorkspaceVisibility(visibility)
	if err := cli.Update(ctx, &w, &client.UpdateOptions{}); err != nil {
		return ctx, err
	}
	return tcontext.InjectInternalWorkspace(ctx, w), nil
}
