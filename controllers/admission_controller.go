package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

const (
	CreatedByLabel  = "audit.k8s.io/created-by"
	ModifiedByLabel = "audit.k8s.io/last-modified-by"
)

type AdmissionController struct {
	decoder runtime.Decoder
}

func NewAdmissionController() *AdmissionController {
	return &AdmissionController{
		decoder: serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer(),
	}
}

func (ac *AdmissionController) Handle(w http.ResponseWriter, r *http.Request) {
	// Parse the AdmissionReview
	admissionReview := &admissionv1.AdmissionReview{}
	if err := json.NewDecoder(r.Body).Decode(admissionReview); err != nil {
		http.Error(w, fmt.Sprintf("could not parse admission review: %v", err), http.StatusBadRequest)
		return
	}
	requestId := admissionReview.Request.UID
	log.Printf("new admission request:%v", requestId)

	// Get the requesting user
	userName := ExtractUserMeta(admissionReview.Request)

	// Create patch operations
	var patches []map[string]interface{}

	// Handle different operations
	switch admissionReview.Request.Operation {
	case admissionv1.Create:
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/labels/" + escapeJSONPointer(CreatedByLabel),
			"value": userName,
		})
	case admissionv1.Update:
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/labels/" + escapeJSONPointer(ModifiedByLabel),
			"value": userName,
		})
	}

	// Create the patch bytes
	patchBytes, err := json.Marshal(patches)
	if err != nil {
		log.Printf("fail to process request with id:%v", requestId)
		http.Error(w, fmt.Sprintf("could not marshal patch: %v", err), http.StatusInternalServerError)
		return
	}

	// Create the admission response
	admissionResponse := &admissionv1.AdmissionResponse{
		UID:     admissionReview.Request.UID,
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *admissionv1.PatchType {
			pt := admissionv1.PatchTypeJSONPatch
			return &pt
		}(),
	}

	// Return the admission review
	admissionReview.Response = admissionResponse
	resp, err := json.Marshal(admissionReview)
	if err != nil {
		log.Printf("fail to process request with id:%v", requestId)
		http.Error(w, fmt.Sprintf("could not marshal response: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(resp); err != nil {
		log.Printf("fail to process request with id:%v", requestId)
		http.Error(w, fmt.Sprintf("fail to write response:%v", err), http.StatusInternalServerError)
	}
}

// escapeJSONPointer escapes / in label names for JSON Pointer compliance
func escapeJSONPointer(s string) string {
	return strings.ReplaceAll(s, "/", "~1")
}

func ExtractUserMeta(request *admissionv1.AdmissionRequest) string {
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

	if len(username) == 0 {
		username = request.UserInfo.Username
	}
	return username
}
