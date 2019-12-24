package controller

import (
	"auth0-ingress-controller/pkg/constants"
	"auth0-ingress-controller/pkg/kube"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/extensions/v1beta1"
)

// Action is an interface for ingress and route actions
type Action interface {
	handle(c *Auth0Controller) error
	GetIngressName(c *Auth0Controller) string
}

// ResourceUpdatedAction provide implementation of action interface
type ResourceUpdatedAction struct {
	resource    interface{}
	oldResource interface{}
}

// ResourceDeletedAction provide implementation of action interface
type ResourceDeletedAction struct {
	resource interface{}
}

func (r ResourceUpdatedAction) GetIngressName(c *Auth0Controller) string {
	rAFuncs := kube.GetResourceActionFuncs(r.resource)
	return c.getIngressName(rAFuncs, r.resource)
}

func getAnnotations(r interface{}) map[string]string {
	return r.(*v1beta1.Ingress).Annotations
}

func (r ResourceDeletedAction) GetIngressName(c *Auth0Controller) string {
	rAFuncs := kube.GetResourceActionFuncs(r.resource)
	return c.getIngressName(rAFuncs, r.resource)
}

func (r ResourceUpdatedAction) handle(c *Auth0Controller) error {
	annotations := getAnnotations(r.resource)
	ingressName := r.GetIngressName(c)

	// Delete all urls from old resource when we edited ingress
	if r.oldResource != nil {
		err := c.deleteCallbackURLs(r.oldResource)
		if err != nil {
			return err
		}
	}

	if value, ok := annotations[constants.Auth0EnabledAnnotation]; ok {
		if value == "true" {
			// Annotation exists and is enabled
			log.Infof("Annotation %s added, ingress: %s", constants.Auth0EnabledAnnotation, ingressName)
			return c.addCallbackURLs(r.resource)
		} else {
			// Annotation exists but is disabled
			log.Infof("Annotation %s has been removed, deleted corresponding URLS, ingress: %s", constants.Auth0EnabledAnnotation, ingressName)
			return c.deleteCallbackURLs(r.resource)
		}
	} else {
		log.Debugf("Not doing anything with this ingress because no annotation %s exists on ingress %s", constants.Auth0EnabledAnnotation, ingressName)
	}
	return nil
}

func (r ResourceDeletedAction) handle(c *Auth0Controller) error {
	annotations := getAnnotations(r.resource)
	if _, ok := annotations[constants.Auth0EnabledAnnotation]; !ok {
		return nil
	}
	if c.config.EnableCallbackDeletion {
		log.Infof("Removing auth0 URLs due to deleted ingress %s", r.GetIngressName(c))
		c.deleteCallbackURLs(r.resource)
	} else {
		log.Info("Auth0 callback URLs deletion is not enabled in config. Skipping deletion.")
	}
	return nil
}
