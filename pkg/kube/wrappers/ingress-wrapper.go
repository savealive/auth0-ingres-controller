package wrappers

import (
	"auth0-ingress-controller/pkg/constants"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"net/url"
	"path"
	"regexp"
)

type IngressWrapper struct {
	Ingress    *v1beta1.Ingress
	Namespace  string
	KubeClient kubernetes.Interface
}

func (iw *IngressWrapper) supportsTLS() bool {
	if iw.Ingress.Spec.TLS != nil && len(iw.Ingress.Spec.TLS) > 0 {
		return true
	}
	return false
}

func (iw *IngressWrapper) rulesExist() bool {
	if iw.Ingress.Spec.Rules != nil && len(iw.Ingress.Spec.Rules) > 0 {
		return true
	}
	return false
}

// i is rule index
func (iw *IngressWrapper) getIngressSubPath(i int) string {
	rule := iw.Ingress.Spec.Rules[i]
	if rule.HTTP != nil {
		if rule.HTTP.Paths != nil && len(rule.HTTP.Paths) > 0 {
			reg := regexp.MustCompile(`(\/[^\*|\(|\)|\?]*)\/?.*`)
			return reg.ReplaceAllString(rule.HTTP.Paths[0].Path, "${1}")
		}
	}
	return ""
}

//GetCallbackURLsFromIngress Returns list of ingress rules
func (iw *IngressWrapper) GetCallbackURLsFromIngress() []interface{} {
	var (
		callbacks []interface{}
		scheme    string
	)
	annotations := iw.Ingress.GetAnnotations()

	if !iw.rulesExist() {
		log.Println("No rules exist in ingress: " + iw.Ingress.GetName())
		return callbacks
	}
	if value, ok := annotations[constants.Auth0CallbackScheme]; ok {
		// we set desired scheme via annotations
		scheme = value
	} else {
		if iw.supportsTLS() {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	// Loop over ingress rules
	for i, r := range iw.Ingress.Spec.Rules {
		var u url.URL
		u.Scheme = scheme
		u.Host = r.Host
		if value, ok := annotations[constants.OverridePathAnnotation]; ok {
			u.Path = value
		} else {
			u.Path = iw.getIngressSubPath(i)
		}
		if value, ok := annotations[constants.Auth0CallbackPath]; ok {
			u.Path = path.Join(u.Path, value)
		}
		callbacks = append(callbacks, u.String())
	}
	return callbacks
}

func (iw *IngressWrapper) hasService() (string, bool) {
	ingress := iw.Ingress
	if ingress.Spec.Rules[0].HTTP != nil &&
		ingress.Spec.Rules[0].HTTP.Paths != nil &&
		len(ingress.Spec.Rules[0].HTTP.Paths) > 0 &&
		ingress.Spec.Rules[0].HTTP.Paths[0].Backend.ServiceName != "" {
		return ingress.Spec.Rules[0].HTTP.Paths[0].Backend.ServiceName, true
	}
	return "", false
}

func (iw *IngressWrapper) tryGetHealthEndpointFromIngress() (string, bool) {

	serviceName, exists := iw.hasService()

	if !exists {
		return "", false
	}

	service, err := iw.KubeClient.CoreV1().Services(iw.Ingress.Namespace).Get(serviceName, meta_v1.GetOptions{})
	if err != nil {
		log.Printf("Get service from kubernetes cluster error:%v", err)
		return "", false
	}

	set := labels.Set(service.Spec.Selector)

	if pods, err := iw.KubeClient.CoreV1().Pods(iw.Ingress.Namespace).List(meta_v1.ListOptions{LabelSelector: set.AsSelector().String()}); err != nil {
		log.Printf("List Pods of service[%s] error:%v", service.GetName(), err)
	} else if len(pods.Items) > 0 {
		pod := pods.Items[0]

		podContainers := pod.Spec.Containers

		if len(podContainers) == 1 {
			if podContainers[0].ReadinessProbe != nil && podContainers[0].ReadinessProbe.HTTPGet != nil {
				return podContainers[0].ReadinessProbe.HTTPGet.Path, true
			}
		} else {
			log.Printf("Pod has %d containers so skipping health endpoint", len(podContainers))
		}
	}

	return "", false
}
