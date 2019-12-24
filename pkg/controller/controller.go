package controller

import (
	"auth0-ingress-controller/pkg/auth0utils"
	"auth0-ingress-controller/pkg/callbacks"
	"auth0-ingress-controller/pkg/config"
	"auth0-ingress-controller/pkg/constants"
	"auth0-ingress-controller/pkg/kube"
	"auth0-ingress-controller/pkg/kube/wrappers"
	"auth0-ingress-controller/pkg/util"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/auth0.v1/management"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"time"
)

// MonitorController which can be used for monitoring ingresses
type Auth0Controller struct {
	kubeClient kubernetes.Interface
	namespace  string
	indexer    cache.Indexer
	queue      workqueue.RateLimitingInterface
	informer   cache.Controller
	management *management.Management
	config     config.Config
}

// NewMonitorController implements a controller to monitor ingresses and routes
func NewAuth0Controller(
	namespace string,
	kubeClient kubernetes.Interface,
	config config.Config,
	resource string,
	restClient rest.Interface) *Auth0Controller {

	controller := &Auth0Controller{
		kubeClient: kubeClient,
		namespace:  namespace,
		config:     config,
	}

	m, err := management.New(config.Client.Domain, config.Client.ClientID, config.Client.ClientSecret)
	if err != nil {
		log.Panic("Unable to create auth0 manager")
	}

	controller.management = m

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// Create the Ingress Watcher
	ingressListWatcher := cache.NewListWatchFromClient(restClient, resource, namespace, fields.Everything())

	indexer, informer := cache.NewIndexerInformer(ingressListWatcher, kube.ResourceMap[resource], time.Duration(config.ResyncPeriod)*time.Second, cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.onResourceAdded,
		UpdateFunc: controller.onResourceUpdated,
		DeleteFunc: controller.onResourceDeleted,
	}, cache.Indexers{})

	controller.indexer = indexer
	controller.informer = informer
	controller.queue = queue

	return controller
}

// Run method starts the controller
func (c *Auth0Controller) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	log.Println("Starting Ingress Monitor controller")

	go c.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	log.Println("Stopping Ingress Monitor controller")
}

func (c *Auth0Controller) runWorker() {
	for c.processNextItem() {
	}
}

func (c *Auth0Controller) processNextItem() bool {
	// Wait until there is a new item in the working queue
	action, quit := c.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two ingresses with the same key are never processed in
	// parallel.
	defer c.queue.Done(action)

	// Invoke the method containing the business logic
	err := action.(Action).handle(c)
	// Handle the error if something went wrong during the execution of the business logic
	c.handleErr(err, action)
	return true
}

func (c *Auth0Controller) getIngressName(rAFuncs callbacks.ResourceActionFuncs, resource interface{}) string {
	format, err := util.GetNameTemplateFormat("{{.Namespace}}/{{.IngressName}}")
	if err != nil {
		log.Fatal("Failed to parse MonitorNameTemplate")
	}
	return fmt.Sprintf(format, rAFuncs.NameFunc(resource), rAFuncs.NamespaceFunc(resource))
}

func (c *Auth0Controller) addCallbackURLs(resource interface{}) error {
	ingressWrapper := wrappers.IngressWrapper{
		Ingress:    resource.(*v1beta1.Ingress),
		Namespace:  resource.(*v1beta1.Ingress).Namespace,
		KubeClient: c.kubeClient,
	}
	rAFuncs := kube.GetResourceActionFuncs(resource)
	annotations := rAFuncs.AnnotationFunc(resource)
	appID, ok := annotations[constants.Auth0AppID]
	if !ok {
		log.Infof("AppID is not set as annotation %s", constants.Auth0AppID)
		return nil
	}
	currentClient, err := c.management.Client.Read(appID)

	if err != nil {
		return err
	}
	client := &management.Client{
		Callbacks:         currentClient.Callbacks,
		WebOrigins:        currentClient.WebOrigins,
		AllowedLogoutURLs: currentClient.AllowedLogoutURLs,
	}
	var count int
	count += auth0utils.AddItem(&client.Callbacks, ingressWrapper.GetCallbackURLsFromIngress()...)
	count += auth0utils.AddItem(&client.WebOrigins, util.NormalizeURL(ingressWrapper.GetCallbackURLsFromIngress())...)
	count += auth0utils.AddItem(&client.AllowedLogoutURLs, util.NormalizeURL(ingressWrapper.GetCallbackURLsFromIngress())...)

	if count > 0 {
		err := c.management.Client.Update(appID, client)
		if err != nil {
			return err
		} else {
			log.Infof("client %s successfully updated. Added urls: %v", appID, ingressWrapper.GetCallbackURLsFromIngress())
		}
	}
	return nil
}

func (c *Auth0Controller) deleteCallbackURLs(resource interface{}) error {
	ingressWrapper := wrappers.IngressWrapper{
		Ingress:    resource.(*v1beta1.Ingress),
		Namespace:  resource.(*v1beta1.Ingress).Namespace,
		KubeClient: c.kubeClient,
	}
	rAFuncs := kube.GetResourceActionFuncs(resource)
	annotations := rAFuncs.AnnotationFunc(resource)
	appID, ok := annotations[constants.Auth0AppID]
	if !ok {
		log.Infof("AppID is not set as annotation %s", constants.Auth0AppID)
		return nil
	}
	currentClient, err := c.management.Client.Read(appID)

	if err != nil {
		return err
	}
	client := &management.Client{
		Callbacks:         currentClient.Callbacks,
		WebOrigins:        currentClient.WebOrigins,
		AllowedLogoutURLs: currentClient.AllowedLogoutURLs,
	}
	var count int
	count += auth0utils.DeleteItem(&client.Callbacks, ingressWrapper.GetCallbackURLsFromIngress()...)
	count += auth0utils.DeleteItem(&client.WebOrigins, util.NormalizeURL(ingressWrapper.GetCallbackURLsFromIngress())...)
	count += auth0utils.DeleteItem(&client.AllowedLogoutURLs, util.NormalizeURL(ingressWrapper.GetCallbackURLsFromIngress())...)

	// Due to bug in auth0 client we set default to http://localhost if list is empty
	if len(client.WebOrigins) == 0 {
		client.WebOrigins = append(client.WebOrigins, "http://localhost:3000", "http://localhost:5000")
	}
	if len(client.Callbacks) == 0 {
		client.Callbacks = append(client.Callbacks, "http://localhost/callback")
	}
	if count > 0 {
		err := c.management.Client.Update(appID, client)
		if err != nil {
			return err
		} else {
			log.Infof("client %s successfully updated. Deleted urls: %v", appID, ingressWrapper.GetCallbackURLsFromIngress())
		}
	}
	return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *Auth0Controller) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.queue.NumRequeues(key) < 5 {
		log.Printf("Error syncing ingress %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	log.Printf("Dropping ingress %q out of the queue: %v", key, err)
}

func (c *Auth0Controller) onResourceAdded(obj interface{}) {
	c.queue.AddAfter(ResourceUpdatedAction{
		resource: obj,
	}, c.config.CreationDelay)
}

func (c *Auth0Controller) onResourceUpdated(old interface{}, new interface{}) {
	c.queue.AddAfter(ResourceUpdatedAction{
		resource:    new,
		oldResource: old,
	}, c.config.CreationDelay)
}

func (c *Auth0Controller) onResourceDeleted(obj interface{}) {
	c.queue.Add(ResourceDeletedAction{
		resource: obj,
	})
}
