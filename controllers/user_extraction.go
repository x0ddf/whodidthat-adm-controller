package controllers

import (
	admissionv1 "k8s.io/api/admission/v1"
)

// ExtractUserMeta extracts username from different authentication providers
func ExtractUserMeta(request *admissionv1.AdmissionRequest) string {
	if request == nil {
		return "system"
	}
	fields := []string{
		"username",    // rancher | IDP username
		"sessionName", // AWS EKS -> username of the current session
		"arn",         // AWS [AKS,GKE]? -> full ARN of the user
	}
	username := ""
	for _, fieldName := range fields {
		if v, ok := request.UserInfo.Extra[fieldName]; ok && len(v) > 0 && len(v[0]) > 0 {
			username = v[0]
		}
	}
	if len(username) != 0 {
		return username
	}
	userInfo := request.UserInfo

	// Start with username
	username = userInfo.Username

	// Handle different authentication cases
	switch {
	// Service accounts (format: system:serviceaccount:namespace:name)
	case username == "system:admin":
		return "system:admin"
	case username == "system:apiserver":
		return "system:apiserver"
	case username == "system:kube-scheduler":
		return "system:kube-scheduler"
	case username == "system:kube-controller-manager":
		return "system:kube-controller-manager"
	// Handle service accounts
	case len(userInfo.Groups) > 0 && contains(userInfo.Groups, "system:serviceaccounts"):
		return username // Return full service account name
	// OAuth/OIDC usually have email as username
	case username != "":
		return username
	// Default fallback
	default:
		return "unknown"
	}
}

// contains checks if a string slice contains a specific value
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
