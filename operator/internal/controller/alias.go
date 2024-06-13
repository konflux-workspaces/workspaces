package controller

import (
	"github.com/konflux-workspaces/workspaces/operator/internal/controller/internalworkspace"
	"github.com/konflux-workspaces/workspaces/operator/internal/controller/usersignup"
)

type (
	UserSignupReconciler = usersignup.UserSignupReconciler
	WorkspaceReconciler  = internalworkspace.WorkspaceReconciler
)
