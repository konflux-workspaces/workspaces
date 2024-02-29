package user

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
	"k8s.io/apimachinery/pkg/runtime"
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
	workspacesiov1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

func whenAnUserOnboards(ctx context.Context) (context.Context, error) {
	u, err := OnBoardUserInKubespaceNamespace(ctx, DefaultUserName)
	if err != nil {
		return ctx, err
	}
	return tcontext.InjectUser(ctx, *u), nil
}

func whenUserRequestsTheListOfWorkspaces(ctx context.Context) (context.Context, error) {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(workspacesiov1alpha1.AddToScheme(scheme))
	utilruntime.Must(toolchainv1alpha1.AddToScheme(scheme))

	u := tcontext.RetrieveUser(ctx)
	k := tcontext.RetrieveUnauthKubeconfig(ctx)
	t := auth.BuildJwtForUser(u.Status.CompliantUsername)
	ts, err := t.SignedString([]byte("randomkey"))
	if err != nil {
		return ctx, err
	}
	k.BearerToken = ts
	k.Host = os.Getenv("PROXY_URL")

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
		return ctx, err
	}
	m, err := apiutil.NewDynamicRESTMapper(cfg, hc)
	if err != nil {
		return ctx, err
	}

	c, err := client.New(k, client.Options{Scheme: scheme, Mapper: m})
	if err != nil {
		return ctx, fmt.Errorf("error building client for host %s and user %s: %w", k.Host, u.Status.CompliantUsername, err)
	}

	ww := workspacesiov1alpha1.WorkspaceList{}
	if err := c.List(ctx, &ww, &client.ListOptions{}); err != nil {
		return ctx, fmt.Errorf("error retrieving workspaces from host %s as user %s: %w", k.Host, u.Status.CompliantUsername, err)
	}

	return tcontext.InjectUserWorkspaces(ctx, ww), nil
}

func whenUserRequestsTheirDefaultWorkspace(ctx context.Context) (context.Context, error) {
	return ctx, godog.ErrPending
}
