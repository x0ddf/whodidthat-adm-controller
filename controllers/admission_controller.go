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

var FallBackResponse = &admissionv1.AdmissionResponse{Allowed: true}

type AdmissionController struct {
	decoder runtime.Decoder
}

func NewAdmissionController() *AdmissionController {
	return &AdmissionController{
		decoder: serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer(),
	}
}

// sanitizeLabelValue converts a username into a valid Kubernetes label value
func sanitizeLabelValue(username string) string {
	// Replace @ with -at- and remove any other invalid characters
	sanitized := strings.ReplaceAll(username, "@", "_at_")

	// Replace any other invalid characters with dashes
	// Valid characters are alphanumeric, '-', '_', and '.'
	var result strings.Builder
	for _, r := range sanitized {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_' || r == '.' {
			result.WriteRune(r)
		} else {
			result.WriteRune('-')
		}
	}

	// Ensure the value starts and ends with an alphanumeric character
	value := result.String()
	value = strings.Trim(value, "-_.")

	// If empty after sanitization, return a default value
	if value == "" {
		return "unknown-user"
	}

	// Kubernetes label values must be 63 characters or less
	if len(value) > 63 {
		return value[:63]
	}

	return value
}

func addLabelPatch(label string, value interface{}) map[string]interface{} {
	return map[string]interface{}{
		"op":    "add",
		"path":  "/metadata/labels/" + escapeJSONPointer(label),
		"value": value,
	}
}
func addAnnotationPatch(annotation string, value interface{}) map[string]interface{} {
	return map[string]interface{}{
		"op":    "add",
		"path":  "/metadata/annotations/" + escapeJSONPointer(annotation),
		"value": value,
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

	username := ExtractUserMeta(admissionReview.Request)
	// Get the requesting user and sanitize it
	sanitized := sanitizeLabelValue(username)

	// Create patch operations
	var patches []map[string]interface{}

	// Check if we need to add the labels field first
	var obj map[string]interface{}
	if err := json.Unmarshal(admissionReview.Request.Object.Raw, &obj); err == nil {
		metadata, ok := obj["metadata"].(map[string]interface{})
		if ok && metadata["labels"] == nil {
			patches = append(patches, map[string]interface{}{
				"op":    "add",
				"path":  "/metadata/labels",
				"value": map[string]string{},
			})
		}
		if ok && metadata["annotations"] == nil {
			patches = append(patches, map[string]interface{}{
				"op":    "add",
				"path":  "/metadata/annotations",
				"value": map[string]string{},
			})
		}
	}

	// Handle different operations
	switch admissionReview.Request.Operation {
	case admissionv1.Create:
		patches = append(patches, addAnnotationPatch(CreatedByLabel, username))
	case admissionv1.Update:
		patches = append(patches, addLabelPatch(ModifiedByLabel, sanitized))
	}
	// Create the admission response
	admissionResponse := &admissionv1.AdmissionResponse{
		UID:     admissionReview.Request.UID,
		Allowed: true,
	}
	// Create the patch bytes

	if patchBytes, err := json.Marshal(patches); err == nil {
		admissionResponse.Patch = patchBytes
		admissionResponse.PatchType = func() *admissionv1.PatchType {
			pt := admissionv1.PatchTypeJSONPatch
			return &pt
		}()
	} else {
		log.Printf("fail to process request with id:%v", requestId)
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

	log.Printf("Processing request: operation=%s, user=%s, resource=%s/%s",
		admissionReview.Request.Operation,
		sanitized,
		admissionReview.Request.Resource.Resource,
		admissionReview.Request.Name)

	log.Printf("Successfully processed request: applied %d patches", len(patches))
}

// escapeJSONPointer escapes / in label names for JSON Pointer compliance
func escapeJSONPointer(s string) string {
	s = strings.ReplaceAll(s, "~", "~0")
	s = strings.ReplaceAll(s, "/", "~1")
	return s
}
