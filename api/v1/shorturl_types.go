// api/v1/shorturl_types.go
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ShortURLSpec defines the desired state of ShortURL
type ShortURLSpec struct {
	// TargetURL is the original URL to be shortened
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format=url
	TargetURL string `json:"targetURL"`
}

// ShortURLStatus defines the observed state of ShortURL
type ShortURLStatus struct {
	// ShortPath is the generated short path
	// +kubebuilder:validation:Pattern=^/[a-zA-Z0-9]+$
	ShortPath string `json:"shortPath,omitempty"`

	// ClickCount is the number of times the short URL has been accessed
	// +kubebuilder:validation:Minimum=0
	ClickCount int64 `json:"clickCount,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Target URL",type=string,JSONPath=`.spec.targetURL`
//+kubebuilder:printcolumn:name="Short Path",type=string,JSONPath=`.status.shortPath`
//+kubebuilder:printcolumn:name="Clicks",type=integer,JSONPath=`.status.clickCount`

// ShortURL is the Schema for the shorturls API
type ShortURL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ShortURLSpec   `json:"spec,omitempty"`
	Status ShortURLStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// ShortURLList contains a list of ShortURL
type ShortURLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ShortURL `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ShortURL{}, &ShortURLList{})
}
